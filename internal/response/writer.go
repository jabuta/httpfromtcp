package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type Writer struct {
	state writerState
	w     io.Writer
}

type writerState string

const (
	uninit       writerState = "Not Started"
	statusLine   writerState = "Status Line Done"
	writeHeaders writerState = "Write Headers Done"
	writeBody    writerState = "Write Body Done"
	done         writerState = "writer done"
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		state: uninit,
		w:     w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != uninit {
		return fmt.Errorf("Called write functions out of order want %s got %s", uninit, w.state)
	}
	w.state = statusLine
	var statusCodeMap = map[StatusCode]string{
		HttpOK:          "OK",
		HttpNotFoud:     "Bad Request",
		HttpServerError: "Internal Server Error",
	}
	statusInfo, _ := statusCodeMap[statusCode]
	_, err := fmt.Fprintf(w.w, "HTTP/1.1 %d %s", statusCode, statusInfo+crlf)
	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != statusLine {
		return fmt.Errorf("Called write functions out of order want %s got %s", statusLine, w.state)
	}
	w.state = writeHeaders
	var headerWrite string
	for key, header := range headers {
		headerWrite += key + ": " + header + crlf
	}
	headerWrite += crlf
	_, err := w.w.Write([]byte(headerWrite))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != writeHeaders {
		return 0, fmt.Errorf("Called write functions out of order want %s got %s", writeHeaders, w.state)
	}
	w.state = done
	return w.w.Write(p)
}
