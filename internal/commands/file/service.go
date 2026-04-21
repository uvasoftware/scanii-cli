package file

import (
	"context"
	sha1hash "crypto/sha1"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/uvasoftware/scanii-cli/internal/client"
	"github.com/uvasoftware/scanii-cli/internal/commands/profile"
	"golang.org/x/sync/errgroup"
)

type service struct {
	client *client.Client
}

func newService(profile *profile.Profile) (*service, error) {

	c, err := profile.Client()
	if err != nil {
		return nil, err
	}
	return &service{client: c}, nil
}

type consumer func(record resultRecord)

func (s *service) retrieve(ctx context.Context, id string) (*resultRecord, error) {
	resp, err := s.client.RetrieveFile(ctx, id)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}

	r := resp.Result
	return &resultRecord{
		id:            *r.Id,
		contentType:   *r.ContentType,
		checksum:      *r.Checksum,
		findings:      *r.Findings,
		contentLength: uint64(*r.ContentLength),
		creationDate:  *r.CreationDate,
		metadata:      *r.Metadata,
	}, nil
}

// process is the main function that processes the files in the stream
// an error is returned only in catastrophic situations, individual file errors are recorded in the resultRecord and passed to the consumer for handling
func (s *service) process(ctx context.Context, stream chan string, maxConcurrency int, callback string, async bool, metadata map[string]string, consumer consumer) error {

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrency)

	for path := range stream {
		g.Go(func() error {

			r := resultRecord{path: path}

			fd, err := os.Open(path)
			if err != nil {
				slog.Error("could not open file", "path", path, "error", err.Error())
				r.err = err
				consumer(r)
				return nil
			}

			type writerResult struct {
				sha1 string
				err  error
			}

			pipeReader, pipeWriter := io.Pipe()
			mpb := multipart.NewWriter(pipeWriter)
			contentType := mpb.FormDataContentType()
			writeDone := make(chan writerResult, 1)

			go func() {
				defer close(writeDone)
				defer func() { _ = fd.Close() }()

				sha1 := sha1hash.New() //nolint:gosec
				fdAndShaReader := io.TeeReader(fd, sha1)

				sendErr := func(prefix string, err error) {
					_ = pipeWriter.CloseWithError(err)
					writeDone <- writerResult{err: fmt.Errorf("%s: %w", prefix, err)}
				}

				filePartWriter, err := mpb.CreateFormFile("file", filepath.Base(path))
				if err != nil {
					sendErr("create form file", err)
					return
				}

				if _, err = io.Copy(filePartWriter, fdAndShaReader); err != nil {
					sendErr("copy file", err)
					return
				}

				for k, v := range metadata {
					if err = mpb.WriteField(fmt.Sprintf("metadata[%s]", k), v); err != nil {
						sendErr("write metadata", err)
						return
					}
				}

				if callback != "" {
					if err = mpb.WriteField("callback", callback); err != nil {
						sendErr("write callback", err)
						return
					}
				}

				if err = mpb.Close(); err != nil {
					sendErr("close multipart writer", err)
					return
				}

				if err = pipeWriter.Close(); err != nil {
					writeDone <- writerResult{err: fmt.Errorf("close pipe writer: %w", err)}
					return
				}

				writeDone <- writerResult{sha1: fmt.Sprintf("%x", sha1.Sum(nil))}
			}()

			waitForWriter := func() writerResult {
				res, ok := <-writeDone
				if !ok {
					return writerResult{}
				}
				return res
			}

			handleWriterError := func(res writerResult) bool {
				if res.err == nil {
					return false
				}
				slog.Error("could not build multipart payload", "path", path, "error", res.err.Error())
				r.err = res.err
				consumer(r)
				return true
			}

			if async {
				result, err := s.client.ProcessFileAsync(ctx, contentType, pipeReader)
				if err != nil {
					_ = pipeWriter.CloseWithError(err)
					writeRes := waitForWriter()
					if handleWriterError(writeRes) {
						return nil
					}
					slog.Error("could not process file", "error", err.Error())
					r.err = err
					consumer(r)
					return nil
				}

				writeRes := waitForWriter()
				if handleWriterError(writeRes) {
					return nil
				}

				if result.StatusCode != http.StatusAccepted {
					slog.Debug("error processing file", "path", path, "status", result.StatusCode)
					r.err = fmt.Errorf("error processing file %s, status code %d", path, result.StatusCode)
				} else {
					r.id = *result.Pending.Id
					r.location = result.Header.Get("Location")
				}

				consumer(r)
			} else {
				result, localErr := s.client.ProcessFile(ctx, contentType, pipeReader)
				if localErr != nil {
					_ = pipeWriter.CloseWithError(localErr)
					writeRes := waitForWriter()
					if handleWriterError(writeRes) {
						return nil
					}
					slog.Error("could not process file", "error", localErr.Error())
					r.err = localErr
					consumer(r)
					return nil
				}

				writeRes := waitForWriter()
				if handleWriterError(writeRes) {
					return nil
				}

				calculatedSha1 := writeRes.sha1
				slog.Debug("calculated sha1", "sha1", calculatedSha1)

				slog.Debug("response", "status", result.StatusCode)

				if result.StatusCode != http.StatusCreated {
					slog.Debug("error processing file", "path", path, "status", result.StatusCode)
					r.err = fmt.Errorf("error processing file %s, status code %d", path, result.StatusCode)
				} else {
					pr := result.Result
					r.id = *pr.Id
					if pr.ContentType != nil {
						r.contentType = *pr.ContentType
					}
					if pr.Checksum != nil {
						r.checksum = *pr.Checksum
					}
					if pr.Findings != nil {
						r.findings = *pr.Findings
					}
					if pr.ContentLength != nil {
						r.contentLength = uint64(*pr.ContentLength)
					}
					if pr.CreationDate != nil {
						r.creationDate = *pr.CreationDate
					}
					if pr.Metadata != nil {
						r.metadata = *pr.Metadata
					}

					if r.checksum != calculatedSha1 {
						slog.Error("checksum mismatch", "expected", calculatedSha1, "actual", r.checksum)
						r.err = fmt.Errorf("checksum mismatch, expected %s, actual %x", calculatedSha1, r.checksum)
					} else {
						slog.Debug("checksum verified", "expected", calculatedSha1, "actual", r.checksum)
					}
				}

				consumer(r)
			}

			return nil
		})

	}
	err := g.Wait()
	if err != nil {
		return err
	}
	return nil
}
