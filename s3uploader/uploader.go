package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bmizerany/pat"
)

func init() {
	mySession = session.New()
}

func NewMux() http.Handler {
	m := pat.New()
	m.Post("/upload", http.HandlerFunc(UploaderHandler))
	m.Get("/", http.FileServer(http.Dir("public")))

	return m
}

func main() {
	http.ListenAndServe(":3000", NewMux())
}

var bucketName = "bk-yesno"
var mySession *session.Session

type UploadResult struct {
	Location  string  `json:"location"`
	VersionID *string `json:"version_id"`
	UploadID  string  `json:"upload_id"`
}

type UploaderManager interface {
	Upload(file io.Reader, key string) (*UploadResult, error)
}

type uploaderManager struct {
	Uploader *s3manager.Uploader
	Bucket   string
}

func (u uploaderManager) Upload(file io.Reader, key string) (*UploadResult, error) {
	params := &s3manager.UploadInput{
		Bucket: aws.String(u.Bucket),
		Key:    aws.String(key),
		Body:   file,
	}

	result, err := u.Uploader.Upload(params)
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		Location:  result.Location,
		VersionID: result.VersionID,
		UploadID:  result.UploadID,
	}, nil
}

func UploaderHandler(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("photo")

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	uploader := NewUploaderManager()

	result, err := uploader.Upload(file, handler.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(result)
}

var NewUploaderManager = func() UploaderManager {
	return uploaderManager{
		Uploader: s3manager.NewUploader(mySession),
		Bucket:   bucketName,
	}
}
