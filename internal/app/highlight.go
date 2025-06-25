package app

import (
	"github.com/gdamore/tcell/v2"
	"go.lsp.dev/protocol"

	"tked/internal/lsp"
	"tked/internal/tklog"
)

// syntaxToken represents a colored range within the buffer.
type syntaxToken struct {
	start int
	end   int
	typ   protocol.SemanticTokenTypes
}

var syntaxTokensProp PropKey

func registerSyntaxProperty() {
	if syntaxTokensProp == nil {
		syntaxTokensProp = RegisterBufferProperty()
	}
}

var defaultSemanticTokenTypes = []protocol.SemanticTokenTypes{
	protocol.SemanticTokenNamespace,
	protocol.SemanticTokenType,
	protocol.SemanticTokenClass,
	protocol.SemanticTokenEnum,
	protocol.SemanticTokenInterface,
	protocol.SemanticTokenStruct,
	protocol.SemanticTokenTypeParameter,
	protocol.SemanticTokenParameter,
	protocol.SemanticTokenVariable,
	protocol.SemanticTokenProperty,
	protocol.SemanticTokenEnumMember,
	protocol.SemanticTokenEvent,
	protocol.SemanticTokenFunction,
	protocol.SemanticTokenMethod,
	protocol.SemanticTokenMacro,
	protocol.SemanticTokenKeyword,
	protocol.SemanticTokenModifier,
	protocol.SemanticTokenComment,
	protocol.SemanticTokenString,
	protocol.SemanticTokenNumber,
	protocol.SemanticTokenRegexp,
	protocol.SemanticTokenOperator,
}

func decodeSemanticTokens(b Buffer, data []uint32) []syntaxToken {
	tokens := []syntaxToken{}
	line := 0
	char := 0
	for i := 0; i+4 < len(data); i += 5 {
		deltaLine := int(data[i])
		deltaStart := int(data[i+1])
		length := int(data[i+2])
		tokenType := int(data[i+3])
		if deltaLine == 0 {
			char += deltaStart
		} else {
			line += deltaLine
			char = deltaStart
		}
		startIdx := indexForPosition(b, line, char)
		endIdx := indexForPosition(b, line, char+length)
		if tokenType >= 0 && tokenType < len(defaultSemanticTokenTypes) {
			tokens = append(tokens, syntaxToken{start: startIdx, end: endIdx, typ: defaultSemanticTokenTypes[tokenType]})
		}
	}
	return tokens
}

var tokenColors = map[protocol.SemanticTokenTypes]tcell.Color{
	protocol.SemanticTokenKeyword:  tcell.ColorYellow,
	protocol.SemanticTokenString:   tcell.ColorGreen,
	protocol.SemanticTokenComment:  tcell.ColorGray,
	protocol.SemanticTokenNumber:   tcell.ColorLightCyan,
	protocol.SemanticTokenFunction: tcell.ColorFuchsia,
}

func colorForToken(t protocol.SemanticTokenTypes) tcell.Color {
	if c, ok := tokenColors[t]; ok {
		return c
	}
	return tcell.ColorWhite
}

func (b *buffer) updateSemanticTokens() {
	registerSyntaxProperty()
	lspClient := lsp.GetLSP(b.GetFilename())
	if lspClient == nil {
		return
	}
	resp, err := lspClient.SemanticTokens(b.GetFilename())
	if err != nil || resp == nil {
		if err != nil {
			tklog.Error("failed to get semantic tokens: %v", err)
		}
		return
	}
	tokens := decodeSemanticTokens(b, resp.Data)
	b.SetProperty(syntaxTokensProp, tokens)
}
