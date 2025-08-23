package request

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	endIndex = min(endIndex, len(cr.data))

	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	if n > cr.numBytesPerRead {
		n = cr.numBytesPerRead
		cr.pos -= n - cr.numBytesPerRead
	}
	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data: "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	}

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data: "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	}

	// Test: Good POST Request with path
	reader = &chunkReader{
		data: "POST /api/users HTTP/1.1\r\nHost: localhost:42069\r\nContent-Type: application/json\r\n\r\n{\"name\":\"test\"}",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "POST", r.RequestLine.Method)
		assert.Equal(t, "/api/users", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	}

	// Test: Invalid number of parts in request line
	reader = &chunkReader{
		data: "/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	}

	// Test: Invalid method (out of order) Request line
	reader = &chunkReader{
		data: "HTTP/1.1 GET /\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	}

	// Test: Invalid version in Request line
	reader = &chunkReader{
		data: "GET / HTTP/2.0\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	}

	// Test: Empty request line
	reader = &chunkReader{
		data: "\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	}

	// Test: Request line with too many spaces
	reader = &chunkReader{
		data: "GET  /  HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	}

	// Test: Request line with extra parts
	reader = &chunkReader{
		data: "GET / HTTP/1.1 EXTRA\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	}

	// Test: Lowercase method
	reader = &chunkReader{
		data: "get / HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	}

	// Test: Invalid HTTP version format
	reader = &chunkReader{
		data: "GET / HTTP1.1\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	}

	// Test: Invalid HTTP version number
	reader = &chunkReader{
		data: "GET / HTTP/1.0\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	}

	// Test: Missing request target
	reader = &chunkReader{
		data: "GET HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	}

	// Test: Invalid request target (not starting with /)
	reader = &chunkReader{
		data: "GET invalid HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	}

	// Test: Request target with query parameters
	reader = &chunkReader{
		data: "GET /search?q=test&limit=10 HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/search?q=test&limit=10", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	}

	// Test: Request target with fragment (should still work)
	reader = &chunkReader{
		data: "GET /page#section HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/page#section", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	}

	// Test: Different HTTP methods
	methods := []string{"PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range methods {
		requestString := method + " /test HTTP/1.1\r\nHost: localhost:42069\r\n\r\n"
		reader = &chunkReader{
			data: requestString,
		}
		for i := 1; i <= len(reader.data); i++ {
			reader.numBytesPerRead = i
			reader.pos = 0
			r, err := RequestFromReader(reader)
			require.NoError(t, err)
			require.NotNil(t, r)
			assert.Equal(t, method, r.RequestLine.Method)
			assert.Equal(t, "/test", r.RequestLine.RequestTarget)
			assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
		}
	}

	// Test: Method with numbers (should fail based on regex)
	reader = &chunkReader{
		data: "GET2 / HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	}

	// Test: Method with spaces (should fail)
	reader = &chunkReader{
		data: "GET POST / HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	}

	// Test: Empty method
	reader = &chunkReader{
		data: " / HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		_, err := RequestFromReader(reader)
		require.Error(t, err)
	}

	// Test: Very long request target
	longPath := "/very/long/path" + strings.Repeat("/segment", 100)
	reader = &chunkReader{
		data: "GET " + longPath + " HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, longPath, r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	}

	// Test: Request target with special characters
	reader = &chunkReader{
		data: "GET /path%20with%20spaces HTTP/1.1\r\nHost: localhost:42069\r\n\r\n",
	}
	for i := 1; i <= len(reader.data); i++ {
		reader.numBytesPerRead = i
		reader.pos = 0
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/path%20with%20spaces", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
	}

}

func TestHeadersParse(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Test: Malformed Header
	reader = &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)
}

func TestBodyParse(t *testing.T) {
	// Test: Standard Body
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	// Test: Body shorter than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Empty Body, 0 reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Nil(t, r.Body)

	// Test: Empty Body, no reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Nil(t, r.Body)

	// Test: No Content-Length but Body Exists
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n" +
			"Ignored content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Nil(t, r.Body)
}
