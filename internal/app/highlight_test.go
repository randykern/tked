package app

import (
	"testing"

	"go.lsp.dev/protocol"

	"tked/internal/rope"
)

func TestDecodeSemanticTokens(t *testing.T) {
	commands = make(map[string]Command)
	registerCommands()
	ResetApp()
	if _, err := NewApp(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	b := NewBuffer("", rope.NewRope("func main"))
	data := []uint32{
		0, 0, 4, uint32(15), 0, // keyword "func"
		0, 5, 4, uint32(12), 0, // function "main"
	}
	tokens := decodeSemanticTokens(b, data)
	if len(tokens) != 2 {
		t.Fatalf("expected 2 tokens got %d", len(tokens))
	}
	if tokens[0].start != 0 || tokens[0].end != 4 || tokens[0].typ != protocol.SemanticTokenKeyword {
		t.Fatalf("first token unexpected: %#v", tokens[0])
	}
	if tokens[1].start != 5 || tokens[1].end != 9 || tokens[1].typ != protocol.SemanticTokenFunction {
		t.Fatalf("second token unexpected: %#v", tokens[1])
	}
}
