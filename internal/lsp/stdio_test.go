package lsp

import (
	"errors"
	"io"
	"testing"
)

type mockReadCloser struct {
	closed bool
	err    error
}

func (m *mockReadCloser) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (m *mockReadCloser) Close() error {
	m.closed = true
	return m.err
}

type mockWriteCloser struct {
	closed bool
	err    error
}

func (m *mockWriteCloser) Write(p []byte) (int, error) {
	return len(p), nil
}

func (m *mockWriteCloser) Close() error {
	m.closed = true
	return m.err
}

func TestStdioClose(t *testing.T) {
	readErr := errors.New("read close error")
	writeErr := errors.New("write close error")
	tests := []struct {
		name     string
		readErr  error
		writeErr error
		expect   error
	}{
		{name: "no errors"},
		{name: "read error", readErr: readErr, expect: readErr},
		{name: "write error", writeErr: writeErr, expect: writeErr},
		{name: "both errors", readErr: readErr, writeErr: writeErr, expect: readErr},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			readCloser := &mockReadCloser{err: tc.readErr}
			writeCloser := &mockWriteCloser{err: tc.writeErr}
			s := stdio{ReadCloser: readCloser, WriteCloser: writeCloser}
			err := s.Close()
			if !errors.Is(err, tc.expect) {
				t.Fatalf("expected error %v, got %v", tc.expect, err)
			}
			if !readCloser.closed || !writeCloser.closed {
				t.Fatalf("expected both closers to be closed, read=%v write=%v", readCloser.closed, writeCloser.closed)
			}
		})
	}
}
