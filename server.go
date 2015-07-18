package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"flag"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var servingDir = flag.String("serving_dir", "/Users/helder/Desktop/www", "The serving directory.")

func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func handler(w http.ResponseWriter, r *http.Request) {
	path := filepath.Clean(r.URL.Path)

	if !strings.HasPrefix(path, *servingDir) {
		http.Redirect(w, r, filepath.Join(*servingDir, path), http.StatusFound)
		return
	}

	if IsDirectory(path) {
		handleDirectory(w, r, path)
	} else {
		handleFile(w, r, path)
	}
}

func handleFile(w http.ResponseWriter, r *http.Request, path string) {
	file, err := loadFile(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", file.ContentType)
	w.Write(file.Content)
}

type FileEntry struct {
	Name    string    `json:"name"`
	ModTime time.Time `json:"time"`
}

func handleDirectory(w http.ResponseWriter, r *http.Request, path string) {
	fs, err := ioutil.ReadDir(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	fileEntries := []FileEntry{}
	for _, f := range fs {
		fileEntries = append(fileEntries, FileEntry{Name: f.Name(), ModTime: f.ModTime()})
	}

	content, err := ioutil.ReadFile("list_dir.html")
	if err != nil {
		return
	}

	j, _ := json.Marshal(fileEntries)
	out := strings.Replace(string(content), "{{ENTRIES_JSON}}", string(j), -1)
	out = strings.Replace(out, "{{CURRENT_PATH}}", path, -1)
	io.WriteString(w, out)
}

type FileContent struct {
	Content     []byte
	ContentType string
}

func loadFile(filename string) (*FileContent, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	content_type := mime.TypeByExtension(filepath.Ext(filename))
	if content_type == "" {
		content_type = "plain/text"
	}

	return &FileContent{Content: content, ContentType: content_type}, nil
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
