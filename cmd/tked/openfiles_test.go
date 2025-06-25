package main

import (
	"os"
	"testing"

	"github.com/gdamore/tcell/v2"
	"tked/internal/app"
)

// dummyApp implements app.App for testing openFiles.
type dummyApp struct {
	opened []string
}

func (d *dummyApp) OpenFile(name string) error  { d.opened = append(d.opened, name); return nil }
func (d *dummyApp) Run(tcell.Screen)            {}
func (d *dummyApp) Settings() app.Settings      { return app.NewSettings() }
func (d *dummyApp) LoadSettings(string) error   { return nil }
func (d *dummyApp) GetStatusBar() app.StatusBar { return nil }
func (d *dummyApp) GetCurrentView() app.View    { return nil }
func (d *dummyApp) SetCurrentView(app.View)     {}
func (d *dummyApp) Views() []app.View           { return nil }
func (d *dummyApp) CloseView(app.View) bool     { return true }

func TestOpenFiles(t *testing.T) {
	app := &dummyApp{}
	files := []string{"a.txt", "b.txt", "c.txt"}

	if err := openFiles(app, files); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(app.opened) != len(files) {
		t.Fatalf("expected %d files opened got %d", len(files), len(app.opened))
	}
	for i, f := range files {
		if app.opened[i] != f {
			t.Fatalf("file %d expected %s got %s", i, f, app.opened[i])
		}
	}
}

func TestOpenFilesFirstFileActive(t *testing.T) {
	app.ResetApp()
	application, err := app.NewApp()
	if err != nil {
		t.Fatalf("unexpected error creating app: %v", err)
	}

	f1, _ := os.CreateTemp("", "file1*.txt")
	defer os.Remove(f1.Name())
	f2, _ := os.CreateTemp("", "file2*.txt")
	defer os.Remove(f2.Name())

	if err := openFiles(application, []string{f1.Name(), f2.Name()}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := application.GetCurrentView().Buffer().GetFilename()
	if got != f1.Name() {
		t.Fatalf("expected first file active got %s", got)
	}
}
