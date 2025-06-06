package request

import (
	"io"
	"log"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("Error reading request: %v", err)
		return nil, err
	}
	requestString := string(request)
	requestLineParts := strings.Split(requestString, " ")
	requestLine := RequestLine{
		HttpVersion:   requestLineParts[2],
		RequestTarget: requestLineParts[1],
		Method:        requestLineParts[0],
	}
	return &Request{
		RequestLine: requestLine,
	}, nil
}
