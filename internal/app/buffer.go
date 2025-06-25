package app

import (
	"fmt"
	"io"
	"path/filepath"

	"tked/internal/lsp"
	"tked/internal/rope"
)

type Buffer interface {
	// GetVersion returns the version of the buffer- which is incremented each time the
	//buffer is modified, including undo/redo. E.g. it is monotonically increasing.
	GetVersion() int32

	// Returns the filename of the buffer.
	GetFilename() string
	// SetFilename sets the filename of the buffer. This will also reset the title
	SetFilename(filename string)

	// GetTitle returns the title of the buffer.
	GetTitle() string
	// SetTitle sets the title of the buffer.
	SetTitle(title string)

	// Returns the contents of the buffer.
	Contents() rope.Rope

	// Returns true if the buffer has been modified since it was last saved.
	IsDirty() bool

	// IndexForRow returns the index for the first character of the requested
	// row, and the second return value is the actual row in case the row argument
	// is past the end of the buffer.
	IndexForRow(row int) (int, int)

	// TODO: Should we have a version that takes a rope?
	// Insert modifies the buffer contents by inserting the given text at the specified index.
	Insert(idx int, text string)

	// Delete modifies the buffer contents by removing the specified range.
	Delete(start, end int)

	// Undo reverts the last editing action. Returns true if an action was undone,
	// false if nothinig to undo.
	Undo() bool

	// Redo reapplies an undone editing action. Returns true if an action was
	// redone, false if nothing to redo.
	Redo() bool

	// Write writes the buffer contents to the writer. It returns the number of
	// bytes written and any error that occurred.
	Write(w io.Writer) (int64, error)

	// Close closes the buffer.
	Close()

	// GetProperty gets a custom property from the buffer.
	GetProperty(prop PropKey) any
	// SetProperty sets a custom property on the buffer- these are stored with
	// each buffer state, so undo/redo will restore the property values.
	SetProperty(prop PropKey, value any)

	// OnChange registers a callback function to be called when the buffer changes.
	OnChange(callback func(buffer Buffer, start, end int, context any), context any) ChangeRegistration
}

// PropKey is a unique identifier for a property.
type PropKey interface {
}

type ChangeRegistration interface {
	Remove()
}

type buffer struct {
	version         int32
	filename        string
	title           string
	contents        *bufferContents
	changeCallbacks []changeCallback
}

type bufferContents struct {
	rope             rope.Rope
	dirty            bool
	properties       []propValue
	subsequentState  *bufferContents
	previousContents *bufferContents
}

type propKey struct {
	id int
}

type propValue struct {
	key   propKey
	value any
}

type changeCallback struct {
	buffer   *buffer
	callback func(buffer Buffer, start, end int, context any)
	context  any
}

var untitledCounter int = 0

func (b *buffer) GetVersion() int32 {
	return b.version
}

func (b *buffer) GetFilename() string {
	return b.filename
}

func (b *buffer) SetFilename(filename string) {
	b.filename = filename
	b.title = filepath.Base(filename)
	if b.filename == "" {
		b.title = fmt.Sprintf("Untitled %d", untitledCounter)
		untitledCounter++
	}
}

func (b *buffer) GetTitle() string {
	return b.title
}

func (b *buffer) SetTitle(title string) {
	b.title = title
}

func (b *buffer) Contents() rope.Rope {
	return b.contents.rope
}

func (b *buffer) IsDirty() bool {
	return b.contents.dirty
}

func (b *buffer) IndexForRow(row int) (int, int) {
	if row < 0 {
		row = 0
	}

	// TODO: This is very slow. We should use a more efficient algorithm.
	idxRowStart := 0
	idx := 0
	currRow := 0
	for {
		if currRow == row {
			return idxRowStart, currRow
		}

		ch, ok := b.contents.rope.Index(idx)
		if !ok {
			return idxRowStart, currRow
		}
		if ch == '\n' {
			currRow++
			idxRowStart = idx + 1
		}
		idx++
	}
}

func (b *buffer) Insert(idx int, text string) {
	// Ensure idx is within bounds of the rope.
	idx = max(0, min(idx, b.contents.rope.Len()))

	// Create a new buffer contents with the new text.
	nc := b.newContents()
	nc.rope = nc.rope.Insert(idx, text)
	nc.dirty = true
	b.version++

	b.notifyChange(idx, idx+len(text))
}

func (b *buffer) Delete(start, end int) {
	// Ensure start and end are within bounds of the rope.
	start = max(0, min(start, b.contents.rope.Len()))
	end = max(0, min(end, b.contents.rope.Len()))

	if start > end {
		start, end = end, start
	}
	if start != end { // Create a new buffer contents with the deleted text.
		nc := b.newContents()
		nc.rope = nc.rope.Delete(start, end)
		nc.dirty = true
		b.version++

		b.notifyChange(start, end)
	}
}

func (b *buffer) Undo() bool {
	if b.contents.previousContents == nil {
		return false
	}

	b.contents = b.contents.previousContents
	b.version++
	b.notifyChange(0, b.contents.rope.Len())
	return true
}

func (b *buffer) Redo() bool {
	if b.contents.subsequentState == nil {
		return false
	}

	b.contents = b.contents.subsequentState
	b.version++
	b.notifyChange(0, b.contents.rope.Len())
	return true
}

func (b *buffer) Write(w io.Writer) (int64, error) {
	n, err := b.contents.rope.Write(w)
	if err == nil {
		b.contents.dirty = false
	}
	return n, err
}

func (b *buffer) Close() {
	lspClient := lsp.GetLSP(b.GetFilename())
	if lspClient != nil {
		lspClient.DidClose(b.GetFilename())
	}
}

func (b *buffer) GetProperty(prop PropKey) any {
	privatePropKey := prop.(propKey)
	for _, pv := range b.contents.properties {
		if pv.key.id == privatePropKey.id {
			return pv.value
		}
	}
	return nil
}

func (b *buffer) SetProperty(prop PropKey, value any) {
	privatePropKey := prop.(propKey)
	for i, pv := range b.contents.properties {
		if pv.key.id == privatePropKey.id {
			b.contents.properties[i].value = value
			return
		}
	}
	b.contents.properties = append(b.contents.properties, propValue{key: privatePropKey, value: value})
}

func (b *buffer) OnChange(callback func(buffer Buffer, start, end int, context any), context any) ChangeRegistration {
	cb := changeCallback{
		buffer:   b,
		callback: callback,
		context:  context,
	}

	b.changeCallbacks = append(b.changeCallbacks, cb)
	return &b.changeCallbacks[len(b.changeCallbacks)-1]
}

func (b *buffer) notifyChange(start, end int) {
	for _, cb := range b.changeCallbacks {
		cb.callback(b, start, end, cb.context)
	}

	lspClient := lsp.GetLSP(b.GetFilename())
	if lspClient != nil {
		lspClient.DidChangeFull(b.GetFilename(), b.GetVersion(), b.Contents().String())
		b.updateSemanticTokens()
	}
}

func (b *buffer) newContents() *bufferContents {
	nc := &bufferContents{
		rope:             b.contents.rope,
		dirty:            b.contents.dirty,
		properties:       append([]propValue{}, b.contents.properties...),
		subsequentState:  nil,
		previousContents: b.contents,
	}
	b.contents.subsequentState = nc
	b.contents = nc

	return nc
}

func (c *changeCallback) Remove() {
	for i := range c.buffer.changeCallbacks {
		if &c.buffer.changeCallbacks[i] == c {
			c.buffer.changeCallbacks = append(c.buffer.changeCallbacks[:i], c.buffer.changeCallbacks[i+1:]...)
			return
		}
	}
}

func NewBuffer(filename string, contents rope.Rope) Buffer {
	if contents == nil {
		contents = rope.NewRope("")
	}

	var b *buffer = &buffer{
		contents: &bufferContents{
			rope:             contents,
			dirty:            false,
			properties:       []propValue{},
			subsequentState:  nil,
			previousContents: nil,
		},
	}
	b.SetFilename(filename)
	registerSyntaxProperty()

	lspClient := lsp.GetLSP(filename)
	if lspClient != nil {
		lspClient.DidOpen(filename, b.version, b.contents.rope.String())
		b.updateSemanticTokens()
	}
	return b
}

func NewBufferFromReader(filename string, r io.Reader) (Buffer, error) {
	contents, err := rope.NewFromReader(r)
	if err != nil {
		return nil, err
	}
	return NewBuffer(filename, contents), nil
}

var propIdCounter int = 0

func RegisterBufferProperty() PropKey {
	propIdCounter++
	return propKey{id: propIdCounter}
}
