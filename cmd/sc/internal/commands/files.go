package commands

import (
	"bytes"
	"context"
	//nolint // we need to use sha1 from crypto package
	sha1hash "crypto/sha1"
	"errors"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	v22 "github.com/uvasoftware/scanii-cli/internal/v22"
	"io"
	"io/fs"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func FileCommand(ctx context.Context) *cobra.Command {

	var metadata, callback string

	init := func() (*v22.Client, error) {
		config, err := loadConfig()
		if err != nil {
			return nil, err
		}
		fmt.Println("⎻⎻⎻⎻")
		fmt.Printf("Using endpoint: %s and API key: %s\n", config.Endpoint, config.APIKey)

		client, err := createClient(config)
		if err != nil {
			return nil, err
		}

		// validating credentials
		if ok, err := callPingEndpoint(ctx, client); ok {
			fmt.Printf("✔ Credentials worked against %s\n", client.Server)

		} else {
			fmt.Printf("❌ Credentials failed against %s\n", client.Server)

			return nil, err
		}

		return client, nil
	}

	parent := cobra.Command{
		Use:   "files",
		Short: "API operations for the files resource",
		Long:  `Files API operations. Detailed API documentation can be found here: https://uvasoftware.github.io/openapi/v22/#/Files`,
	}

	parent.PersistentFlags().StringVarP(&metadata, "metadata", "m", "", "Metadata in the format key=value,key2=value2 to be associated with the request")

	concurrencyLimit := 32
	ignoreHidden := false
	processFileCmd := cobra.Command{
		Use:        "process [flags] [path]",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"file/directory"},
		Short:      "Process a local file synchronously",
		Long: `Process a local file synchronously. The file can be a single file or a directory. 
If a directory is provided, all files in the directory will be processed recursively.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := init()
			if err != nil {
				return err
			}

			_, err = callFileProcess(ctx, client, args[0], concurrencyLimit, ignoreHidden, metadata, false)
			return err
		},
	}

	processFileCmd.PersistentFlags().IntVarP(&concurrencyLimit, "concurrency", "p", 32, "Number of concurrent requests to use")
	processFileCmd.PersistentFlags().BoolVarP(&ignoreHidden, "ignore-hidden", "i", true, "Ignore hidden files")
	parent.AddCommand(&processFileCmd)

	asyncCmd := &cobra.Command{
		Use:        "async [flags] [file]",
		Short:      "Process a local file asynchronously",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"file/directory"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := init()
			if err != nil {
				return err
			}
			_, err = callFileProcess(ctx, client, args[0], concurrencyLimit, ignoreHidden, metadata, true)
			return err
		},
	}

	asyncCmd.PersistentFlags().StringVarP(&callback, "callback", "c", "", "Callback URL to be invoked when processing is complete")
	asyncCmd.PersistentFlags().IntVarP(&concurrencyLimit, "concurrency", "p", 32, "Number of concurrent requests to use")
	asyncCmd.PersistentFlags().BoolVarP(&ignoreHidden, "ignore-hidden", "i", true, "Ignore hidden files")
	parent.AddCommand(asyncCmd)

	fetchCmd := cobra.Command{
		Use:        "fetch [flags] [url]",
		Short:      "Process a remote URL asynchronously",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"file/directory"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := init()
			if err != nil {
				return err
			}

			_, err = callFilesFetch(ctx, client, args[0], callback, metadata)
			return err

		},
	}

	fetchCmd.PersistentFlags().StringVarP(&callback, "callback", "c", "", "Callback URL to be invoked when processing is complete")
	parent.AddCommand(&fetchCmd)

	retrieveFileCmd := &cobra.Command{
		Use:        "retrieve [id]",
		Short:      "Retrieves a previously processed file by id",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"id"},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := init()
			if err != nil {
				return err
			}

			_, err = callFileRetrieve(ctx, client, args[0])
			return err
		},
	}

	parent.AddCommand(retrieveFileCmd)
	return &parent

}

func callFileRetrieve(ctx context.Context, client *v22.Client, s string) (*resultRecord, error) {
	if s == "" {
		return nil, errors.New("id cannot be empty")
	}

	startTime := time.Now()
	slog.Debug("retrieving file", "id", s)

	file, err := client.RetrieveFile(ctx, s)
	if err != nil {
		return nil, err
	}
	defer file.Body.Close()

	if file.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error retrieving file with id %s, status code %d", s, file.StatusCode)
	}

	parsedResult, err := v22.ParseRetrieveFileResponse(file)
	if err != nil {
		return nil, err
	}
	fmt.Printf("✔ Result retrieved in %s\n", time.Since(startTime))

	result := resultRecord{
		id: *parsedResult.JSON200.Id,
	}
	if parsedResult.JSON200.Metadata != nil {
		result.metadata = *parsedResult.JSON200.Metadata
	}

	if parsedResult.JSON200.ContentType != nil {
		result.contentType = *parsedResult.JSON200.ContentType
	}
	if parsedResult.JSON200.Checksum != nil {
		result.checksum = *parsedResult.JSON200.Checksum
	}
	if parsedResult.JSON200.Findings != nil {
		result.findings = *parsedResult.JSON200.Findings
	}
	if parsedResult.JSON200.ContentLength != nil {
		result.contentLength = uint64(*parsedResult.JSON200.ContentLength)
	}
	if parsedResult.JSON200.CreationDate != nil {
		result.creationDate = *parsedResult.JSON200.CreationDate
	}
	if parsedResult.JSON200.Error != nil {
		result.err = *parsedResult.JSON200.Error
	}

	printFileResult(&result)
	return &result, nil
}

// callFilesFetch processes a remote url
func callFilesFetch(ctx context.Context, client *v22.Client, location, callback, metadata string) (*resultRecord, error) {
	slog.Debug("processing location", "url", location)

	// verifying url
	if _, err := url.Parse(location); err != nil {
		println("Unable to parse url", err.Error())
		return nil, err
	}

	m := extractMedata(metadata)

	// because of how we pass metadata arguments, we must manually encode the payload
	form := url.Values{}
	for k, v := range m {
		form.Add(fmt.Sprintf("metadata[%s]", k), v)
	}
	form.Add("location", location)

	if callback != "" {
		form.Add("callback", callback)
	}

	slog.Debug("form", "form", form.Encode())
	httpResponse, err := client.ProcessFileFetchWithBody(ctx, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	defer httpResponse.Body.Close()

	parsedResult, err := v22.ParseProcessFileFetchResponse(httpResponse)
	if err != nil {
		return nil, err
	}

	if httpResponse.StatusCode != http.StatusAccepted {
		slog.Debug("error processing file", "status", httpResponse.StatusCode, "body", string(parsedResult.Body))

		return nil, fmt.Errorf("error: %s", *parsedResult.JSON400.Error)
	}

	fmt.Println()
	fmt.Printf("Request accepted for processing with id %s\n", *parsedResult.JSON202.Id)
	if callback != "" {
		fmt.Printf("Callback URL: %s\n", callback)
	}
	fmt.Printf("Location: %s\n", httpResponse.Header.Get("Location"))
	fmt.Println()
	fmt.Printf("Protip: you can retrieve the result by running: `sc files retrieve %s`\n", *parsedResult.JSON202.Id)

	return &resultRecord{
		id:       *parsedResult.JSON202.Id,
		location: httpResponse.Header.Get("Location"),
	}, nil
}

type resultRecord struct {
	path          string
	err           string
	contentType   string
	findings      []string
	checksum      string
	id            string
	location      string
	contentLength uint64
	creationDate  string
	metadata      map[string]string
}

// callFileProcess processes a file or directory
func callFileProcess(ctx context.Context, client *v22.Client, path string, concurrencyLimit int, ignoreHidden bool, metadata string, async bool) ([]*resultRecord, error) {
	slog.Debug("processing file", "path", path)
	slog.Debug("concurrency limit", "limit", concurrencyLimit)
	slog.Debug("ignore hidden", "ignore", ignoreHidden)

	startTime := time.Now()

	stat, err := os.Stat(path)
	if err != nil {
		println("Unable to stat path", err.Error())
		return nil, err
	}
	if stat.IsDir() {
		fmt.Println("Processing directory: ", path)
	} else {
		fmt.Println("Processing file: ", path)
	}

	var files []string

	err = filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if ignoreHidden {
			if strings.HasPrefix(path, ".") && !strings.HasPrefix(path, "./") {
				slog.Debug("skipping hidden file", "path", path)
				return nil
			}
		}

		if !info.Mode().IsRegular() {
			slog.Debug("skipping irregular file", "path", path)
			return nil
		}
		if info.IsDir() {
			slog.Debug("skipping directory", "path", path)
			return nil
		}

		if info.Size() == 0 {
			slog.Warn("ignoring empty file", "path", path)
			return nil

		}

		files = append(files, path)
		return nil

	})

	if err != nil {
		return nil, err
	}

	if len(files) > 1 {
		fmt.Printf("Found %s file(s)\n", humanize.Comma(int64(len(files))))
	}

	workChan := make(chan string)
	resultChan := make(chan resultRecord)

	// starting workers, if file count is less than concurrency, we will only start as many as needed
	if len(files) < concurrencyLimit {
		concurrencyLimit = len(files)
		slog.Debug("reducing concurrency limit", "limit", concurrencyLimit)
	}

	for i := 0; i < concurrencyLimit; i++ {
		workerNum := i
		// spawning worker:
		go func() {
			slog.Debug("working started", "worker", workerNum)
			for p := range workChan {
				r := resultRecord{
					path: p,
				}
				calculatedSha1 := ""
				fd, err := os.Open(p)
				if err != nil {
					slog.Error("could not open file", "path", p, "error", err.Error())
					r.err = err.Error()
					resultChan <- r
				}

				// connects the http clients with the multipart producer
				pipeReader, pipeWriter := io.Pipe()

				mpb := multipart.NewWriter(pipeWriter)

				// firing off multipart stream producer, so we don't need to buffer it in-memory
				go func() {
					// let's also verify the checksum while reading the file to avoid buffering it in memory
					sha1 := sha1hash.New() //nolint:gosec
					fdAndShaReader := io.TeeReader(fd, sha1)

					filePartWriter, err := mpb.CreateFormFile("file", filepath.Base(p))
					if err != nil {
						slog.Error("could not create form field", "error", err.Error())
						_ = pipeWriter.CloseWithError(err)
						return
					}

					if _, err := io.Copy(filePartWriter, fdAndShaReader); err != nil {
						slog.Error("could not copy file to form field", "error", err.Error())
						_ = pipeWriter.CloseWithError(err)
						return
					}
					calculatedSha1 = fmt.Sprintf("%x", sha1.Sum(nil))
					slog.Debug("calculated sha1", "sha1", calculatedSha1)

					m := extractMedata(metadata)

					// because of how we pass metadata arguments, we must manually encode the payload
					for k, v := range m {
						if err = mpb.WriteField(fmt.Sprintf("metadata[%s]", k), v); err != nil {
							slog.Error("could not write metadata", "key", k, "value", v, "error", err.Error())
							_ = pipeWriter.CloseWithError(err)
							return
						}
					}

					_ = mpb.Close()
					_ = pipeWriter.Close()

				}()

				switch async {
				case false:
					var localErr error

					//nolint
					resp, localErr := client.ProcessFileWithBody(ctx, mpb.FormDataContentType(), pipeReader)
					if localErr != nil {
						slog.Error("could not process file", "error", localErr.Error())
						r.err = localErr.Error()
						break
					}

					v, localErr := v22.ParseProcessFileResponse(resp)

					if localErr != nil {
						slog.Error("could not parse response", "error", localErr.Error())
						r.err = localErr.Error()
						break
					}
					if v.StatusCode() != http.StatusCreated {
						slog.Debug("error processing file", "path", path, "status", v.StatusCode(), "body", string(v.Body))
						r.err = fmt.Errorf("error processing file %s, status code %d", path, v.StatusCode()).Error()
					} else {
						r.id = *v.JSON201.Id
						if v.JSON201.ContentType != nil {
							r.contentType = *v.JSON201.ContentType
						}
						// if not nil, copy values over:
						if v.JSON201.Checksum != nil {
							r.checksum = *v.JSON201.Checksum
						}
						if v.JSON201.Findings != nil {
							r.findings = *v.JSON201.Findings
						}
						if v.JSON201.ContentLength != nil {
							r.contentLength = uint64(*v.JSON201.ContentLength)
						}
						if v.JSON201.CreationDate != nil {
							r.creationDate = *v.JSON201.CreationDate
						}
						if v.JSON201.Metadata != nil {
							r.metadata = *v.JSON201.Metadata
						}

						// verifying checksum
						if r.checksum != calculatedSha1 {
							slog.Error("checksum mismatch", "expected", calculatedSha1, "actual", r.checksum)
							r.err = fmt.Errorf("checksum mismatch, expected %s, actual %x", calculatedSha1, r.checksum).Error()

						} else {
							slog.Debug("checksum verified", "expected", calculatedSha1, "actual", r.checksum)
						}

					}
				case true:
					//nolint:bodyclose // parse closes the body
					resp, err := client.ProcessFileAsyncWithBody(ctx, mpb.FormDataContentType(), pipeReader)
					if err != nil {
						slog.Error("could not process file", "error", err.Error())
						break
					}

					v, err := v22.ParseProcessFileAsyncResponse(resp)
					if err != nil {
						slog.Error("could not parse response", "error", err.Error())
						break
					}
					if v.StatusCode() != http.StatusAccepted {
						slog.Debug("error processing file", "path", path, "status", v.StatusCode(), "body", string(v.Body))
						r.err = fmt.Errorf("error processing file %s, status code %d", path, v.StatusCode()).Error()

					} else {
						r.id = *v.JSON202.Id
						r.location = resp.Header.Get("Location")
					}
				}

				_ = fd.Close()
				// sending result
				resultChan <- r
			}
			slog.Debug("worker shutting down", "worker", workerNum)
		}()
	}

	go func() {
		// publishing files to be processed:
		for _, f := range files {
			slog.Debug("publishing file", "path", f)
			workChan <- f
		}
	}()

	pb := progressbar.Default(int64(len(files)), "Processing files")
	results := make([]*resultRecord, 0)
	successCounter, errorCounter, findingsCounter := 0, 0, 0

	// waiting for results:
	for i := 0; i < len(files); i++ {
		result := <-resultChan
		results = append(results, &result)

		if result.err != "" {
			errorCounter++
		} else {
			successCounter++
		}

		if len(result.findings) > 0 {
			findingsCounter++
		}

		err = pb.Add(1)
		if err != nil {
			return nil, err
		}
	}

	// shutting down worker pools:
	close(workChan)

	fmt.Println()
	fmt.Printf("✔ Completed in %s\n", time.Since(startTime))

	// shorter output for async
	if async {
		for _, r := range results {
			printFileResult(r)
		}

		return results, nil

	}

	if len(results) == 1 {
		// single result mode
		r := results[0]
		printFileResult(r)
		return results, nil

	}

	// multi result mode:
	fmt.Printf("✔ Files with findings: %d, unable to process: %d and successfully processed: %d\n", findingsCounter, errorCounter, successCounter)
	if findingsCounter > 0 {
		fmt.Printf("Files with findings:\n")
		for _, r := range results {
			if len(r.findings) > 0 {
				printFileResult(r)
			}
		}
	}
	return results, nil

}

// extractMedata parses the metadata string and returns a map of key/value pairs
func extractMedata(metadata string) map[string]string {
	result := make(map[string]string)
	if metadata == "" {
		return result
	}

	parts := strings.Split(metadata, ",")
	for _, p := range parts {
		kv := strings.Split(p, "=")
		if len(kv) != 2 {
			slog.Warn("invalid metadata entry", "entry", p)
			continue
		}
		result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}
	return result
}

func printFileResult(result *resultRecord) {
	fmt.Printf("------\n")

	if result.path != "" {
		fmt.Printf("  %-15s %s\n", "path:", result.path)
	}

	if result.err != "" {
		fmt.Printf("  %-15s %s\n", "error:", result.err)
		return
	} else {

		fmt.Printf("  %-15s %s\n", "id:", result.id)

		if result.checksum != "" {
			fmt.Printf("  %-15s %s\n", "checksum/sha1:", result.checksum)
		}

		if result.location != "" {
			fmt.Printf("  %-15s %s\n", "location:", result.location)
		}

		if result.contentType != "" {
			fmt.Printf("  %-15s %s\n", "content type:", result.contentType)
		}

		if result.contentLength != 0 {
			fmt.Printf("  %-15s %s\n", "content length:", humanize.Bytes(result.contentLength))
		}

		if result.creationDate != "" {
			fmt.Printf("  %-15s %s\n", "creation date:", result.creationDate)
		}

		if len(result.findings) > 0 {
			fmt.Printf("  %-15s %s\n", "findings:", strings.Join(result.findings, ","))
		} else {
			fmt.Printf("  %-15s none\n", "findings:")
		}

		if len(result.metadata) > 0 {
			fmt.Printf("  metadata:\n")
			for k, v := range result.metadata {
				fmt.Printf("    %s → %s\n", k, v)
			}
		} else {
			fmt.Printf("  %-15s none\n", "metadata:")
		}
	}

	fmt.Printf("------\n")
}

func runLocationProcess(ctx context.Context, client *v22.Client, location, callback, metadata string) (*resultRecord, error) {
	slog.Debug("processing location", "url", location)

	// verifying url
	if _, err := url.Parse(location); err != nil {
		println("Unable to parse url", err.Error())
		return nil, err
	}

	m := extractMedata(metadata)

	// because of how we pass metadata arguments, we must manually encode the payload
	body := bytes.Buffer{}
	mp := multipart.NewWriter(&body)

	err := mp.WriteField("location", location)
	if err != nil {
		return nil, err
	}

	for k, v := range m {
		err = mp.WriteField(fmt.Sprintf("metadata[%s]", k), v)
		if err != nil {
			return nil, err
		}
	}

	if callback != "" {
		err = mp.WriteField("callback", callback)
		if err != nil {
			return nil, err
		}
	}
	err = mp.Close()
	if err != nil {
		return nil, err
	}

	httpResponse, err := client.ProcessFileWithBody(ctx, mp.FormDataContentType(), &body)
	if err != nil {
		return nil, err
	}

	defer httpResponse.Body.Close()

	parsedResult, err := v22.ParseProcessFileResponse(httpResponse)
	if err != nil {
		return nil, err
	}

	if httpResponse.StatusCode != http.StatusCreated {
		slog.Debug("error processing file", "status", httpResponse.StatusCode, "body", string(parsedResult.Body))

		return nil, fmt.Errorf("error: %s", *parsedResult.JSON400.Error)
	}

	r := resultRecord{}
	r.id = *parsedResult.JSON201.Id
	if parsedResult.JSON201.ContentType != nil {
		r.contentType = *parsedResult.JSON201.ContentType
	}
	// if not nil, copy values over:
	if parsedResult.JSON201.Checksum != nil {
		r.checksum = *parsedResult.JSON201.Checksum
	}
	if parsedResult.JSON201.Findings != nil {
		r.findings = *parsedResult.JSON201.Findings
	}
	if parsedResult.JSON201.ContentLength != nil {
		r.contentLength = uint64(*parsedResult.JSON201.ContentLength)
	}
	if parsedResult.JSON201.CreationDate != nil {
		r.creationDate = *parsedResult.JSON201.CreationDate
	}
	if parsedResult.JSON201.Metadata != nil {
		r.metadata = *parsedResult.JSON201.Metadata
	}

	printFileResult(&r)
	return &r, nil
}
