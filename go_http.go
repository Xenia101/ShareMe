package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func uploadFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the file")
		fmt.Println(err)
		return
	}

	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	filename := strings.Split(handler.Filename, ".")
	tempName := "*." + filename[1]

	tempFile, err := ioutil.TempFile("upload_files", tempName)
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	tempFile.Write(fileBytes)

	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func setupRoutes() {
	http.HandleFunc("/", index)
	http.HandleFunc("/upload", uploadFile)
	http.ListenAndServe(":5000", nil)
}

func main() {
	fmt.Println("hello World")
	setupRoutes()
}
