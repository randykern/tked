package app

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
	"go.uber.org/zap"
)

type lspClient struct {
	cancel context.CancelFunc
	conn   jsonrpc2.Conn
	server protocol.Server
	cmd    *exec.Cmd
}

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

type noopClient struct{}

func (noopClient) Progress(context.Context, *protocol.ProgressParams) error { return nil }
func (noopClient) WorkDoneProgressCreate(context.Context, *protocol.WorkDoneProgressCreateParams) error {
	return nil
}
func (noopClient) LogMessage(context.Context, *protocol.LogMessageParams) error { return nil }
func (noopClient) PublishDiagnostics(context.Context, *protocol.PublishDiagnosticsParams) error {
	return nil
}
func (noopClient) ShowMessage(context.Context, *protocol.ShowMessageParams) error { return nil }
func (noopClient) ShowMessageRequest(context.Context, *protocol.ShowMessageRequestParams) (*protocol.MessageActionItem, error) {
	return nil, nil
}
func (noopClient) Telemetry(context.Context, interface{}) error                           { return nil }
func (noopClient) RegisterCapability(context.Context, *protocol.RegistrationParams) error { return nil }
func (noopClient) UnregisterCapability(context.Context, *protocol.UnregistrationParams) error {
	return nil
}
func (noopClient) ApplyEdit(context.Context, *protocol.ApplyWorkspaceEditParams) (bool, error) {
	return false, nil
}
func (noopClient) Configuration(context.Context, *protocol.ConfigurationParams) ([]interface{}, error) {
	return nil, nil
}
func (noopClient) WorkspaceFolders(context.Context) ([]protocol.WorkspaceFolder, error) {
	return nil, nil
}

func newLSPClient(command string, filename string) (*lspClient, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, command)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}

	rw := stdio{ReadCloser: stdout, WriteCloser: stdin}
	stream := jsonrpc2.NewStream(rw)
	ctx, conn, server := protocol.NewClient(ctx, noopClient{}, stream, zap.NewNop())

	root := uri.File(filepath.Dir(filename))
	_, err = server.Initialize(ctx, &protocol.InitializeParams{
		ProcessID:    int32(os.Getpid()),
		RootURI:      protocol.DocumentURI(root),
		Capabilities: protocol.ClientCapabilities{},
	})
	if err != nil {
		cancel()
		cmd.Process.Kill()
		return nil, err
	}
	server.Initialized(ctx, &protocol.InitializedParams{})
	/*
		server.DidOpen(ctx, &protocol.DidOpenTextDocumentParams{
			TextDocument: protocol.TextDocumentItem{
				URI:        protocol.DocumentURI(uri.File(filename)),
				LanguageID: protocol.LanguageIdentifier(strings.TrimPrefix(filepath.Ext(filename), ".")),
				Version:    0,
				Text:       v.Buffer().Contents().String(),
			},
		})
	*/

	return &lspClient{
		cancel: cancel,
		conn:   conn,
		server: server,
		cmd:    cmd,
	}, nil
}

func (c *lspClient) Close() {
	if c == nil {
		return
	}
	if c.server != nil {
		c.server.Shutdown(context.Background())
	}
	if c.conn != nil {
		c.conn.Close()
	}
	if c.cancel != nil {
		c.cancel()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		c.cmd.Process.Kill()
	}
}
