package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

const crlf = "\r\n"

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.Headers{}
	h.Set("Content-Length", fmt.Sprintf("%v", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	var headerWrite string
	for key, header := range headers {
		headerWrite += key + ": " + header + crlf
	}
	headerWrite += crlf
	_, err := w.Write([]byte(headerWrite))
	return err
}
