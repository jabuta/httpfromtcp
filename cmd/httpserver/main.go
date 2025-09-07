package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const port = 42069

func main() {
	server, err := server.Serve(port, testHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func testHandler(w *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		httpbinProxyHandler(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handle400(w, nil)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handle500(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/video" {
		handleVideo(w, req)
		return
	}
	handle200(w, req)
}

func handle200(w *response.Writer, _ *request.Request) {
	res := []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`)
	w.WriteStatusLine(response.HttpOK)
	headers := response.GetDefaultHeaders(len(res))
	headers.Overwrite("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody(res)
}

func handle400(w *response.Writer, _ *request.Request) {
	res := []byte(`<html>
 <head>
   <title>400 Bad Request</title>
 </head>
 <body>
   <h1>Bad Request</h1>
   <p>Your request honestly kinda sucked.</p>
 </body>
</html>`)
	w.WriteStatusLine(response.HttpNotFoud)
	headers := response.GetDefaultHeaders(len(res))
	headers.Overwrite("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody(res)
}

func handle500(w *response.Writer, _ *request.Request) {
	res := []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`)
	w.WriteStatusLine(response.HttpServerError)
	headers := response.GetDefaultHeaders(len(res))
	headers.Overwrite("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody(res)
}

func httpbinProxyHandler(w *response.Writer, req *request.Request) {

	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	fmt.Println(req.RequestLine.RequestTarget)

	binResp, err := http.Get(fmt.Sprintf("https://httpbin.org%s", path))
	fmt.Printf("https://httpbin.org/%s\n", path)
	if err != nil {
		handle500(w, req)
		return
	}
	defer binResp.Body.Close()

	w.WriteStatusLine(http.StatusOK)

	// Set headers
	h := response.GetDefaultHeaders(0)
	h.Delete("Content-Length")
	h.Overwrite("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")
	w.WriteHeaders(h)

	b := make([]byte, 1024)

	lenResp := 0
	hashResp := sha256.New()
	for {
		n, err := binResp.Body.Read(b)
		fmt.Println("bytes read: ", n)
		if n > 0 {
			if _, writeErr := w.WriteChunkedBody(b[:n]); writeErr != nil {
				fmt.Printf("Write error: %s\n", writeErr)
			}
			lenResp += n
			hashResp.Write(b[:n])
		}
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error: %s\n", err)
				break
			}
			break
		}
	}
	_, err = w.WriteChunkedBodyEnd()
	if err != nil {
		fmt.Printf("Error writing chunked body end: %v", err)
	}

	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", hashResp.Sum(nil)))
	trailers.Set("X-Content-Length", strconv.Itoa(lenResp))
	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Printf("Error writing Trailers: %v", err)
	}
}

func handleVideo(w *response.Writer, req *request.Request) {
	videoFile, err := os.Open("assets/vim.mp4")
	if err != nil {
		fmt.Println(err)
		handle500(w, req)
		return
	}

	w.WriteStatusLine(http.StatusOK)

	// Set headers
	h := response.GetDefaultHeaders(0)
	h.Delete("Content-Length")
	h.Overwrite("Content-Type", "video/mp4")
	h.Overwrite("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")
	w.WriteHeaders(h)

	b := make([]byte, 1024)
	lenResp := 0
	hashResp := sha256.New()

	for {
		n, err := videoFile.Read(b)
		fmt.Println("bytes sent: ", n)
		if n > 0 {
			if _, writeErr := w.WriteChunkedBody(b[:n]); writeErr != nil {
				fmt.Printf("Write error: %s\n", writeErr)
			}
			lenResp += n
			hashResp.Write(b[:n])
		}
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error: %s\n", err)
				break
			}
			break
		}
	}
	_, err = w.WriteChunkedBodyEnd()
	if err != nil {
		fmt.Printf("Error writing chunked body end: %v", err)
	}

	trailers := headers.NewHeaders()
	trailers.Set("X-Content-SHA256", fmt.Sprintf("%x", hashResp.Sum(nil)))
	trailers.Set("X-Content-Length", strconv.Itoa(lenResp))
	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Printf("Error writing Trailers: %v", err)
	}
}
