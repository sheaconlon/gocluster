package gocluster

import (
	"os"
	"net/http"
	"time"
	"context"
	"io"
	"bytes"
	"mime/multipart"
	"strconv"
)

// serveNPollInterval is the interval at which serveN checks whether it should shut down the server.
const serveNPollInterval = 5 * time.Second

// serveNShutdownTimeout is the timeout serveN uses when shutting down the server.
const serveNShutdownTimeout = 10 * time.Second

// megabyte is the number of bytes in a megabyte.
const megabyte = 1000000

// receiveFilesMaxMemory is the maximum amount of memory that receiveFiles may use at a time, in bytes.
const receiveFilesMaxMemory = 100*megabyte

// serveN serves n HTTP requests. It listens for requests on the given port and endpoint and handles them using
// the given serve function.
func serveN(port string, endpoint string, serve func(w http.ResponseWriter, r *http.Request), n int) {
	server := &http.Server{Addr: port}
	var requestsDone = 0
	http.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request){
		serve(w, r)
		requestsDone++
	})
	go server.ListenAndServe()
	ticker := time.NewTicker(serveNPollInterval)
	for range ticker.C{
		if requestsDone >= n {
			ctx, _ := context.WithTimeout(context.Background(), serveNShutdownTimeout)
			server.Shutdown(ctx)
			ticker.Stop()
			break
		}
	}
}

// ReceiveFiles receives files over HTTP. It listens for HTTP requests to the given port and endpoint, with form
// uploads containing a field with formFieldName. It saves the files uploaded to the directory at path, with
// filenames that are successive integers. It receives n files.
func ReceiveFiles(port string, endpoint string, formFieldName string, path string, n int) {
	var fileNumbers = make(chan int, n)
	for i := 0; i < n; i++ {
		fileNumbers <- i
	}
	handler := func(w http.ResponseWriter, r *http.Request) {
		// Adapted from https://astaxie.gitbooks.io/build-web-application-with-golang/en/04.5.html.
		r.ParseMultipartForm(receiveFilesMaxMemory)
		uploadedFile, _, err := r.FormFile(formFieldName)
		Check(err)
		defer uploadedFile.Close()

		file, err := os.OpenFile(path + "/" + strconv.Itoa(<- fileNumbers), os.O_WRONLY|os.O_CREATE, 0666)
		Check(err)
		defer file.Close()

		io.Copy(file, uploadedFile)
		return
	}
	serveN(port, endpoint, handler, n)
}

// SendFile sends a file over HTTP. It sends an HTTP request to the given address, port, and endpoint, with a
// form upload containing a field with formFieldName. It uploads the file at path.
func SendFile(address string, port string, endpoint string, formFieldName string, path string) {
	// Adapted from https://astaxie.gitbooks.io/build-web-application-with-golang/en/04.5.html.
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile(formFieldName, path)
	Check(err)
	fh, err := os.Open(path)
	Check(err)
	defer fh.Close()
	_, err = io.Copy(fileWriter, fh)
	Check(err)
	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()
	resp, err := http.Post("http://" + address + port + endpoint, contentType, bodyBuf)
	Check(err)
	defer resp.Body.Close()
	return
}