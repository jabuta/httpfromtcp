package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

type Headers map[string]string

const crlf = "\r\n"

func NewHeaders() Headers {
	h := make(Headers)
	return h
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return len([]byte(crlf)), true, nil
	}
	idxSep := bytes.Index(data, []byte(":"))
	if idxSep == -1 {
		return 0, false, fmt.Errorf("no colon found")
	}
	key := string(data[:idxSep])
	key = strings.TrimLeft(key, " ")
	err = validateKeyWhitespace(key)
	if err != nil {
		return 0, false, err
	}
	key = strings.TrimSpace(key)
	value := string(data[idxSep+1 : idx])
	value = strings.TrimSpace(value)
	err = h.Set(key, value)
	if err != nil {
		return 0, false, err
	}
	return idx + len([]byte(crlf)), false, nil
}

func (h Headers) Set(key, value string) error {
	if !isValidKey(key) {
		return fmt.Errorf("invalid key characters")
	}
	key = strings.ToLower(key)
	if _, ok := h[key]; ok {
		h[key] = h[key] + ", " + value
		return nil
	}
	h[key] = value
	return nil
}

func (h Headers) Overwrite(key, value string) error {
	if !isValidKey(key) {
		return fmt.Errorf("invalid key characters")
	}
	key = strings.ToLower(key)
	h[key] = value
	return nil
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	value, ok := h[key]
	return value, ok
}

func validateKeyWhitespace(key string) error {
	for _, r := range key {
		if unicode.IsSpace(r) {
			return fmt.Errorf("malformed key with space")
		}
	}
	return nil
}
