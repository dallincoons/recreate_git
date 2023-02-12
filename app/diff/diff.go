package diff

import (
	"fmt"
	"jit/format"
	"strings"
)

type Edit struct {
	kind  EditType
	aLine *Line
	bLine *Line
}

type Line struct {
	Number int
	Text   string
}

type Hunk struct {
	aStart int
	bStart int
	Edits  []Edit
}

func (h *Hunk) Header() string {
	aOffset, alineSize := h.offsetsFor('a', h.aStart)
	bOffset, blineSize := h.offsetsFor('b', h.bStart)

	return format.Color("bold", fmt.Sprintf("@@ -%d,%d +%d,%d @@", aOffset, alineSize, bOffset, blineSize))
}

func (h *Hunk) offsetsFor(lineType byte, start int) (int, int) {
	var lines []*Line
	for _, edit := range h.Edits {
		if lineType == 'a' {
			if edit.aLine != nil {
				lines = append(lines, edit.aLine)
				start = edit.aLine.Number
			}
		}

		if lineType == 'b' {
			if edit.bLine != nil {
				lines = append(lines, edit.bLine)
				start = edit.bLine.Number
			}
		}
	}

	return start, len(lines)
}

func NewHunk(aStart, bStart int, edits []Edit) *Hunk {
	return &Hunk{
		aStart: aStart,
		bStart: bStart,
		Edits:  edits,
	}
}

func HunkBuild(hunk *Hunk, edits []Edit, offsetStart int) int {
	counter := -1
	offset := offsetStart

	for counter != 0 {
		if offset >= 0 && counter > 0 {
			hunk.Edits = append(hunk.Edits, edits[offset])
		}

		offset += 1
		if offset >= len(edits) {
			break
		}

		var editType *EditType

		if offset+HUNK_CONTEXT < len(edits) {
			editType = &edits[offset+HUNK_CONTEXT].kind
		}

		if editType == nil {
			counter -= 1
			continue
		}

		switch *editType {
		case EditTypeInsert:
		case EditTypeDelete:
			counter = 2*HUNK_CONTEXT + 1
		default:
			counter -= 1
		}
	}

	return offset
}

type EditType string

var (
	EditTypeInsert EditType = "insert"
	EditTypeDelete EditType = "delete"
	EditTypeEqual  EditType = "equal"
)

func NewEdit(eType EditType, aLine, bLine *Line) Edit {
	return Edit{
		kind:  eType,
		aLine: aLine,
		bLine: bLine,
	}
}

func NewLine(number int, text string) Line {
	return Line{
		Number: number,
		Text:   text,
	}
}

const HUNK_CONTEXT = 3

func Lines(data []string) []Line {
	lines := make([]Line, 0, len(data))

	for i, str := range data {
		lines = append(lines, NewLine(i, strings.Trim(str, "\000")))
	}

	return lines
}

func DiffHunks(a, b []Line) []*Hunk {
	//for _, e := range Diff(a, b) {
	//	var edit string
	//	if e.aLine != nil {
	//		edit = e.aLine.Text
	//	} else if e.bLine != nil {
	//		edit = e.bLine.Text
	//	}
	//
	//	fmt.Println(e.getDisplaySymbol(), edit)
	//}

	return hunkFilter(Diff(a, b))
}

func Diff(a, b []Line) []Edit {
	myers := NewMyers(a, b)

	return myers.diff()
}

func hunkFilter(edits []Edit) []*Hunk {
	editLength := len(edits)

	hunks := make([]*Hunk, 0)
	offset := 0

	for true {
		for offset < editLength && offset >= 0 && edits[offset].kind == EditTypeEqual {
			offset += 1
		}

		if offset >= editLength {
			return hunks
		}

		offset -= HUNK_CONTEXT + 1

		aStart := 0
		bStart := 0
		if offset > 0 {
			aStart = edits[offset].aLine.Number
			bStart = edits[offset].bLine.Number
		}

		newHunk := NewHunk(aStart, bStart, make([]Edit, 0))

		hunks = append(hunks, newHunk)
		offset = HunkBuild(newHunk, edits, offset)
	}

	return hunks
}
