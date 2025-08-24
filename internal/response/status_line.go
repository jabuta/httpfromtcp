package response

import (
	"fmt"
	"io"
)

type StatusCode int

const (
	HttpOK          StatusCode = 200
	HttpNotFoud     StatusCode = 400
	HttpServerError StatusCode = 500
)

var statusCodeMap = map[StatusCode]string{
	HttpOK:          "OK",
	HttpNotFoud:     "Bad Request",
	HttpServerError: "Internal Server Error",
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine, _ := statusCodeMap[statusCode]
	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s", statusCode, statusLine+crlf)
	return err
}
