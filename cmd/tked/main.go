package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"

	"tked/internal/app"
)

func main() {
	// Parse command line arguments
	flag.Parse()

	// Create the application instance
	application, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	// Load the settings
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get users home directory: %v", err)
	}

	err = application.LoadSettings(filepath.Join(homeDirectory, ".tked.toml"))
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to load settings: %v", err)
	}

	// Was a filename provided on the command line? If so, open it.
	if flag.NArg() > 0 {
		filename := flag.Arg(0)
		err = application.OpenFile(filename)
		if err != nil {
			log.Fatalf("Failed to open file: %v", err)
		}
	}

	// Create the screen (tcell)
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("Failed to create screen: %v", err)
	}

	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}

	quit := func() {
		// You have to catch panics in a defer, clean up, and
		// re-raise them - otherwise your application can
		// die without leaving any diagnostic trace.
		maybePanic := recover()
		screen.Fini()
		if maybePanic != nil {
			panic(maybePanic)
		}
	}
	defer quit()

	// Start the application event loop
	application.Run(screen)
}
