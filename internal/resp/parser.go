package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ReadCommand reads one command from r and returns it as a slice of strings.
// Supports both RESP array (*) and inline formats.
func ReadCommand(r *bufio.Reader) ([]string, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if b == '*' {
		return readArray(r)
	}
	return readInline(r, b)
}

func readArray(r *bufio.Reader) ([]string, error) {
	countLine, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	count, err := strconv.Atoi(strings.TrimSpace(string(countLine)))
	if err != nil {
		return nil, fmt.Errorf("invalid array length: %w", err)
	}

	parts := make([]string, 0, count)
	for range count {
		prefix, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if prefix != '$' {
			return nil, fmt.Errorf("expected '$', got %q", prefix)
		}
		lenLine, err := r.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		n, err := strconv.Atoi(strings.TrimSpace(string(lenLine)))
		if err != nil {
			return nil, err
		}
		buf := make([]byte, n+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		parts = append(parts, string(buf[:n]))
	}
	return parts, nil
}

func readInline(r *bufio.Reader, first byte) ([]string, error) {
	rest, err := r.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}
	line := strings.TrimRight(string(append([]byte{first}, rest...)), "\r\n")
	return strings.Fields(line), nil
}
