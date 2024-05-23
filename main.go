package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	MaxBodyBytes = 1 * 1024 * 1024 // 1mb
)

func manageFiles(w http.ResponseWriter, req *http.Request) {
	err := req.ParseMultipartForm(MaxBodyBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get archived file from multipart form
	mForm := req.MultipartForm

	if len(mForm.File) == 0 {
		http.Error(w, "Wrong request parameters", http.StatusBadRequest)
		return
	}

	file, _, err := req.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer file.Close()

	var buf bytes.Buffer

	n, err := io.Copy(&buf, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if n == 0 {
		http.Error(w, "empty file", http.StatusBadRequest)
		return
	}

	fmt.Printf("4) incoming file size: %d\n", len(buf.Bytes()))

	unGzArch, err := UnGzip(buf.Bytes())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("5) file size after ungzip: %d\n", len(unGzArch))

	procData, err := Untar(unGzArch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("6) file size after untar: %d\n", len(procData))

	if len(procData) > 100 { // lets print 100 bytes
		procData = procData[0:100]
	}

	fmt.Println(string(procData))
}

func sendTestFile() {
	cl := NewHttpClient(HttpClientOpts{
		Timeout: 10,
	})

	err := cl.UploadFileByParts("Lovecraft.fb2", "http://127.0.0.1:8999/file")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	go sendTestFile()

	http.HandleFunc("/file", manageFiles)
	fmt.Println("Server is running at http://127.0.0.1:8999")
	log.Fatal(http.ListenAndServe(":8999", nil))
}
