package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"net/url"
	"strings"
)

type requestStatus int

const (
	requestStatusInitialized requestStatus = iota
	RequestStatusParsingHeaders
	requestStatusDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	status      requestStatus
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"
const buffersize int = 8

func RequestFromReader(reader io.Reader) (*Request, error) {

	request := &Request{
		status: requestStatusInitialized,
	}
	buffer := make([]byte, buffersize)
	readToIndex := 0

	for request.status != requestStatusDone {
		if readToIndex >= len(buffer) {
			tempBuff := make([]byte, len(buffer)*2)
			copy(tempBuff, buffer)
			buffer = tempBuff
		}

		nRead, err := reader.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.status != requestStatusDone {
					return nil, fmt.Errorf("Incomplete Request")
				}
				break
			}
			return nil, fmt.Errorf("error reading request: %v", err)
		}
		readToIndex += nRead

		nParsed, err := request.parse(buffer[:readToIndex])
		if err != nil {
			return nil, fmt.Errorf("error parsing buffer: %v", err)
		}
		copy(buffer, buffer[nParsed:])
		readToIndex -= nParsed
	}
	fmt.Printf("HERE WITH REQUEST:\n%v\n", request)
	return request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.status != requestStatusDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.status {
	case requestStatusInitialized:
		requestLine, bytesRead, err := parseRequestLine(data)
		if err != nil {
			return 0, fmt.Errorf("Encountered an error parsing request line:\n %s", err)
		}
		if bytesRead == 0 {
			fmt.Printf("Need more data, so far %v\n", len(data))
			return 0, nil
		}
		fmt.Printf("read data, so far %v bytes read\n", bytesRead)
		r.RequestLine = *requestLine
		r.status = RequestStatusParsingHeaders
		return bytesRead, nil
	case RequestStatusParsingHeaders:
		if r.Headers == nil {
			r.Headers = headers.NewHeaders()
		}
		bytesRead, doneHeaders, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		fmt.Printf("i have parsed %v bytes for header, and done is %v\n", bytesRead, doneHeaders)
		if doneHeaders {
			r.status = requestStatusDone
		}
		return bytesRead, nil
	case requestStatusDone:
		return 0, fmt.Errorf("request already parsed")
	default:
		return 0, fmt.Errorf("unrecognized status")

	}
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, fmt.Errorf("couldnt get requestline from string: %s", err)
	}
	return requestLine, idx + len([]byte(crlf)), nil
}
func requestLineFromString(str string) (*RequestLine, error) {

	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed request line")
	}

	method := parts[0]
	if len(method) == 0 {
		return nil, fmt.Errorf("empty method: %s", method)
	}
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	requestTarget := parts[1]
	if _, err := url.ParseRequestURI(requestTarget); err != nil {
		return nil, err
	}

	httpVersion := parts[2]
	versionParts := strings.Split(httpVersion, "/")

	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start line: %s", httpVersion)
	}

	httpPart := versionParts[0]

	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unsupported HTTP verison: %s", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   version,
	}, nil
}
