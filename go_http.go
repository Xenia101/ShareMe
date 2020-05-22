package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var tpl = template.Must(template.ParseGlob("code.html"))

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func SliceIndex(limit int, predicate func(i int) bool) int {
	for i := 0; i < limit; i++ {
		if predicate(i) {
			return i
		}
	}
	return -1
}

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

	hash := strings.ToUpper(GetMD5Hash(handler.Filename)[:6])
	hashed_f := handler.Filename + "." + hash
	path := filepath.Join("upload_files", hashed_f)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	err = tpl.ExecuteTemplate(w, "code.html", hash)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func download(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		return
	}
	privateCode := strings.ToUpper(r.PostFormValue("privateCode"))

	files, err := ioutil.ReadDir("./upload_files")
	if err != nil {
		log.Fatal(err)
	}

	sliceFiles := make([]string, len(files))
	for _, file := range files {
		t := strings.Split(file.Name(), ".")
		if file != nil {
			sliceFiles = append(sliceFiles, t[len(t)-1])
		}
	}
	var editFileList []string
	for _, v := range sliceFiles {
		if v != "" {
			editFileList = append(editFileList, v)
		}
	}

	isin := SliceIndex(len(editFileList), func(i int) bool { return editFileList[i] == privateCode })
	if isin != -1 {
		for _, file := range files {
			t := strings.Split(file.Name(), ".")
			if t[len(t)-1] == privateCode {
				fmt.Println(file.Name())
			}
		}
	} else {
		fmt.Println("x")
		http.Redirect(w, r, "#", 301)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("static/"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))
	mux.HandleFunc("/upload", uploadFile)
	mux.HandleFunc("/download", download)
	mux.HandleFunc("/", index)

	log.Println("Starting server on :5000")
	err := http.ListenAndServe(":5000", mux)
	log.Fatal(err)
}
