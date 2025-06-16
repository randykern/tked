# Coding Guidelines

This repo implements **tked**, a simple command line text editor written in Go.  The following guidelines summarize the repository layout and expectations for contributions.

## Repository Structure

- `cmd/tked/`: entry point for the application (`main.go`).
- `internal/app/`: contains editor logic and interfaces for MVC components.
- `internal/rope/`: implementation of a rope data structure with unit tests.
- `doc/architecture/decisions/`: ADRs describing project decisions.

## Architecture Notes

- The project is written in Go and follows a standard Go module layout.
- Overall architecture follows the Model–View–Controller pattern where:
  - the **Buffer** interface represents the model,
  - the **View** interface is the view,
  - and the **App** interface acts as the controller.
- The editor uses the [Tcell](https://github.com/gdamore/tcell) package for terminal interactions.
- Buffers store data using a rope data structure from `internal/rope`.

## Testing and Formatting

- Unit tests live in `*_test.go` files next to the code under test.
- Run tests using `go test ./...` before committing.
- Format code with `gofmt -w` on any changed Go files.

## Keybindings

The editor supports the following keybindings:

- `Esc` – exit the editor
- Arrow keys (`Up`/`Down`) – move the cursor
- `Pgup` / `Pgdn` – page up/down
- `Ctrl+Z` / `Ctrl+R` – undo/redo
- `Ctrl+S` – save the current buffer
- An asterisk (`*`) in the status line indicates unsaved changes.

