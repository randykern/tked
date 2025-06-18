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

- `Esc`: Exit the editor
- `Up`: Move up a line
- `Down`: Move down a line
- `PgUp`: Move up a page
- `PgDn`: Move down a page
- `Ctrl+Z`: Undo the last edit
- `Ctrl+R`: Redo the last undone edit
- `Ctrl+S`: Save the current buffer
- `Ctrl+O`: Open a new file


## Running Tests

To execute all unit tests run:

```bash
go test ./...
```
