package main

import (
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
		handleHttpBinProxy(w, req.RequestLine.RequestTarget)
		return
	}
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
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
		return
	case "/myproblem":
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
		return
	default:
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
}

func handleHttpBinProxy(w *response.Writer, path string) {

	path = strings.Trim(path, "httpbin/")

	binResp, err := http.Get(fmt.Sprintf("https://httpbin.org/%s", path))
	if err != nil {
		// if urlErr, ok := err.(*url.Error); ok {
		// 	if urlErr.Timeout() {
		// 		w.WriteStatusLine(response.HttpServerError)
		// 		return
		// 	}
		// }
		w.WriteStatusLine(response.HttpServerError)
		return
	}
	// Set headers
	h := headers.NewHeaders()
	for key, values := range binResp.Header {
		for _, value := range values {
			h.Set(key, value)
		}
	}

	// Get Status Code and send to client and respond with headers
	statusCode := response.StatusCode(binResp.StatusCode)
	w.WriteStatusLine(statusCode)
	w.WriteHeaders(h)

	b := make([]byte, 1024)

	for {
		n, err := binResp.Body.Read(b)
		fmt.Println("bytes read: ", n)
		if n > 0 {
			if _, writeErr := w.WriteBody(b[:n]); writeErr != nil {
				fmt.Printf("Write error: %s", writeErr)
			}
		}
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error: %s", err)
				return
			}
			break
		}
	}

}
