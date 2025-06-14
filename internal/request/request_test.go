package request

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	r, err := RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	r, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good POST Request with path
	r, err = RequestFromReader(strings.NewReader("POST /api/users HTTP/1.1\r\nHost: localhost:42069\r\nContent-Type: application/json\r\n\r\n{\"name\":\"test\"}"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "/api/users", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Invalid number of parts in request line
	_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid method (out of order) Request line
	_, err = RequestFromReader(strings.NewReader("HTTP/1.1 GET /\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid version in Request line
	_, err = RequestFromReader(strings.NewReader("GET / HTTP/2.0\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)

	// Test: Empty request line
	_, err = RequestFromReader(strings.NewReader("\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)

	// Test: Request line with too many spaces
	_, err = RequestFromReader(strings.NewReader("GET  /  HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)

	// Test: Request line with extra parts
	_, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1 EXTRA\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)

	// Test: Lowercase method
	_, err = RequestFromReader(strings.NewReader("get / HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid HTTP version format
	_, err = RequestFromReader(strings.NewReader("GET / HTTP1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid HTTP version number
	_, err = RequestFromReader(strings.NewReader("GET / HTTP/1.0\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)

	// Test: Missing request target
	_, err = RequestFromReader(strings.NewReader("GET HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid request target (not starting with /)
	_, err = RequestFromReader(strings.NewReader("GET invalid HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)

	// Test: Request target with query parameters
	r, err = RequestFromReader(strings.NewReader("GET /search?q=test&limit=10 HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/search?q=test&limit=10", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Request target with fragment (should still work)
	r, err = RequestFromReader(strings.NewReader("GET /page#section HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/page#section", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Different HTTP methods
	methods := []string{"PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range methods {
		requestString := method + " /test HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"
		r, err = RequestFromReader(strings.NewReader(requestString))
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, method, r.RequestLine.Method)
		assert.Equal(t, "/test", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	}

	// Test: Method with numbers (should fail based on regex)
	_, err = RequestFromReader(strings.NewReader("GET2 / HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)

	// Test: Method with spaces (should fail)
	_, err = RequestFromReader(strings.NewReader("GET POST / HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)

	// Test: Empty method
	_, err = RequestFromReader(strings.NewReader(" / HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.Error(t, err)

	// Test: Very long request target
	longPath := "/very/long/path" + strings.Repeat("/segment", 100)
	r, err = RequestFromReader(strings.NewReader("GET " + longPath + " HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, longPath, r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Request target with special characters
	r, err = RequestFromReader(strings.NewReader("GET /path%20with%20spaces HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/path%20with%20spaces", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Request with only LF instead of CRLF
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1\nHost: localhost:42069\n\n"))
	// This might pass depending on implementation, but testing edge case
	// The current implementation splits on \r\n so this might not work as expected
}
