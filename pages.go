package main

import (
	"html/template"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	minio "github.com/minio/minio-go"
)

// BucketPage defines the details page of a bucket
type BucketPage struct {
	BucketName string
	Objects    []ObjectWithIcon
}

// ObjectWithIcon is a minio object with an added icon
type ObjectWithIcon struct {
	minio.ObjectInfo
	Icon string
}

// BucketPageHandler shows the details page of a bucket
func (s *Server) BucketPageHandler(w http.ResponseWriter, r *http.Request) {
	bucketName := mux.Vars(r)["bucketName"]
	var objects []ObjectWithIcon

	lp := path.Join("templates", "layout.html")
	bp := path.Join("templates", "bucket.html")

	t, err := template.ParseFiles(lp, bp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	doneCh := make(chan struct{})

	objectCh := s.s3.ListObjectsV2(bucketName, "", false, doneCh)
	for object := range objectCh {
		if object.Err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		objectWithIcon := ObjectWithIcon{object, icon(object.Key)}
		objects = append(objects, objectWithIcon)
	}

	bucketPage := BucketPage{
		BucketName: bucketName,
		Objects:    objects,
	}

	err = t.ExecuteTemplate(w, "layout", bucketPage)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// BucketsPageHandler shows all buckets
func (s *Server) BucketsPageHandler(w http.ResponseWriter, r *http.Request) {
	lp := path.Join("templates", "layout.html")
	ip := path.Join("templates", "index.html")

	t, err := template.ParseFiles(lp, ip)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	buckets, err := s.s3.ListBuckets()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = t.ExecuteTemplate(w, "layout", buckets)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// IndexHandler forwards to "/buckets"
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/buckets", http.StatusPermanentRedirect)
}

// icon returns an icon for a file type
func icon(fileName string) string {
	e := path.Ext(fileName)

	switch e {
	case ".tgz":
		return "archive"
	case ".png", ".jpg", ".gif", ".svg":
		return "photo"
	case ".mp3":
		return "music_note"
	}

	return "insert_drive_file"
}
