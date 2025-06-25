package lsp

import (
	"go.lsp.dev/protocol"
	"testing"
)

func TestParseTextDocumentSyncOptions(t *testing.T) {
	input := map[string]interface{}{
		"openClose": true,
		"change":    1.0,
	}
	opts := parseTextDocumentSyncOptions(input)
	if !opts.OpenClose {
		t.Errorf("expected OpenClose true")
	}
	if opts.Change != protocol.TextDocumentSyncKindFull {
		t.Errorf("expected Change full got %v", opts.Change)
	}

	input2 := map[string]interface{}{
		"change": 2.0,
	}
	opts2 := parseTextDocumentSyncOptions(input2)
	if opts2.OpenClose {
		t.Errorf("expected OpenClose false")
	}
	if opts2.Change != protocol.TextDocumentSyncKindIncremental {
		t.Errorf("expected Change incremental got %v", opts2.Change)
	}
}

func TestGetLSPUnknownExtension(t *testing.T) {
	oldActive := activeLSPs
	oldStart := startLSPClientFunc
	activeLSPs = nil
	called := false
	startLSPClientFunc = func(cmd string) (*lspClient, error) {
		called = true
		return &lspClient{name: cmd}, nil
	}
	defer func() {
		activeLSPs = oldActive
		startLSPClientFunc = oldStart
	}()

	client := GetLSP("file.txt")
	if client != nil {
		t.Fatalf("expected nil client for txt got %#v", client)
	}
	if called {
		t.Fatalf("startLSPClientFunc should not be called")
	}
	if v, ok := activeLSPs[".txt"]; !ok || v != nil {
		t.Fatalf("expected cached nil for .txt")
	}
}

func TestGetLSPGoCaching(t *testing.T) {
	oldActive := activeLSPs
	oldStart := startLSPClientFunc
	activeLSPs = nil
	count := 0
	created := &lspClient{name: "stub"}
	startLSPClientFunc = func(cmd string) (*lspClient, error) {
		count++
		return created, nil
	}
	defer func() {
		activeLSPs = oldActive
		startLSPClientFunc = oldStart
	}()

	c1 := GetLSP("a.go")
	if c1 == nil {
		t.Fatalf("expected client")
	}
	c2 := GetLSP("b.go")
	if c1 != c2 {
		t.Fatalf("expected same cached client")
	}
	if count != 1 {
		t.Fatalf("expected startLSPClientFunc once got %d", count)
	}
}

func TestShutdownAllNilMap(t *testing.T) {
	oldActive := activeLSPs
	activeLSPs = nil
	ShutdownAll()
	activeLSPs = oldActive
}
