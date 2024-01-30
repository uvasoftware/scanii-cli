package v22

import (
	"context"
	"io"
	"log/slog"
	"scanii-cli/internal/engine"
	"strings"
)

var (
	errorNoFileSent = "Regrettably, you did not send us any content to process - please see https://docs.scanii.com"
)

type FakeHandler struct {
	engine *engine.Engine
}

func (m FakeHandler) Account(ctx context.Context, request AccountRequestObject) (AccountResponseObject, error) {

	//TODO implement me
	panic("implement me")
}

func (m FakeHandler) CreateToken(ctx context.Context, request CreateTokenRequestObject) (CreateTokenResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (m FakeHandler) DeleteToken(ctx context.Context, request DeleteTokenRequestObject) (DeleteTokenResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (m FakeHandler) GetToken(ctx context.Context, request GetTokenRequestObject) (GetTokenResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (m FakeHandler) ProcessFile(ctx context.Context, request ProcessFileRequestObject) (ProcessFileResponseObject, error) {
	slog.Debug("handling files")
	result := engine.Result{}
	id := generateId()
	fileFound := false
	metadata := make(map[string]string)

	for {
		part, err := request.Body.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		slog.Debug("processing part", "name", part.FormName())
		if part.FormName() == "file" {
			fileFound = true
			result, err = m.engine.Process(part)
			if err != nil {
				return nil, err
			}
		}
		if strings.HasPrefix(part.FormName(), "metadata[") {
			k := extractMetadataKey(part.FormName())
			buf := strings.Builder{}
			_, err := io.Copy(&buf, part)
			if err != nil {
				slog.Warn("could not parse metadata", "key", k, "error", err.Error())
			}

			metadata[k] = buf.String()
		}
	}

	if !fileFound {
		resp := ProcessFile400JSONResponse{}
		resp.Body.Error = &errorNoFileSent
		resp.Body.Id = &id
		return resp, nil
	}

	// file found:
	length := float32(result.ContentLength)
	resp := ProcessFile201JSONResponse{}
	resp.Body.Id = &id
	resp.Body.Findings = &result.Findings
	resp.Body.Checksum = &result.Sha1
	resp.Body.ContentLength = &length
	resp.Body.ContentType = &result.ContentType
	resp.Body.Metadata = &metadata
	resp.Body.CreationDate = &result.CreationDate

	// saving metadata
	result.Metadata = metadata

	slog.Debug("saving result")
	err := defaultStore.save(id, &result)
	if err != nil {
		return nil, err
	}

	return resp, nil

}

func (m FakeHandler) ProcessFileAsync(ctx context.Context, request ProcessFileAsyncRequestObject) (ProcessFileAsyncResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (m FakeHandler) ProcessFileFetch(ctx context.Context, request ProcessFileFetchRequestObject) (ProcessFileFetchResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (m FakeHandler) RetrieveFile(ctx context.Context, request RetrieveFileRequestObject) (RetrieveFileResponseObject, error) {
	return nil, nil
}

func (m FakeHandler) Ping(ctx context.Context, request PingRequestObject) (PingResponseObject, error) {
	resp := Ping200JSONResponse{}
	message := "pong"
	key := ctx.Value("key").(string)

	resp.Body.Message = &message
	resp.Body.Key = &key

	return resp, nil
}
