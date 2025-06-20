# tked

`tked` is a lightweight command line text editor written in Go. It is intended
to be easy to build and simple to use while still supporting basic features such
as multiple buffers, undo/redo and customizable keybindings.

## Building

Clone the repository and run:

```bash
go build ./cmd/tked
```

This produces a `tked` executable in the current directory.

## Usage

Run the editor with an optional file name:

```bash
tked [file]
```

Configuration is read from `~/.tked.toml` if it exists.

### Default Keybindings

- `Ctrl+D`: Exit the editor
- `Ctrl+N`: New file
- `Ctrl+O`: Open a new file
- `Ctrl+S`: Save the current buffer
- `Ctrl+W`: Save the current buffer to a new filename
- `Ctrl+Q`: Close the current buffer, exiting if no buffers remain
- `Up`: Move up a line
- `Down`: Move down a line
- `PgUp`: Move up a page
- `PgDn`: Move down a page
- `Ctrl+Z`: Undo the last edit
- `Ctrl+R`: Redo the last undone edit
- `Alt+Left`: Move to previous view
- `Alt+Right`: Move to next view


## Running Tests

To execute all unit tests run:

```bash
go test ./...
```
