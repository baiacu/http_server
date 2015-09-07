package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
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
	log.Print("Opening: " + path)

	// Special case for angular script, needed by the handleDirectory.
	if path == "/_angular.js" {
		handleFile(w, r, "_angular.js")
		return
	}

	path = filepath.Join(*servingDir, path)

	if IsDirectory(path) {
		// Is there an /path/index.html?
		if handleFileIndex(w, r, path) {
			return;
		}

		handleDirectory(w, r, path)
	} else {
		handleFile(w, r, path)
	}
}

func handleFileIndex(w http.ResponseWriter, r *http.Request, path string) bool {
	file, err := loadFile(filepath.Join(path, "/index.html"))
	if err != nil {
		return false
	}

	w.Header().Set("Content-Type", file.ContentType)
	w.Write(file.Content)
	return true
}

func handleFile(w http.ResponseWriter, r *http.Request, path string) {
	file, err := loadFile(path)
	if err != nil {
		log.Print(err)
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", file.ContentType)
	w.Write(file.Content)
}

type FileEntry struct {
	Name     string    `json:"name"`
	FullPath string    `json:"path"`
	ModTime  time.Time `json:"time"`
}

func handleDirectory(w http.ResponseWriter, r *http.Request, dir string) {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	fileEntries := []FileEntry{}
	for _, f := range fs {
		fe := FileEntry{
			Name:     f.Name(),
			FullPath: strings.TrimPrefix(filepath.Join(dir, f.Name()), *servingDir),
			ModTime:  f.ModTime(),
		}
		fileEntries = append(fileEntries, fe)
	}

	content, err := ioutil.ReadFile("list_dir.html")
	if err != nil {
		return
	}

	j, _ := json.Marshal(fileEntries)
	out := strings.Replace(string(content), "{{ENTRIES_JSON}}", string(j), -1)
	out = strings.Replace(out, "{{CURRENT_PATH}}", dir, -1)
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
		content_type = "text/plain"
	}

	return &FileContent{Content: content, ContentType: content_type}, nil
}

func main() {
	flag.Parse()
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
