package app

import (
	"io"
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"

	"tked/internal/rope"
	"tked/internal/tklog"
)

type View interface {
	// Returns the buffer in this view.
	Buffer() Buffer

	// TODO: It might be nice to remove this. It is only used for drawing
	// and scrolling to ensure the cursor is visible.
	// Size returns the number of rows and columns visible in the view.
	Size() (int, int)
	// Resize updates the number of rows and columns visible in the view.
	Resize(rows, cols int)

	// TopLeft returns the top row and left column offsets.
	TopLeft() (int, int)
	// SetTopLeft updates the view's top row and left column offsets.
	SetTopLeft(top, left int)

	// Cursor returns the current cursor position as row and column indexes.
	Cursor() (int, int)
	// SetCursor updates the current cursor position and moves the viewport
	// to ensure the cursor is visible.
	SetCursor(row, col int)

	// Selections returns the list of selected regions in the view.
	Selections() []Selection
	// SetSelections replaces the current list of selections.
	SetSelections(selections []Selection)

	// Anchor returns the selection anchor position if present.
	Anchor() (int, int, bool)
	// SetAnchor updates the selection anchor position.
	SetAnchor(row, col int)
	// ClearAnchor removes the selection anchor.
	ClearAnchor()

	// InsertRune inserts a rune into the buffer at the cursor position.
	InsertRune(r rune)
	// DeleteRune deletes a rune. When forward is true it deletes the rune
	// under the cursor (Delete key behaviour). Otherwise it deletes the
	// rune before the cursor (Backspace behaviour).
	DeleteRune(forward bool)

	// Draw renders the view's contents on the provided screen.
	// topOffset and leftOffset specify where to start drawing on the screen.
	Draw(screen tcell.Screen, topOffset, leftOffset int)

	// Save writes the buffer contents to disk using the filename. If fileanme
	// is empty, save uses the existing filename if set, otherwise it returns an error.
	Save(filename string) error
}

type view struct {
	buffer Buffer
	width  int
	height int
	top    int
	left   int

	// anchor holds the position where a selection started. When nil, there
	// is no active selection anchor.
	anchor *cursor
}

type cursor struct {
	row int
	col int
}

// Selection represents a region of selected text within a view.
// The start position is inclusive and the end position is exclusive.
type Selection struct {
	StartRow int
	StartCol int
	EndRow   int
	EndCol   int
}

// orderedSelection returns a Selection with start and end points ordered such
// that the start position is before the end position in the buffer.
func orderedSelection(aRow, aCol, row, col int) Selection {
	if row < aRow || (row == aRow && col < aCol) {
		return Selection{StartRow: row, StartCol: col, EndRow: aRow, EndCol: aCol}
	}
	return Selection{StartRow: aRow, StartCol: aCol, EndRow: row, EndCol: col}
}

func (v *view) Buffer() Buffer {
	return v.buffer
}

func (v *view) Size() (int, int) {
	return v.height, v.width
}

func (v *view) Resize(rows, cols int) {
	v.height = max(1, rows)
	v.width = max(1, cols)
}

func (v *view) TopLeft() (int, int) {
	return v.top, v.left
}

func (v *view) SetTopLeft(top, left int) {
	v.top = max(0, top)
	v.left = max(0, left)
}

func (v *view) Cursor() (int, int) {
	c := v.buffer.GetProperty(cursorProp).(*cursor)
	return c.row, c.col
}

func (v *view) SetCursor(row, col int) {
	// Ensure the cursor is within the bounds of the buffer.
	row = max(0, row)
	col = max(0, col)

	idxRowStart, row := v.buffer.IndexForRow(row)
	colInfos := parseRow(v.buffer, row, idxRowStart)
	if len(colInfos) == 0 {
		col = 0
	} else {
		col = min(col, len(colInfos))
	}

	v.buffer.SetProperty(cursorProp, &cursor{
		row: row,
		col: col,
	})

	// Adjust the viewport to ensure the cursor is visible.
	v.ensureCursorVisible()
}

func (v *view) Selections() []Selection {
	prop := v.buffer.GetProperty(selectionsProp)
	if prop == nil {
		return nil
	}
	selections := prop.([]Selection)
	out := make([]Selection, len(selections))
	copy(out, selections)
	return out
}

func (v *view) SetSelections(selections []Selection) {
	if selections == nil {
		v.buffer.SetProperty(selectionsProp, nil)
		return
	}
	copySelections := make([]Selection, len(selections))
	copy(copySelections, selections)
	v.buffer.SetProperty(selectionsProp, copySelections)
}

// Anchor returns the current selection anchor. The boolean return value is
// false when there is no active anchor.
func (v *view) Anchor() (int, int, bool) {
	if v.anchor == nil {
		return 0, 0, false
	}
	return v.anchor.row, v.anchor.col, true
}

// SetAnchor sets the selection anchor to the provided position.
func (v *view) SetAnchor(row, col int) {
	v.anchor = &cursor{row: row, col: col}
}

// ClearAnchor removes any active selection anchor.
func (v *view) ClearAnchor() {
	v.anchor = nil
}

func (v *view) InsertRune(r rune) {
	cursorRow, cursorCol := v.Cursor()
	idxForRow, cursorRow := v.buffer.IndexForRow(cursorRow)
	idx := idxForRow
	colInfos := parseRow(v.buffer, cursorRow, idx)
	if len(colInfos) != 0 {
		colInfoIdx := cursorCol
		if cursorCol >= len(colInfos) {
			colInfoIdx = len(colInfos) - 1
			idx = colInfos[colInfoIdx].idx + 1
		} else {
			idx = colInfos[colInfoIdx].idx
		}
	}
	v.buffer.Insert(idx, string(r))

	if r == '\n' {
		cursorRow++
		cursorCol = 0
	} else {
		cursorCol++

		// parse the row to find the next character, in case we inserted a tab
		colInfos = parseRow(v.buffer, cursorRow, idxForRow)
		if cursorCol < len(colInfos) {
			for ; !colInfos[cursorCol].newChar; cursorCol++ {
				if cursorCol >= len(colInfos) {
					break
				}
			}
		}
	}

	v.SetCursor(cursorRow, cursorCol)
}

func (v *view) DeleteRune(forward bool) {
	// If there is a selection, delete the selected text and clear selections.
	selections := v.Selections()
	if len(selections) > 0 {
		startIdx := v.indexForRowCol(selections[0].StartRow, selections[0].StartCol)
		endIdx := v.indexForRowCol(selections[0].EndRow, selections[0].EndCol)
		v.buffer.Delete(startIdx, endIdx)
		v.SetCursor(selections[0].StartRow, selections[0].StartCol)
		v.SetSelections([]Selection{})
		return
	}

	cursorRow, cursorCol := v.Cursor()

	idxForRow, cursorRow := v.buffer.IndexForRow(cursorRow)
	colInfos := parseRow(v.buffer, cursorRow, idxForRow)

	// if cursor is inside a multi-character rune, delete the entire rune
	if cursorCol < len(colInfos) && !colInfos[cursorCol].newChar {
		idx := colInfos[cursorCol].idx
		v.buffer.Delete(idx, idx+1)

		// move the cursor to where the multi-character rune started
		for ; !colInfos[cursorCol].newChar; cursorCol-- {
			if cursorCol < 0 {
				tklog.Panic("cursorCol < 0") // bug, not error
			}
		}
		v.SetCursor(cursorRow, cursorCol)
	} else {
		var idx int
		if cursorCol >= len(colInfos) {
			idx = colInfos[len(colInfos)-1].idx + 1
		} else {
			idx = colInfos[cursorCol].idx
		}

		if forward {
			v.buffer.Delete(idx, idx+1)
			// Cursor doesn't move in this case
		} else {
			// Cursor moves back one character, handling the case where it was at the start of a line
			cursorCol--
			if cursorCol < 0 {
				cursorRow--
				if cursorRow < 0 {
					cursorRow = 0
				} else {
					// Set cursorCol to the end of the previous line

					// Start at the beginning of the previous line
					idxForRow, cursorRow = v.buffer.IndexForRow(cursorRow)

					// Scan to the end of the line
					colInfos = parseRow(v.buffer, cursorRow, idxForRow)
					cursorCol = len(colInfos)
				}
			}

			// Actaully delete the character now- we don't do it before
			// because we need to know the length of the previous line
			v.buffer.Delete(idx-1, idx)
			v.SetCursor(cursorRow, cursorCol)
		}
	}
}

func (v *view) Draw(screen tcell.Screen, topOffset, leftOffset int) {
	viewHeight, viewWidth := v.Size()
	viewTop, viewLeft := v.TopLeft()
	selections := v.Selections()

	idxRowStart, _ := v.buffer.IndexForRow(viewTop)
	for row := viewTop; row < viewTop+viewHeight; row++ {
		colInfos := parseRow(v.buffer, row, idxRowStart)
		if len(colInfos) == 0 {
			idxRowStart++
		} else {
			for col, colInfo := range colInfos {
				if col >= viewLeft && col < viewLeft+viewWidth {
					style := tcell.StyleDefault
					if isSelected(selections, row, col) {
						style = style.Reverse(true)
					}
					screen.SetContent(leftOffset+col-viewLeft, topOffset+row-viewTop, colInfo.r, nil, style)
				}
				idxRowStart = colInfo.idx + 2
			}
		}
	}

	cursorRow, cursorCol := v.Cursor()
	if cursorRow >= viewTop && cursorCol >= viewLeft && cursorRow < viewTop+viewHeight-1 && cursorCol < viewLeft+viewWidth {
		screen.ShowCursor(leftOffset+cursorCol-viewLeft, topOffset+cursorRow-viewTop)
	} else {
		screen.HideCursor()
	}
}

func (v *view) Save(filename string) error {
	if filename == "" {
		filename = v.buffer.GetFilename()
	}

	if filename == "" {
		return os.ErrInvalid
	}

	dir, name := filepath.Split(filename)

	// Create a temporary file in the same directory so that os.Rename works
	// across filesystems.
	tmp, err := os.CreateTemp(dir, name+".tmp*")
	if err != nil {
		return err
	}

	// Write contents to the temporary file first.
	if _, err := v.buffer.Write(tmp); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}

	// Atomically replace the target file.
	if err := os.Rename(tmp.Name(), filename); err != nil {
		os.Remove(tmp.Name())
		return err
	}

	return nil
}

func (v *view) ensureCursorVisible() {
	cursorRow, cursorCol := v.Cursor()
	if cursorRow < v.top {
		v.top = cursorRow
	} else if cursorRow >= v.top+v.height-1 {
		v.top = cursorRow - v.height + 2
	}

	if cursorCol < v.left {
		v.left = cursorCol
	} else if cursorCol >= v.left+v.width-1 {
		v.left = cursorCol - v.width + 1
	}
}

func (v *view) onBufferChange(buffer Buffer, start, end int, context any) {
	v.ensureCursorVisible()
}

// Create a new view with the given filename and contents. If contents is nil,
// an empty rope is used. The empty filename is used for unnamed views.
func NewView(filename string, contents rope.Rope) View {
	registerViewProperties()

	if contents == nil {
		contents = rope.NewRope("")
	}

	v := &view{
		buffer: NewBuffer(filename, contents),
		width:  80,
		height: 24,
		top:    0,
		left:   0,
		anchor: nil,
	}
	v.SetCursor(0, 0)
	v.SetSelections([]Selection{})
	v.buffer.OnChange(v.onBufferChange, v)
	return v
}

// Create a new view with the given filename and contents read from the reader.
func NewViewFromReader(filename string, r io.Reader) (View, error) {
	contents, err := rope.NewFromReader(r)
	if err != nil {
		return nil, err
	}
	return NewView(filename, contents), nil
}

var cursorProp PropKey
var selectionsProp PropKey

func registerViewProperties() {
	if cursorProp == nil {
		cursorProp = RegisterBufferProperty()
	}
	if selectionsProp == nil {
		selectionsProp = RegisterBufferProperty()
	}
}

type colInfo struct {
	newChar bool
	r       rune
	idx     int
}

func (v *view) indexForRowCol(row, col int) int {
	idxRowStart, actualRow := v.buffer.IndexForRow(row)
	row = actualRow
	colInfos := parseRow(v.buffer, row, idxRowStart)
	if len(colInfos) == 0 {
		return idxRowStart
	}
	if col >= len(colInfos) {
		return colInfos[len(colInfos)-1].idx + 1
	}
	return colInfos[col].idx
}

func parseRow(buffer Buffer, row int, idxStart int) []colInfo {
	tabWidth := GetApp().Settings().TabWidth()
	colInfos := []colInfo{}

	idx := idxStart
	col := 0
	for {
		r, ok := buffer.Contents().Index(idx)
		if !ok || r == '\n' {
			// End of buffer or end of line
			break
		} else if r == '\t' {
			width := tabWidth
			if col%tabWidth != 0 {
				width = tabWidth - col%tabWidth
			}
			for i := range width {
				colInfos = append(colInfos, colInfo{
					newChar: i == 0,
					r:       ' ',
					idx:     idx,
				})
			}
			col += width
			idx++
		} else {
			colInfos = append(colInfos, colInfo{
				newChar: true,
				r:       rune(r),
				idx:     idx,
			})
			col++
			idx++
		}
	}

	return colInfos
}

func isSelected(selections []Selection, row, col int) bool {
	for _, s := range selections {
		if row < s.StartRow || row > s.EndRow {
			continue
		}
		if s.StartRow == s.EndRow {
			if col >= s.StartCol && col < s.EndCol {
				return true
			}
		} else if row == s.StartRow {
			if col >= s.StartCol {
				return true
			}
		} else if row == s.EndRow {
			if col < s.EndCol {
				return true
			}
		} else {
			return true
		}
	}
	return false
}
