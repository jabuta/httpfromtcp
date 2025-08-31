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
	stateStatusLine    writerState = "Write Status Line"
	stateWriteHeaders  writerState = "Write Headers"
	stateWriteBody     writerState = "Write Body"
	stateWriteTrailers writerState = "Write Trailers"
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		state: stateStatusLine,
		w:     w,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != stateStatusLine {
		return fmt.Errorf("Called write functions out of order want %s got %s", stateStatusLine, w.state)
	}
	defer func() { w.state = stateWriteHeaders }()
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
	if w.state != stateWriteHeaders {
		return fmt.Errorf("Called write functions out of order want %s got %s", stateWriteHeaders, w.state)
	}
	defer func() { w.state = stateWriteBody }()
	var headerWrite string
	for key, header := range headers {
		headerWrite += key + ": " + header + crlf
	}
	headerWrite += crlf
	_, err := w.w.Write([]byte(headerWrite))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.state != stateWriteBody {
		return 0, fmt.Errorf("Called write functions out of order want %s got %s", stateWriteBody, w.state)
	}
	defer func() { w.state = stateWriteTrailers }()

	return w.w.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != stateWriteBody {
		return 0, fmt.Errorf("Called write functions out of order want %s got %s", stateWriteBody, w.state)
	}

	chunkSize := len(p)

	bytesWritten := 0
	n, err := fmt.Fprintf(w.w, "%x\r\n", chunkSize)
	if err != nil {
		return bytesWritten, err
	}
	bytesWritten += n

	n, err = w.w.Write(p)
	if err != nil {
		return bytesWritten, err
	}
	bytesWritten += n

	n, err = fmt.Fprint(w.w, "\r\n")
	if err != nil {
		return bytesWritten, err
	}
	bytesWritten += n
	return bytesWritten, nil
}

func (w *Writer) WriteChunkedBodyEnd() (int, error) {
	if w.state != stateWriteBody {
		return 0, fmt.Errorf("Called write functions out of order want %s got %s", stateWriteBody, w.state)
	}
	defer func() { w.state = stateWriteTrailers }()
	return w.w.Write([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(traiers headers.Headers) error {
	if w.state != stateWriteTrailers {
		return fmt.Errorf("Called write functions out of order want %s got %s", stateWriteTrailers, w.state)
	}
	var headerWrite string
	for key, header := range traiers {
		headerWrite += key + ": " + header + crlf
	}
	headerWrite += crlf
	_, err := w.w.Write([]byte(headerWrite))
	return err
}
