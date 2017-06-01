package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockS3Uploader struct {
	Resp *UploadResult
}

func (u mockS3Uploader) Upload(file io.Reader, key string) (*UploadResult, error) {
	return u.Resp, nil
}

func TestUploadPhoto(t *testing.T) {
	assert := assert.New(t)
	result := &UploadResult{
		Location:  "https://location.com/1",
		VersionID: nil,
		UploadID:  "1",
	}

	NewUploaderManager = func() UploaderManager {
		return mockS3Uploader{Resp: result}
	}

	path := os.Getenv("GOPATH") + "/pi.png"
	file, _ := os.Open(path)
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	description, _ := writer.CreateFormField("description")
	io.WriteString(description, "Photo Description")

	part, _ := writer.CreateFormFile("photo", filepath.Base(path))
	io.Copy(part, file)
	writer.Close()
	mux := NewMux()
	res := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	mux.ServeHTTP(res, req)

	expectedResult := new(UploadResult)
	json.NewDecoder(res.Body).Decode(expectedResult)
	assert.Equal(res.Code, http.StatusCreated)
	assert.Equal(expectedResult.Location, result.Location)
	assert.Equal(expectedResult.UploadID, result.UploadID)
}
