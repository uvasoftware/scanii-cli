package v22

import (
	"io"
	"log/slog"
	"net/http"
	"scanii-cli/internal/engine"
	"scanii-cli/internal/helpers"
	"strings"
)

var (
	errorNoFileSent          = "Regrettably, you did not send us any content to process - please see https://docs.scanii.com"
	errorFileAndLocationSent = "File and Location cannot be passed in at the same time"
	errorArgMissing          = "A required argument is missing"
)

type FakeHandler struct {
	engine  *engine.Engine
	baseurl string
	store   store
}

func (h FakeHandler) ProcessFileAsync(w http.ResponseWriter, r *http.Request) {
	slog.Debug("handling files")

	result := engine.Result{}
	id := generateId()
	fileFound := false
	metadata := make(map[string]string)
	var contents io.Reader

	reader, err := r.MultipartReader()
	if err != nil {
		h.renderServerError(w, err.Error())
		return
	}
	for {
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		slog.Debug("processing part", "name", part.FormName())
		if part.FormName() == "file" {
			fileFound = true
			contents = part

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
		h.renderClientError(http.StatusBadRequest, w, errorNoFileSent)
		return
	}

	// this is really a hack since processing is happening sync anyway
	result, err = h.engine.Process(contents)
	if err != nil {
		h.renderServerError(w, err.Error())
		return
	}

	if contents.(io.Closer) != nil {
		err := contents.(io.Closer).Close()
		if err != nil {
			slog.Warn("could not close file", "error", err)
		}
		// keep going
	}

	// saving metadata
	result.Metadata = metadata

	slog.Debug("saving result")
	err = h.store.save(id, &result)
	if err != nil {
		h.renderServerError(w, err.Error())
		return
	}

	foo := ProcessingPendingResponse{
		Id: &id,
	}

	headers := http.Header{}
	headers.Set("Location", h.baseurl+"/v2.2/files/"+id)

	err = helpers.WriteJSON(w, http.StatusCreated, foo, headers)
	if err != nil {
		h.renderServerError(w, err.Error())
	}

}

func (h FakeHandler) ProcessFileFetch(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h FakeHandler) RetrieveFile(w http.ResponseWriter, r *http.Request, id string) {

	if id == "" {
		h.renderClientError(http.StatusBadRequest, w, errorArgMissing)
		return
	}

	result, err := h.store.load(id)
	if err != nil {
		h.renderClientError(http.StatusNotFound, w, "Sadly, we could not find a file by that id %s")
		return
	}

	length := float32(result.ContentLength)
	resp := ProcessingResponse{
		Id:            &id,
		Findings:      &result.Findings,
		Checksum:      &result.Sha1,
		ContentLength: &length,
		ContentType:   &result.ContentType,
		Metadata:      &result.Metadata,
		CreationDate:  &result.CreationDate,
	}

	err = helpers.WriteJSON(w, http.StatusOK, resp, nil)
	if err != nil {
		h.renderServerError(w, err.Error())
	}
}

func (h FakeHandler) Ping(w http.ResponseWriter, r *http.Request) {
	key := r.Context().Value("key").(string)

	resp := map[string]string{
		"message": "pong",
		"key":     key,
	}
	err := helpers.WriteJSON(w, http.StatusOK, resp, nil)
	if err != nil {
		h.renderServerError(w, err.Error())
	}

}

func (h FakeHandler) Account(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h FakeHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h FakeHandler) DeleteToken(w http.ResponseWriter, r *http.Request, id string) {
	//TODO implement me
	panic("implement me")
}

func (h FakeHandler) GetToken(w http.ResponseWriter, r *http.Request, id string) {
	//TODO implement me
	panic("implement me")
}

func (h FakeHandler) ProcessFile(w http.ResponseWriter, r *http.Request) {

	slog.Debug("handling files")

	result := engine.Result{}
	id := generateId()
	fileFound, locationFound := false, false
	metadata := make(map[string]string)

	reader, err := r.MultipartReader()
	if err != nil {
		h.renderClientError(http.StatusMethodNotAllowed, w, err.Error())
		return
	}
	for {
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		// odd but can happen:
		if part == nil {
			continue
		}

		slog.Debug("processing part", "name", part.FormName())
		if part.FormName() == "file" {
			fileFound = true

			// performing analysis, it has to happen while we're parsing the stream
			result, err = h.engine.Process(part)
			if err != nil {
				h.renderServerError(w, err.Error())
				return
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
		if part.FormName() == "location" {
			builder := strings.Builder{}
			_, err := io.Copy(&builder, part)
			if err != nil {
				h.renderServerError(w, err.Error())
				return
			}

			locationFound = true
			location := builder.String()
			// fetching content
			slog.Debug("fetching content from", "location", location)
			resp, err := http.Get(location)
			if err != nil {
				h.renderClientError(http.StatusBadRequest, w, err.Error())
			}

			// performing analysis, it has to happen while we're parsing the stream
			result, err = h.engine.Process(resp.Body)
			if err != nil {
				h.renderServerError(w, err.Error())
				return
			}

		}
	}

	// todo: this is inefficient as we're doing the analysis twice
	if locationFound && fileFound {
		h.renderClientError(http.StatusBadRequest, w, errorFileAndLocationSent)
		return
	}

	if !fileFound {
		h.renderClientError(http.StatusBadRequest, w, errorNoFileSent)
		return
	}

	// saving metadata
	result.Metadata = metadata

	slog.Debug("saving result")
	err = h.store.save(id, &result)
	if err != nil {
		h.renderServerError(w, err.Error())
		return
	}

	if result.ContentLength == 0 {
		h.renderClientError(http.StatusBadRequest, w, errorNoFileSent)
		return
	}

	// sending response
	length := float32(result.ContentLength)

	foo := ProcessingResponse{
		Id:            &id,
		Findings:      &result.Findings,
		Checksum:      &result.Sha1,
		ContentLength: &length,
		ContentType:   &result.ContentType,
		Metadata:      &metadata,
		CreationDate:  &result.CreationDate,
	}

	headers := http.Header{}
	headers.Set("Location", h.baseurl+"/v2.2/files/"+id)

	err = helpers.WriteJSON(w, http.StatusCreated, foo, headers)
	if err != nil {
		h.renderServerError(w, err.Error())
	}

}
func (h FakeHandler) renderServerError(w http.ResponseWriter, message string) {
	h.renderClientError(http.StatusInternalServerError, w, message)
}

func (h FakeHandler) renderClientError(status int, w http.ResponseWriter, message string) {
	err := helpers.WriteJSON(w, status, ErrorResponse{
		Error: &message,
	}, nil)
	if err != nil {
		panic(err)
	}
}