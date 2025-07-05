package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
)

type status int

const (
	initialized status = iota
	done
)

const buffersize int = 8

type Request struct {
	RequestLine RequestLine
	Status      status
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"

func RequestFromReader(reader io.Reader) (*Request, error) {

	request := newRequest()
	buffer := make([]byte, buffersize)
	readToIndex := 0

	for request.Status == initialized {
		if readToIndex >= len(buffer) {
			tempBuff := make([]byte, len(buffer)*2)
			copy(tempBuff, buffer)
			buffer = tempBuff
		}
		nRead, err := reader.Read(buffer[readToIndex:])
		readToIndex += nRead
		if errors.Is(err, io.EOF) {
			request.Status = done
		}

		if readToIndex > 0 {
			nParsed, err := request.parse(buffer[:readToIndex])
			if err != nil {
				return nil, fmt.Errorf("error parsing buffer: %v", err)
			}
			if nParsed > 0 {
				tempBuff := make([]byte, readToIndex-nParsed)
				copy(tempBuff, buffer[nParsed:readToIndex])
				buffer = tempBuff
				readToIndex -= nParsed
			}
		}

		if err != nil {
			return nil, fmt.Errorf("error reading request: %v", err)
		}
	}
	return request, nil
}

// creates a new empty *Request with an initialized Status
func newRequest() *Request {
	return &Request{
		Status: initialized,
	}
}

func (r *Request) parse(data []byte) (int, error) {
	if r.Status == done {
		return 0, fmt.Errorf("request already parsed")
	}
	if r.Status != initialized {
		return 0, fmt.Errorf("unrecognized status")
	}
	requestLine, bytesRead, err := parseRequestLine(data)
	if err != nil {
		return 0, fmt.Errorf("Encountered an error parsing request line:\n %s", err)
	} else if bytesRead == 0 {
		return 0, nil
	}
	r.RequestLine = *requestLine
	r.Status = done
	return bytesRead, err
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	bytesRead := idx + len([]byte(crlf))
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, bytesRead, fmt.Errorf("couldnt get requestline from string: %s", err)
	}
	return requestLine, bytesRead, nil
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

	if len(versionParts) > 2 {
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
