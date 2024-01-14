package v22

import (
	"context"
	"io"
	"log/slog"
	"scanii-cli/internal/engine"
)

type MockHandler struct {
	engine *engine.Engine
}

func (m MockHandler) Account(ctx context.Context, request AccountRequestObject) (AccountResponseObject, error) {

	//TODO implement me
	panic("implement me")
}

func (m MockHandler) CreateToken(ctx context.Context, request CreateTokenRequestObject) (CreateTokenResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockHandler) DeleteToken(ctx context.Context, request DeleteTokenRequestObject) (DeleteTokenResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockHandler) GetToken(ctx context.Context, request GetTokenRequestObject) (GetTokenResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockHandler) ProcessFile(ctx context.Context, request ProcessFileRequestObject) (ProcessFileResponseObject, error) {
	slog.Debug("handling files")
	result := engine.Result{}

	for {
		part, err := request.Body.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		slog.Debug("processing part", "name", part.FormName())
		result, err = m.engine.Process(part)
		if err != nil {
			return nil, err
		}
	}
	length := float32(result.ContentLength)

	resp := ProcessFile201JSONResponse{}
	resp.Body.Findings = &result.Findings
	resp.Body.Checksum = &result.Sha1
	resp.Body.ContentLength = &length
	resp.Body.Metadata = &map[string]MetadataObject{
		"hello": {
			"foo": "bar",
		},
	}

	return resp, nil
}

func (m MockHandler) ProcessFileAsync(ctx context.Context, request ProcessFileAsyncRequestObject) (ProcessFileAsyncResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockHandler) ProcessFileFetch(ctx context.Context, request ProcessFileFetchRequestObject) (ProcessFileFetchResponseObject, error) {
	//TODO implement me
	panic("implement me")
}

func (m MockHandler) RetrieveFile(ctx context.Context, request RetrieveFileRequestObject) (RetrieveFileResponseObject, error) {
	return nil, nil
}

func (m MockHandler) Ping(ctx context.Context, request PingRequestObject) (PingResponseObject, error) {
	resp := Ping200JSONResponse{}
	message := "pong"
	key := "124"
	resp.Body.Message = &message
	resp.Body.Key = &key

	return resp, nil
}
