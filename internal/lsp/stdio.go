package lsp

import "io"

type stdio struct {
	io.ReadCloser
	io.WriteCloser
}

func (s stdio) Close() error {
	err1 := s.ReadCloser.Close()
	err2 := s.WriteCloser.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
