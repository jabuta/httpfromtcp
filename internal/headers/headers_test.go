package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.loopHelper(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.testGet("Host"))
	assert.Equal(t, 25, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.loopHelper(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid character header
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.loopHelper(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid 2 headers with existing headers
	headers = NewHeaders()
	// Add an existing header
	headers.Set("Existing", "value")

	data = []byte("Host: localhost:42069\r\nContent-Type: application/json\r\n\r\n")
	n, done, err = headers.loopHelper(data)
	require.NoError(t, err)
	require.NotNil(t, headers)

	// Check all headers exist
	assert.Equal(t, "value", headers.testGet("Existing"))
	assert.Equal(t, "localhost:42069", headers.testGet("Host"))
	assert.Equal(t, "application/json", headers.testGet("Content-Type"))
	assert.Equal(t, 57, n)
	assert.True(t, done)

	// Test: Valid done
	headers = NewHeaders()
	// Test with just the end marker
	data = []byte("\r\n")
	n, done, err = headers.loopHelper(data)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test with multiple headers ending properly
	headers = NewHeaders()
	data = []byte("Authorization: Bearer token123\r\nUser-Agent: Go-Client/1.0\r\n\r\n")
	n, done, err = headers.loopHelper(data)
	require.NoError(t, err)
	assert.Equal(t, "Bearer token123", headers.testGet("Authorization"))
	assert.Equal(t, "Go-Client/1.0", headers.testGet("User-Agent"))
	assert.Equal(t, 61, n)
	assert.True(t, done)

	// Test with multiple headers repeated headers
	headers = NewHeaders()
	data = []byte("Set-Person: lane-loves-go\r\nSet-Person: prime-loves-zig\r\nSet-Person: tj-loves-ocaml\r\n\r\n")
	n, done, err = headers.loopHelper(data)
	require.NoError(t, err)
	assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", headers.testGet("Set-Person"))
	assert.Equal(t, 86, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.loopHelper(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test another invalid case - space before colon
	headers = NewHeaders()
	data = []byte("Host : localhost:42069\r\n\r\n")
	n, done, err = headers.loopHelper(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test invalid case - no colon
	headers = NewHeaders()
	data = []byte("Invalid Header Line\r\n\r\n")
	n, done, err = headers.loopHelper(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func (h Headers) loopHelper(data []byte) (int, bool, error) {
	done := false
	n := 0
	var nloc int
	var err error
	for !done {
		nloc, done, err = h.Parse(data[n:])
		if err != nil {
			return n, done, err
		}
		n += nloc
	}
	return n, done, nil
}
func (h Headers) testGet(key string) string {
	value, _ := h.Get(key)
	return value
}
