package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const storageDir = "./data"

func main() {
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Fatalf("Failed to create a storageDir : %v", err)
	}
	http.HandleFunc("/file", putFileHandler)

	http.HandleFunc("/file/", getFileHandler)

	addr := ":9090"
	fmt.Println("Starting Server on ", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server Failed  : %v", err)
	}

}

func putFileHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPut {
		http.Error(w, "only PUT allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("name")
	if filename == "" {
		filename = r.Header.Get("X-Filename")
	}

	if filename == "" {
		http.Error(w, "missing 'name' query parameter or x-Filename header", http.StatusBadRequest)
		return
	}

	filename = filepath.Base(filename)
	dest := filepath.Join(storageDir, filename)

	f, err := os.Create(dest)
	if err != nil {
		http.Error(w, "failed to create file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	written, err := io.Copy(f, r.Body)
	if err != nil {
		http.Error(w, "failed to write file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Stored %s (%d bytes)\n", filename, written)

}

func getFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "only Get Allowed", http.StatusMethodNotAllowed)
		return
	}

	prefix := "/file/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		http.NotFound(w, r)
		return
	}

	raw := strings.TrimPrefix(r.URL.Path, prefix)

	filename := filepath.Base(raw)

	if filename == "" {
		http.Error(w, "Missing filename in path", http.StatusBadRequest)
		return
	}

	path := filepath.Join(storageDir, filename)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, path)
}
