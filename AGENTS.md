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
- The max() function is a go built in function
- buffer.go implements the buffer structure, which is the model
- command.go provide the Comand interface and a RegisterCommand and GetCommand functions to interact with the command registry
- commands.go implement all of the editor commands, like CommandExit, CommandUndo, etc
- core.go implements the app structure, which has core application / controller logic
- keybinding.go implements configurable keybindings to bind key events to Command implementations
- settings.go implements editor configuration and settings
- statusbar.go implements the status bar for the overall application
- view.go implements the View interface and view structure, the view in the application architecture

## Testing and Formatting

- Unit tests live in `*_test.go` files next to the code under test. Write thorough unit tests for all changes.
- Run tests using `go test ./...` before committing.
- Format code with `gofmt -w` on any changed Go files.
