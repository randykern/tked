package lsp

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/zap"

	"tked/internal/tklog"
)

type LSPClient interface {
	DidChangeFull(filename string, version int32, contents string)
	DidClose(filename string)
	DidOpen(filename string, version int32, contents string)

	// TODO: Cleanup server capabilities
	ServerTextDocumentSyncOptions() protocol.TextDocumentSyncOptions
}

type lspClient struct {
	name                          string
	cancel                        context.CancelFunc
	conn                          jsonrpc2.Conn
	server                        protocol.Server
	cmd                           *exec.Cmd
	serverTextDocumentSyncOptions protocol.TextDocumentSyncOptions
}

func (c *lspClient) DidChangeFull(filename string, version int32, contents string) {
	if c.serverTextDocumentSyncOptions.Change != protocol.TextDocumentSyncKindNone {
		err := c.server.DidChange(context.TODO(), &protocol.DidChangeTextDocumentParams{
			TextDocument: protocol.VersionedTextDocumentIdentifier{
				TextDocumentIdentifier: protocol.TextDocumentIdentifier{
					URI: protocol.DocumentURI(filename),
				},
				Version: version,
			},
			ContentChanges: []protocol.TextDocumentContentChangeEvent{
				{
					Text: contents,
				},
			},
		})
		if err != nil {
			tklog.Error("LSP error on DidChange %s: %v", filename, err)
		} else {
			tklog.Info("LSP did change full: %s(%s)", c.name, filename)
		}
	} else {
		tklog.Info("LSP did change full not supported", filename)
	}
}

func (c *lspClient) DidClose(filename string) {
	if c.ServerTextDocumentSyncOptions().OpenClose {
		err := c.server.DidClose(context.TODO(), &protocol.DidCloseTextDocumentParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: protocol.DocumentURI(filename),
			},
		})
		if err != nil {
			tklog.Error("LSP error on DidClose %s: %v", filename, err)
		} else {
			tklog.Info("LSP did close: %s(%s)", c.name, filename)
		}
	} else {
		tklog.Info("LSP did close not supported", filename)
	}
}

func (c *lspClient) DidOpen(filename string, version int32, contents string) {
	if c.serverTextDocumentSyncOptions.OpenClose {
		err := c.server.DidOpen(context.TODO(), &protocol.DidOpenTextDocumentParams{
			TextDocument: protocol.TextDocumentItem{
				URI:        protocol.DocumentURI(filename),
				LanguageID: protocol.LanguageIdentifier(strings.TrimPrefix(filepath.Ext(filename), ".")),
				Version:    version,
				Text:       contents,
			}})
		if err != nil {
			tklog.Error("LSP error on DidOpen %s: %v", filename, err)
		} else {
			tklog.Info("LSP did open: %s(%s)", c.name, filename)
		}
	} else {
		tklog.Info("LSP did open not supported", filename)
	}
}

func (c *lspClient) shutdown() {
	// TODO: This block makes Shutdown hang
	/*
		if c.cancel != nil {
			c.cancel()
		}
	*/

	if c.server != nil {
		c.server.Shutdown(context.TODO())
		c.server.Exit(context.TODO())
	}
	if c.conn != nil {
		c.conn.Close()
	}

	tklog.Info("LSP shutdown: %s", c.name)
}

func (c *lspClient) ServerTextDocumentSyncOptions() protocol.TextDocumentSyncOptions {
	return c.serverTextDocumentSyncOptions
}

func (lspClient) Progress(context.Context, *protocol.ProgressParams) error { return nil }
func (lspClient) WorkDoneProgressCreate(context.Context, *protocol.WorkDoneProgressCreateParams) error {
	return nil
}
func (lspClient) LogMessage(context.Context, *protocol.LogMessageParams) error { return nil }
func (lspClient) PublishDiagnostics(context.Context, *protocol.PublishDiagnosticsParams) error {
	return nil
}
func (lspClient) ShowMessage(context.Context, *protocol.ShowMessageParams) error { return nil }
func (lspClient) ShowMessageRequest(context.Context, *protocol.ShowMessageRequestParams) (*protocol.MessageActionItem, error) {
	return nil, nil
}
func (lspClient) Telemetry(context.Context, interface{}) error                           { return nil }
func (lspClient) RegisterCapability(context.Context, *protocol.RegistrationParams) error { return nil }
func (lspClient) UnregisterCapability(context.Context, *protocol.UnregistrationParams) error {
	return nil
}
func (lspClient) ApplyEdit(context.Context, *protocol.ApplyWorkspaceEditParams) (bool, error) {
	return false, nil
}
func (lspClient) Configuration(context.Context, *protocol.ConfigurationParams) ([]interface{}, error) {
	return nil, nil
}
func (lspClient) WorkspaceFolders(context.Context) ([]protocol.WorkspaceFolder, error) {
	return nil, nil
}

func startLSPClient(command string) (*lspClient, error) {
	workspaceFolder, err := os.Getwd()
	if err != nil {
		tklog.Error("startLSPClient %s: %v", command, err)
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, command)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		tklog.Error("startLSPClient %s: %v", command, err)
		return nil, err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		tklog.Error("startLSPClient %s: %v", command, err)
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		cancel()
		tklog.Error("startLSPClient %s: %v", command, err)
		return nil, err
	}

	client := &lspClient{
		name:   command,
		cancel: cancel,
		conn:   nil,
		server: nil,
		cmd:    cmd,
		serverTextDocumentSyncOptions: protocol.TextDocumentSyncOptions{
			OpenClose: false,
			Change:    protocol.TextDocumentSyncKindNone,
		},
	}

	rw := stdio{ReadCloser: stdout, WriteCloser: stdin}
	stream := jsonrpc2.NewStream(rw)
	ctx, conn, server := protocol.NewClient(ctx, client, stream, zap.NewNop())
	client.conn = conn
	client.server = server

	initializeResponse, err := server.Initialize(ctx, &protocol.InitializeParams{
		ProcessID: int32(os.Getpid()),
		ClientInfo: &protocol.ClientInfo{
			Name:    "tked",
			Version: "0.1.0", // TODO: Get the version from the build info
		},
		Capabilities: protocol.ClientCapabilities{
			Workspace: &protocol.WorkspaceClientCapabilities{
				WorkspaceFolders: true,
			},
			TextDocument: &protocol.TextDocumentClientCapabilities{
				// TODO: Many capabilities here
			},
			Window: &protocol.WindowClientCapabilities{
				// TODO: workDoneProgress, showMessage and showDocument
			},
			General: &protocol.GeneralClientCapabilities{},
		},
		WorkspaceFolders: []protocol.WorkspaceFolder{
			{
				URI:  workspaceFolder,
				Name: filepath.Base(workspaceFolder),
			},
		},
	})
	if err != nil {
		cancel()
		cmd.Process.Kill()
		tklog.Error("startLSPClient %s: %v", command, err)
		return nil, err
	}

	client.serverTextDocumentSyncOptions = parseTextDocumentSyncOptions(initializeResponse.Capabilities.TextDocumentSync)

	err = server.Initialized(ctx, &protocol.InitializedParams{})
	if err != nil {
		tklog.Error("startLSPClient %s: %v", command, err)
		return nil, err
	}

	tklog.Info("LSP initialized: %s", command)
	return client, nil
}

func parseTextDocumentSyncOptions(textDocumentSync interface{}) protocol.TextDocumentSyncOptions {
	textDocumentSyncMap := textDocumentSync.(map[string]interface{})
	var textDocumentSyncOptions protocol.TextDocumentSyncOptions
	textDocumentSyncOptions.OpenClose = false
	textDocumentSyncOptions.Change = protocol.TextDocumentSyncKindNone

	for key, value := range textDocumentSyncMap {
		switch key {
		case "openClose":
			switch value {
			case true:
				textDocumentSyncOptions.OpenClose = true
			}
		case "change":
			switch value {
			case 1.0:
				textDocumentSyncOptions.Change = protocol.TextDocumentSyncKindFull
			case 2.0:
				textDocumentSyncOptions.Change = protocol.TextDocumentSyncKindIncremental
			}
		}
	}

	return textDocumentSyncOptions
}

var activeLSPs map[string]LSPClient

func GetLSP(filename string) LSPClient {
	// TODO: Concurrency?
	if activeLSPs == nil {
		activeLSPs = map[string]LSPClient{}
	}

	if filename == "" {
		return nil
	}

	// We use the file extension to determine the LSP client to use
	ext := filepath.Ext(filename)
	client, ok := activeLSPs[ext]
	if !ok {
		// cache the nil value so we don't keep trying to create an LSP for this filetype
		activeLSPs[ext] = nil

		// TODO: Add support for other languages, and LSP mapping in the editor settings
		var lspServerCommand string
		switch ext {
		case ".go":
			lspServerCommand = "gopls"
		}

		if lspServerCommand == "" {
			tklog.Warn("no LSP server command found for %s", ext)
			return nil
		}

		var err error
		client, err = startLSPClient(lspServerCommand)
		if err != nil {
			tklog.Error("failed to create LSP client for %s: %v", ext, err)
		} else {
			activeLSPs[ext] = client
		}
	}

	return client
}

func ShutdownAll() {
	if activeLSPs == nil {
		return
	}

	for _, lsp := range activeLSPs {
		if lsp != nil {
			lsp.(*lspClient).shutdown()
		}
	}
}
