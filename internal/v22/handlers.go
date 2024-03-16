package v22

import (
	"errors"
	"fmt"
	"github.com/uvasoftware/scanii-cli/internal/engine"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

var (
	errorNoFileSent          = "Regrettably, you did not send us any content to process - please see https://docs.scanii.com"
	errorFileAndLocationSent = "File and Location cannot be passed in at the same time"
	errorArgMissing          = "A required argument is missing"
	errorCloudNotDownload    = "Sadly, we could not download content for processing due to a network error."
)

const basePath = "/v2.2/files/"

type FakeHandler struct {
	engine  *engine.Engine
	baseurl string
	store   store
}

func (h FakeHandler) ProcessFileAsync(w http.ResponseWriter, r *http.Request) {
	id := generateID()
	result := engine.Result{ID: id}
	fileFound := false
	metadata := make(map[string]string)

	reader, err := r.MultipartReader()
	if err != nil {
		h.renderClientError(http.StatusMethodNotAllowed, w, err.Error())
		return
	}
	for {
		part, err := reader.NextPart()
		if err != nil {
			if errors.Is(err, io.EOF) {
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
	resp := ProcessingPendingResponse{
		Id: &id,
	}

	headers := http.Header{}
	headers.Set("Location", h.baseurl+basePath+id)

	err = writeJSON(w, http.StatusAccepted, resp, headers)
	if err != nil {
		h.renderServerError(w, err.Error())
	}
}

func (h FakeHandler) ProcessFileFetch(w http.ResponseWriter, r *http.Request) {

	id := generateID()
	metadata := make(map[string]string)
	result := engine.Result{
		ID: id,
	}

	err := r.ParseForm()
	if err != nil {
		h.renderClientError(http.StatusBadRequest, w, err.Error())
		return
	}

	location := r.Form.Get("location")
	if location == "" {
		h.renderClientError(http.StatusBadRequest, w, errorArgMissing)
		return
	}

	// extracting metadata
	for k, v := range r.Form {
		if strings.HasPrefix(k, "metadata[") {
			metadata[extractMetadataKey(k)] = v[0]
		}
	}

	// fetching content
	slog.Debug("fetching content from", "location", location)
	//nolint:gosec
	request, err := http.NewRequestWithContext(r.Context(), http.MethodGet, location, http.NoBody)
	if err != nil {
		h.renderServerError(w, err.Error())
		return
	}

	httpResponse, err := http.DefaultClient.Do(request)
	if err != nil {
		h.renderClientError(http.StatusBadRequest, w, err.Error())
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode == http.StatusOK {
		// performing analysis, it has to happen while we're parsing the stream
		result, err = h.engine.Process(httpResponse.Body)
		if err != nil {
			h.renderServerError(w, err.Error())
			return
		}
	} else {
		result.Error = errorCloudNotDownload
	}

	// saving result
	result.Metadata = metadata
	err = h.store.save(id, &result)
	if err != nil {
		h.renderServerError(w, err.Error())
		return
	}

	headers := http.Header{}
	headers.Set("Location", h.baseurl+basePath+id)

	// sending response
	resp := ProcessingPendingResponse{
		Id: &id,
	}

	err = writeJSON(w, http.StatusAccepted, resp, headers)
	if err != nil {
		h.renderServerError(w, err.Error())
	}
}

func (h FakeHandler) RetrieveFile(w http.ResponseWriter, _ *http.Request, id string) {
	if id == "" {
		h.renderClientError(http.StatusBadRequest, w, errorArgMissing)
		return
	}

	result := engine.Result{}
	err := h.store.load(id, &result)
	if err != nil {
		h.renderClientError(http.StatusNotFound, w, "Sadly, we could not find a file by that id %s")
		return
	}

	if result.Error != "" {
		resp := ErrorResponse{
			Error:    &result.Error,
			Metadata: &result.Metadata,
			Id:       &result.ID,
		}

		err = writeJSON(w, http.StatusOK, resp, nil)
		if err != nil {
			h.renderServerError(w, err.Error())
		}
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

	err = writeJSON(w, http.StatusOK, resp, nil)
	if err != nil {
		h.renderServerError(w, err.Error())
	}
}

func (h FakeHandler) Ping(w http.ResponseWriter, r *http.Request) {
	key := r.Context().Value(keyInContext).(string)

	resp := map[string]string{
		"message": "pong",
		"key":     key,
	}
	err := writeJSON(w, http.StatusOK, resp, nil)
	if err != nil {
		h.renderServerError(w, err.Error())
	}

}

func (h FakeHandler) Account(w http.ResponseWriter, r *http.Request) {

	key := r.Context().Value(keyInContext).(string)

	// account basically returns made up stuff
	balance := float32(42_000)
	startingBalance := float32(100_000)
	billingEmail := "admin@example.com"
	accountName := "ACME Inc."
	creationDate := time.Now().UTC().AddDate(0, 0, -30).Format(time.RFC3339)
	accountSub := "Premium"

	// keys
	keyCreationDate := time.Now().UTC().AddDate(0, 0, -30).Format(time.RFC3339)
	lastSeenDate := time.Now().UTC().AddDate(0, 0, -30).Format(time.RFC3339)
	keyDetectionCategoriesEnabled := []string{"malware", "unsafe_language", "unsafe_image"}
	keyActive := true
	keyTags := []string{"tag1", "tag2"}

	apiKey := &ApiKey{
		Active:                     &keyActive,
		CreationDate:               &keyCreationDate,
		DetectionCategoriesEnabled: &keyDetectionCategoriesEnabled,
		LastSeenDate:               &lastSeenDate,
		Tags:                       &keyTags,
	}

	// user:
	userCreationDate := time.Now().UTC().AddDate(0, 0, -30).Format(time.RFC3339)

	user1 := &User{
		CreationDate: &userCreationDate,
		LastLogin:    nil,
	}

	resp := AccountInfo{
		Balance:          &balance,
		BillingEmail:     &billingEmail,
		CreationDate:     &creationDate,
		Keys:             &map[string]ApiKey{key: *apiKey},
		ModificationDate: &creationDate,
		Name:             &accountName,
		StartingBalance:  &startingBalance,
		Subscription:     &accountSub,
		Users:            &map[string]User{"bob@example.com": *user1},
	}

	err := writeJSON(w, http.StatusOK, resp, nil)
	if err != nil {
		h.renderServerError(w, err.Error())
	}

}

func (h FakeHandler) CreateToken(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		h.renderServerError(w, err.Error())
	}
	timeoutInSeconds := 300
	if r.Form.Get("timeout") != "" {
		timeoutInSeconds, err = strconv.Atoi(r.Form.Get("timeout"))
		if err != nil {
			h.renderClientError(http.StatusBadRequest, w, "could not parse timeout")
			return
		}
	}

	id := generateID()
	creationDate := time.Now().UTC().Format(time.RFC3339)
	expirationDate := time.Now().UTC().Add(time.Second * time.Duration(timeoutInSeconds)).Format(time.RFC3339)
	token := &AuthToken{
		CreationDate:   &creationDate,
		ExpirationDate: &expirationDate,
		Id:             &id,
	}

	err = h.store.save(id, token)
	if err != nil {
		h.renderServerError(w, err.Error())
		return
	}

	err = writeJSON(w, http.StatusCreated, token, nil)
	if err != nil {
		h.renderServerError(w, err.Error())
	}

}

func (h FakeHandler) DeleteToken(w http.ResponseWriter, _ *http.Request, id string) {
	found, err := h.store.remove(id)
	if !found {
		h.renderClientError(http.StatusNotFound, w, "could not find token")
		return
	}

	if err != nil {
		h.renderServerError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h FakeHandler) RetrieveToken(w http.ResponseWriter, _ *http.Request, id string) {
	token := &AuthToken{}
	err := h.store.load(id, token)
	if err != nil {
		h.renderClientError(http.StatusNotFound, w, "could not find token")
		return
	}

	err = writeJSON(w, http.StatusOK, token, nil)
	if err != nil {
		h.renderServerError(w, err.Error())
	}
}

func (h FakeHandler) ProcessFile(w http.ResponseWriter, r *http.Request) {

	result := engine.Result{}
	id := generateID()
	fileFound, locationFound := false, false
	metadata := make(map[string]string)

	reader, err := r.MultipartReader()
	if err != nil {
		h.renderClientError(http.StatusMethodNotAllowed, w, err.Error())
		return
	}
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
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
			req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, location, http.NoBody)
			if err != nil {
				h.renderServerError(w, err.Error())
				return
			}
			resp, err := http.DefaultClient.Do(req) //nolint:gosec
			if err != nil {
				h.renderClientError(http.StatusBadRequest, w, err.Error())
				return
			}
			defer resp.Body.Close() //nolint
			if resp.StatusCode == http.StatusOK {
				// performing analysis, it has to happen while we're parsing the stream
				result, err = h.engine.Process(resp.Body)
				if err != nil {
					h.renderServerError(w, err.Error())
					return
				}
			} else {
				result.Error = errorCloudNotDownload
			}

		}
	}

	// todo: this is inefficient as we're doing the analysis twice
	if locationFound && fileFound {
		h.renderClientError(http.StatusBadRequest, w, errorFileAndLocationSent)
		return
	}

	if !fileFound && !locationFound {
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
	headers.Set("Location", h.baseurl+basePath+id)

	err = writeJSON(w, http.StatusCreated, foo, headers)
	if err != nil {
		h.renderServerError(w, err.Error())
	}

}
func (h FakeHandler) renderServerError(w http.ResponseWriter, message string) {
	trace := fmt.Sprintf("%s\n%s", message, debug.Stack())
	slog.Error(trace)
	h.renderClientError(http.StatusInternalServerError, w, message)
}

func (h FakeHandler) renderClientError(status int, w http.ResponseWriter, message string) {
	err := writeJSON(w, status, ErrorResponse{
		Error: &message,
	}, nil)
	if err != nil {
		panic(err)
	}
}
