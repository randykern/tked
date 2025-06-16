package main

import (
	"flag"
	"log"

	"github.com/gdamore/tcell/v2"

	"tked/internal/app"
)

func main() {
	flag.Parse()

	app := app.New()

	// Is there a file to open?
	if flag.NArg() > 0 {
		filename := flag.Arg(0)
		app.OpenFile(filename)
	}

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

	app.Run(screen)
}
