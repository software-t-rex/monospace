/*
Copyright © 2024 Jonathan Gotti <jgotti at jgotti dot org>
SPDX-FileType: SOURCE
SPDX-License-Identifier: MIT
SPDX-FileCopyrightText: 2024 Jonathan Gotti <jgotti@jgotti.org>
*/

package lineEditor

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/software-t-rex/monospace/gomodules/ui/pkg/ansi"
)

// @TODO: add support for \t\n in WrappedLine mode

type VisualEditMode int

const (
	VModeHidden      VisualEditMode = iota // no visual edition
	VModeUnbounded                         // visual edition on a single line without managing term width
	VModeWrappedLine                       // visual edition wrapping lines when term width is reached
	VModeFramedLine                        // visual edition with a single line frame of term width
	VModeMaskedLine                        // visual edition with masking of the content
)

var ErrMaxLen = fmt.Errorf("max length reached")

type LineEditor struct {
	value            []rune
	pos              int // cursor position inside the value (!= visual cursor position)
	out              io.Writer
	completing       bool
	compStart        int
	compSuggests     []string
	compSuggestIndex int
	visualMode       VisualEditMode
	maxLen           int
	maxWidth         int // as received from the config

	// this is the terminal width in characters
	termWidth int
	wrapWidth int

	// this is the offset at wich the first line starts
	visualStartOffset int

	// used to keep track of the window start position in framed mode
	frameStartPos int
}
type LineEditorOptions struct {
	// [2]int{Row, Col} allow placement of the visual edition within the terminal
	// Row is relative to the current cursor position Positive values will move up and negative values will move down
	// Col is an absolute position on the line and can only be positive (0 based)
	// Col should at least be equal to prompt size and will set the visualStartOffset
	VisualPos [2]int
	// this is the maximum width of the visual editor in characters
	// It will be capped to termWidth - visualStartOffset - 1 (for the cursor)
	MaxWidth int
	// this is the maximum length of the value (0 for no limit)
	MaxLen int
	// this is the initial value when edition starts
	Value string
}

func NewLineEditor(out io.Writer, visualEditionMode VisualEditMode) *LineEditor {
	return &LineEditor{
		out:        out,
		visualMode: visualEditionMode,
	}
}

// you should have set termWidth before calling this method
func (l *LineEditor) SetConfig(cfg LineEditorOptions) *LineEditor {
	l.maxLen = cfg.MaxLen
	l.maxWidth = cfg.MaxWidth
	l.checkFrameWidth()
	// set initial position
	if cfg.VisualPos[0] != 0 {
		if cfg.VisualPos[0] > 0 {
			ansi.CtrlUp.Fprintf(l.out, cfg.VisualPos[0])
		} else {
			ansi.CtrlDown.Fprintf(l.out, -cfg.VisualPos[0])
		}
	}
	if cfg.VisualPos[1] != 0 {
		l.visualStartOffset = cfg.VisualPos[1]
	}
	if l.visualStartOffset > 0 { //negative values are not allowed
		ansi.CtrlHorizAbs.Fprintf(l.out, l.visualStartOffset+1)
	}
	if cfg.Value != "" {
		l.Insert([]rune(cfg.Value))
	}
	return l
}

func (l *LineEditor) SetVisualStartOffset(offset int) *LineEditor {
	l.visualStartOffset = offset
	return l
}

func (l *LineEditor) SetTermWidth(width int) *LineEditor {
	l.termWidth = width
	return l
}

func (l *LineEditor) GetValue() []rune {
	return l.value
}
func (l *LineEditor) GetStringValue() string {
	return string(l.value)
}

func (l *LineEditor) IsCompleting() bool {
	return l.completing
}

func (l *LineEditor) checkFrameWidth() {
	calcWidth := l.termWidth - l.visualStartOffset
	if l.visualMode != VModeWrappedLine {
		calcWidth -= 1 // for the cursor
	}
	if l.maxWidth > 0 {
		l.wrapWidth = min(l.maxWidth, calcWidth)
	} else {
		l.wrapWidth = calcWidth
	}
}

func (l *LineEditor) Len() int { return len(l.value) }

func (l *LineEditor) RingBell() error {
	_, err := fmt.Fprint(l.out, "\a")
	return err
}

func (l *LineEditor) Insert(s []rune) error {
	slen := len(s)
	if slen > 1 {
		s = Sanitize(s, false)
		slen = len(s)
		if slen < 1 {
			return nil
		}
	}
	// don't exceed maxLen
	if l.maxLen > 0 && l.Len()+slen > l.maxLen {
		l.RingBell()
		if l.Len() == l.maxLen {
			return ErrMaxLen
		} else {
			s = s[:max(0, l.maxLen-l.Len())]
			slen = len(s)
			if slen < 1 {
				return ErrMaxLen
			}
		}
	}
	remain := append(s, l.value[l.pos:]...)
	return l.RewriteFrom(l.pos, l.pos+slen, remain)
}

func (l *LineEditor) Clear() error {
	return l.RewriteFrom(0, 0, []rune{})
}
func (l *LineEditor) Delete() error {
	if l.pos >= l.Len() { // nothing to delete
		return nil
	}
	return l.RewriteFrom(l.pos, l.pos, copyLineFrom(l, l.pos+1))
}
func (l *LineEditor) DeleteBackward() error {
	if l.pos < 1 { // nothing to delete
		return nil
	}
	return l.RewriteFrom(l.pos-1, l.pos-1, copyLineFrom(l, l.pos))
}
func (l *LineEditor) DeleteToStart() error {
	if l.pos < 1 { // nothing to delete
		return nil
	}
	return l.RewriteFrom(0, 0, copyLineFrom(l, l.pos))
}
func (l *LineEditor) DeleteToEnd() error {
	if l.pos >= l.Len() {
		return nil
	}
	return l.RewriteFrom(l.pos, l.pos, []rune{})
}
func (l *LineEditor) DeleteWord() error {
	if l.pos >= l.Len() { // nothing to delete
		return nil
	}
	endPos := l.FindWordBoundary("right")
	if endPos == l.pos {
		return nil
	}
	return l.RewriteFrom(l.pos, l.pos, copyLineFrom(l, endPos))
}
func (l *LineEditor) DeleteWordBackward() error {
	if l.pos < 1 { // nothing to delete
		return nil
	}
	start := l.FindWordBoundary("left")
	if start == l.pos {
		return nil
	}
	return l.RewriteFrom(start, start, copyLineFrom(l, l.pos))
}

func (l *LineEditor) MoveStartOfLine() error {
	if l.visualMode != VModeWrappedLine {
		return l.MoveStart()
	}
	_, col := l.getVisualRowColFromPos(l.pos)
	if col == 0 {
		return nil
	}
	return l.moveCursorToPosition(l.pos - col)
}
func (l *LineEditor) MoveEndOfLine() error {
	if l.visualMode != VModeWrappedLine {
		return l.MoveEnd()
	}
	_, col := l.getVisualRowColFromPos(l.pos)
	if col == l.wrapWidth {
		return nil
	}
	remainSpaceOnLine := l.wrapWidth - col - 1
	return l.moveCursorToPosition(min(l.Len(), l.pos+remainSpaceOnLine))
}
func (l *LineEditor) MoveStart() error {
	if l.pos < 1 {
		return nil
	} else if l.visualMode == VModeFramedLine && l.frameStartPos > 0 {
		return l.RewriteFrom(0, 0, l.value)
	}
	return l.moveCursorToPosition(0)
}
func (l *LineEditor) MoveEnd() error {
	if l.pos >= l.Len() {
		return nil
	} else if l.visualMode == VModeFramedLine && l.frameStartPos+l.wrapWidth < l.Len() {
		return l.RewriteFrom(l.Len()-l.wrapWidth, l.Len(), copyLineFrom(l, l.Len()-l.wrapWidth))
	}
	return l.moveCursorToPosition(l.Len())
}
func (l *LineEditor) MoveLeft() error {
	newPos := l.pos - 1
	if l.pos <= 0 {
		return nil
	} else if l.visualMode == VModeFramedLine && newPos < l.frameStartPos {
		return l.RewriteFrom(newPos, newPos, copyLineFrom(l, newPos))
	}
	return l.moveCursorToPosition(newPos)
}
func (l *LineEditor) MoveRight() error {
	newPos := l.pos + 1
	if l.pos >= l.Len() {
		return nil
	} else if l.visualMode == VModeFramedLine && newPos > l.frameStartPos+l.wrapWidth {
		return l.RewriteFrom(newPos-l.wrapWidth, newPos, copyLineFrom(l, newPos-l.wrapWidth))
	}
	return l.moveCursorToPosition(newPos)
}
func (l *LineEditor) MoveWordLeft() error {
	if l.pos < 1 {
		return nil
	}
	start := l.FindWordBoundary("left")
	if l.visualMode == VModeFramedLine && start < l.frameStartPos {
		return l.RewriteFrom(start, start, copyLineFrom(l, start))
	}
	return l.moveCursorToPosition(start)
}
func (l *LineEditor) MoveWordRight() error {
	if l.pos >= l.Len() {
		return nil
	}
	end := l.FindWordBoundary("right")
	if l.visualMode == VModeFramedLine && end > l.frameStartPos+l.wrapWidth {
		return l.RewriteFrom(end-l.wrapWidth, end, copyLineFrom(l, end-l.wrapWidth))
	}
	return l.moveCursorToPosition(end)
}
func (l *LineEditor) MoveUp() error {
	if l.visualMode != VModeWrappedLine && !(l.visualMode == VModeFramedLine && l.termWidth > 0) { // up is only supported in wrapped line mode
		return l.RingBell()
	}
	row, _ := l.getVisualRowColFromPos(l.pos)
	if row < 1 {
		return l.RingBell() // can't move up
	}
	l.pos -= l.wrapWidth
	_, err := ansi.CtrlUp.Fprintf(l.out, 1)
	return err
}
func (l *LineEditor) MoveDown() error {
	if l.visualMode != VModeWrappedLine && !(l.visualMode == VModeFramedLine && l.termWidth > 0) { // down is only supported in wrapped line mode
		return l.RingBell()
	}
	row, _ := l.getVisualRowColFromPos(l.pos)
	if row >= l.Len()/l.wrapWidth {
		return l.RingBell() // can't move down
	}
	return l.moveCursorToPosition(min(l.Len(), l.pos+l.wrapWidth))
}

// this is will move cursor to the given position (it does not redraw content)
// Cursor position will be capped in the actual viewing area.
// It is used internally in visual mode to move the visual cursor without redrawing the content
//
// pos will be updated
func (l *LineEditor) moveCursorToPosition(targetPos int) error {
	if l.pos == targetPos {
		return nil // nothing to do
	} else if l.visualMode <= VModeHidden {
		l.pos = targetPos
		return nil
	}
	row, col := l.getVisualRowColFromPos(l.pos)
	targetRow, targetCol := l.getVisualRowColFromPos(targetPos)
	sb := strings.Builder{}
	// move visual cursor to fromPos
	if targetRow > row {
		sb.WriteString(ansi.CtrlDown.Sprintf(targetRow - row))
	} else if targetRow < row {
		sb.WriteString(ansi.CtrlUp.Sprintf(row - targetRow))
	}
	if targetCol > col {
		sb.WriteString(ansi.CtrlForward.Sprintf(targetCol - col))
	} else if targetCol < col {
		sb.WriteString(ansi.CtrlBackward.Sprintf(col - targetCol))
	}
	l.pos = targetPos
	_, err := fmt.Fprintf(l.out, "%s", sb.String())
	return err
}

// it will move cursor to insertPos and redraw the content from that position with the new content
// it will update the value accordingly and update the cursor position to finalPos
func (l *LineEditor) RewriteFrom(insertPos int, finalPos int, contentFromInsertPos []rune) error {
	newValue := append(copyLineTo(l, insertPos), contentFromInsertPos...)
	if l.visualMode <= VModeHidden { // no visual edition
		l.value = newValue
		l.pos = finalPos
		return nil
	}

	// in all other cases we consider the cursor to be abble to move to the new position
	if err := l.moveCursorToPosition(insertPos); err != nil {
		return fmt.Errorf("redraw failed move cursor %w", err)
	}
	if err := l.visualClearToEnd(); err != nil {
		return fmt.Errorf("redraw failed clear %w", err)
	}

	// write the new content into buffer and then to screen
	l.value = newValue
	var appendErr error
	switch l.visualMode {
	case VModeWrappedLine:
		appendErr = l.appendWrapped(contentFromInsertPos)
	case VModeFramedLine:
		appendErr = l.appendFramed(contentFromInsertPos, finalPos)
	case VModeUnbounded:
		appendErr = l.appendUnbounded(contentFromInsertPos)
	case VModeMaskedLine:
		contentFromInsertPos = []rune(strings.Repeat("*", len(contentFromInsertPos)))
		if l.termWidth > 0 {
			appendErr = l.appendWrapped(contentFromInsertPos)
		} else {
			appendErr = l.appendUnbounded(contentFromInsertPos)
		}
	default: // should not happen unless we add new visual mode
		appendErr = l.appendUnbounded(contentFromInsertPos)
	}
	if appendErr != nil {
		return fmt.Errorf("redraw failed append %w", appendErr)
	}
	// now ce can update the cursor position to final destination
	return l.moveCursorToPosition(finalPos)
}

// this function is used to print the content of the line editor as it was in its last state
// It will not clean the screen before printing and will not move the cursor
// Just print the content of the line editor as it is respecting the visual mode
func (l *LineEditor) Sprint() (string, error) {
	if l.visualMode <= VModeHidden {
		return "", nil
	}
	buf := new(strings.Builder)
	// out := bufio.NewWriter(buf)
	tmpLine := NewLineEditor(buf, l.visualMode).
		SetTermWidth(l.termWidth).
		SetConfig(LineEditorOptions{
			MaxWidth: l.maxWidth,
			MaxLen:   l.maxLen,
		}).
		SetVisualStartOffset(l.visualStartOffset)
	errWrite := tmpLine.Insert(l.value)
	return buf.String(), errWrite
}

func (l *LineEditor) appendUnbounded(str []rune) error {
	l.pos = l.Len()
	_, err := fmt.Fprint(l.out, string(str))
	return err
}

// appendFramed will append the string to the screen and update the frameStartPos if necessary
// toPos is the position where the cursor should be after the append
// it won't update the cursor position just need to know it to ensure the frameStartPos is correct
func (l *LineEditor) appendFramed(str []rune, toPos int) error {
	needToDrawFullFrame := false
	endOfFrame := l.frameStartPos + l.wrapWidth
	length := l.Len()
	if l.pos < l.frameStartPos || l.pos > endOfFrame || length < l.frameStartPos || length > endOfFrame {
		// if any position is out of the frame we dumbly redraw all the frame
		// this is a simple implementation that leave room for optimization if needed
		needToDrawFullFrame = true
	}
	if !needToDrawFullFrame {
		// we are in the frame so just append the string
		return l.appendUnbounded(str)
	}
	_, col := l.getVisualRowColFromPos(l.pos)
	var moveClearErr error
	if col > 0 {
		moveClearErr = l.moveCursorToPosition(l.frameStartPos)
		if moveClearErr == nil {
			moveClearErr = l.visualClearToEnd()
		}
	}
	// ensure frame will cover the cursor position
	if toPos < l.frameStartPos {
		l.frameStartPos = toPos
	} else if toPos > l.frameStartPos+l.wrapWidth {
		l.frameStartPos = toPos - l.wrapWidth
	}
	endInsertPos := min(l.frameStartPos+l.wrapWidth, length)
	framedVal := l.value[l.frameStartPos:endInsertPos]
	l.pos = endInsertPos
	if moveClearErr != nil {
		return fmt.Errorf("append failed move/clear %w", moveClearErr)
	}
	_, err := fmt.Fprint(l.out, string(framedVal))
	return err
}

func (l *LineEditor) appendWrapped(str []rune) error {
	lineRemainingSpace := l.wrapWidth - l.pos%l.wrapWidth
	if len(str) < lineRemainingSpace {
		// we can write the string on the current line
		l.pos = l.Len()
		_, err := fmt.Fprintf(l.out, "%s", string(str))
		return err
	}
	sb := strings.Builder{}
	// we need to split the string in lines that fits in l.wrapWidth
	// filling the first line remaining spaces
	sb.WriteString(string(str[:lineRemainingSpace]))
	str = str[lineRemainingSpace:]
	// move to next line
	for len(str) > l.wrapWidth {
		addVisualNewLine(&sb, l.visualStartOffset)
		sb.WriteString(string(str[:l.wrapWidth]))
		str = str[l.wrapWidth:]
	}
	// is there is still some content to write write a last line
	if len(str) >= 0 {
		addVisualNewLine(&sb, l.visualStartOffset)
		sb.WriteString(string(str))
		if len(str) == l.wrapWidth {
			addVisualNewLine(&sb, l.visualStartOffset)
		}
	}
	l.pos = l.Len()
	_, err := fmt.Fprintf(l.out, "%s", sb.String())
	return err
}

// clear the current line from cursor position to the end
// restore the cursor position
func (l *LineEditor) visualClearToEnd() error {
	if l.visualMode <= VModeHidden {
		return nil
	}
	row, col := l.getVisualRowColFromPos(l.pos)
	endRow, _ := l.getVisualRowColFromPos(l.Len())
	sb := strings.Builder{}
	// erase end of current line
	sb.WriteString(ansi.CtrlEraseL.Sprintf(0))
	remainingLines := endRow - row
	if remainingLines > 0 {
		// erase all lines between row and endRow
		sb.WriteString(ansi.CtrlHorizAbs.Sprintf(max(1, l.visualStartOffset-1)))
		sb.WriteString(strings.Repeat(ansi.CtrlDown.Sprintf(1)+ansi.CtrlEraseL.Sprintf(0), remainingLines))
		// move back to current line
		sb.WriteString(ansi.CtrlUp.Sprintf(remainingLines))
		// move back to col
		sb.WriteString(ansi.CtrlForward.Sprintf(col + l.visualStartOffset))
	}
	_, err := fmt.Fprintf(l.out, "%s", sb.String())
	return err
}

// transform linear position to visual row and col
func (l *LineEditor) getVisualRowColFromPos(pos int) (row, col int) {
	switch l.visualMode {
	case VModeHidden:
		return 0, 0
	case VModeUnbounded:
		return 0, pos + l.visualStartOffset
	case VModeWrappedLine:
		if pos < l.wrapWidth { // we are on the first line
			return 0, pos
		}
		return pos / l.wrapWidth, (pos % l.wrapWidth)
	case VModeFramedLine:
		return 0, min(max(0, pos-l.frameStartPos), l.frameStartPos+l.wrapWidth)
	case VModeMaskedLine:
		if l.termWidth > 0 {
			if pos < l.wrapWidth { // we are on the first line
				return 0, pos
			}
			return pos / l.wrapWidth, (pos % l.wrapWidth)
		}
		return 0, pos + l.visualStartOffset
	}
	return 0, 0
}

func (l *LineEditor) FindWordBoundaryFromPos(pos int, direction string) int {
	if direction == "left" {
		for pos > 0 && unicode.IsSpace(l.value[pos-1]) {
			pos--
		}
		for pos > 0 && !unicode.IsSpace(l.value[pos-1]) {
			pos--
		}
	} else if direction == "right" {
		for pos < l.Len() && unicode.IsSpace(l.value[pos]) {
			pos++
		}
		for pos < l.Len() && !unicode.IsSpace(l.value[pos]) {
			pos++
		}
	}
	return pos
}
func (l *LineEditor) FindWordBoundary(direction string) int {
	return l.FindWordBoundaryFromPos(l.pos, direction)
}

// Returns 2 strings the first is the part of the word to complete before the cursor
// the second is the whole word (including part of the word after the cursor if any)
//
// In order to implement completion:
//   - call CompletionStart to get the start and end of the word to complete
//   - call CompletionSuggests with the list of suggestions
//   - call CompletionNext to complete with the first suggestion
//   - call CompletionNext to complete with the next suggestion
//   - call CompletionEnd to end the completion mode
func (l *LineEditor) CompletionStart() (string, string) {
	if l.visualMode <= VModeHidden { // doesn't make sense in hidden mode
		return "", ""
	}
	l.completing = true
	start := l.FindWordBoundary("left")
	endWord := l.FindWordBoundaryFromPos(start, "right") // not starting from pos to avoid getting next word ending
	l.compStart = start
	return string(copyLineRange(l, start, l.pos)), string(copyLineRange(l, start, endWord))
}

// sets the completion suggestions and complete with the first one if any
func (l *LineEditor) CompletionSuggests(suggestions []string) *LineEditor {
	for i, s := range suggestions {
		suggestions[i] = string(Sanitize([]rune(s), false))
	}
	l.compSuggests = suggestions
	l.compSuggestIndex = -1
	l.CompletionNext()
	return l
}

// end the completion mode see [*lineEditor.CompletionStart]
func (l *LineEditor) CompletionEnd() *LineEditor {
	l.completing = false
	l.compStart = 0
	l.compSuggestIndex = -1
	l.compSuggests = nil
	return l
}

// complete with the next suggestion
func (l *LineEditor) CompletionNext() error {
	if !l.completing || len(l.compSuggests) == 0 {
		return nil
	}
	l.compSuggestIndex++
	if l.compSuggestIndex >= len(l.compSuggests) {
		l.compSuggestIndex = 0
	}
	return l.complete(l.compSuggests[l.compSuggestIndex])
}

// replace current word with given completion
func (l *LineEditor) complete(completion string) error {
	comp := []rune(completion)
	compLen := len(comp)
	// get word boundaries
	endWordPos := l.FindWordBoundaryFromPos(l.compStart, "right")
	if l.maxLen > 0 {
		lenWithoutWord := l.Len() - (endWordPos - l.compStart)
		if lenWithoutWord+compLen > l.maxLen {
			l.RingBell()
			comp = comp[:max(0, l.maxLen-lenWithoutWord)]
			compLen = len(comp)
			if compLen < 1 {
				return ErrMaxLen
			}
		}
	}
	// visual edition
	return l.RewriteFrom(l.compStart, l.compStart+len(comp), append(comp, copyLineFrom(l, endWordPos)...))
}

// -- utility functions
func copyLineRange(l *LineEditor, start int, end int) []rune {
	res := make([]rune, end-start)
	copy(res, l.value[start:end])
	return res
}
func copyLineFrom(l *LineEditor, start int) []rune {
	return copyLineRange(l, start, l.Len())
}
func copyLineTo(l *LineEditor, end int) []rune {
	return copyLineRange(l, 0, end)
}

// add a visual new line with offset in a string builder
// used internally by lineEditor
func addVisualNewLine(sb *strings.Builder, offset int) {
	sb.WriteString("\r\n")
	if offset > 0 {
		sb.WriteString(ansi.CtrlForward.Sprintf(offset))
	}
}

// returns a copy of value with special chars escaped
func Sanitize(value []rune, preserveBlank bool) []rune {
	// removes non printable characters or escape them and replace all blank spaces by a single space
	res := make([]rune, 0, len(value))
	inEscape := false
	for i, r := range value {
		if inEscape {
			if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r == '~' || r == '^' {
				if r == 'O' && value[i-1] == '\x1b' { // some escape sequence start with \x1bO
					continue
				}
				inEscape = false
			}
			continue
		}
		if unicode.IsPrint(r) {
			res = append(res, r)
		} else if unicode.IsSpace(r) {
			if !preserveBlank {
				res = append(res, ' ')
			} else {
				res = append(res, r)
			}

		} else if r == '\x1b' {
			inEscape = true
			continue
		}
		// leave out other non printable characters
	}
	return res
}
