package loom

type scrollbarAxis int

const (
	verticalScrollbar scrollbarAxis = iota
	horizontalScrollbar
)

type scrollbarDrag struct {
	active  bool
	start   int
	grab    int
	initial int
}

func scrollbarThumbLength(track, content, viewport int) int {
	if track <= 0 || content <= viewport || content <= 0 || viewport <= 0 {
		return 0
	}
	thumb := (track*viewport + content/2) / content
	if thumb < 1 {
		thumb = 1
	}
	if thumb > track {
		thumb = track
	}
	return thumb
}

func scrollbarThumbStart(track, thumb, offset, maxOffset int) int {
	travel := track - thumb
	if travel <= 0 || maxOffset <= 0 {
		return 0
	}
	return (offset*travel + maxOffset/2) / maxOffset
}

func scrollbarOffset(track, thumb, pointer, grab, maxOffset int) int {
	travel := track - thumb
	if travel <= 0 || maxOffset <= 0 {
		return 0
	}
	pos := pointer - grab
	if pos < 0 {
		pos = 0
	}
	if pos > travel {
		pos = travel
	}
	return (pos*maxOffset + travel/2) / travel
}

// scrollbarUnitsPerCell is the number of half-row units in one track row. The
// unit count per cell is one constant so finer steps can follow.
const scrollbarUnitsPerCell = 2

// scrollbarUnitPointer returns the unit at the middle of track row row. The
// pointer reports whole rows, so every press and drag uses the row centre.
func scrollbarUnitPointer(row int) int {
	return row*scrollbarUnitsPerCell + scrollbarUnitsPerCell/2
}

// scrollbarUnitThumb returns the thumb start and length in half-row units for
// a track of rows cells. The thumb is at least one cell long and never longer
// than the track. Content that fits the viewport keeps a one-cell thumb at the
// top, as the cell-based bar always did.
func scrollbarUnitThumb(rows, content, viewport, offset, maxOffset int) (start, length int) {
	units := rows * scrollbarUnitsPerCell
	if units <= 0 || content <= 0 || viewport <= 0 {
		return 0, 0
	}
	if content <= viewport {
		return 0, min(scrollbarUnitsPerCell, units)
	}
	length = (units*viewport + content/2) / content
	length = min(max(length, scrollbarUnitsPerCell), units)
	travel := units - length
	if travel <= 0 || maxOffset <= 0 {
		return 0, length
	}
	offset = min(max(offset, 0), maxOffset)
	return (offset*travel + maxOffset/2) / maxOffset, length
}

// scrollbarUnitOffset maps a pointer unit to a content offset. The thumb's
// first unit is pointer-grab, so a drag keeps the grabbed half-row phase.
func scrollbarUnitOffset(rows, length, pointer, grab, maxOffset int) int {
	travel := rows*scrollbarUnitsPerCell - length
	if travel <= 0 || maxOffset <= 0 {
		return 0
	}
	pos := min(max(pointer-grab, 0), travel)
	return (pos*maxOffset + travel/2) / travel
}

// scrollbarRowTouchesThumb reports whether any part of track row row lies on
// the thumb. A row the thumb only partly covers still counts as thumb.
func scrollbarRowTouchesThumb(row, start, length int) bool {
	lo := row * scrollbarUnitsPerCell
	return start < lo+scrollbarUnitsPerCell && start+length > lo
}

// scrollbarColumn returns the cells of one scrollbar column of rows cells. With
// a themed track and thumb on colour output, each row is a space on the track or
// thumb background, or a lower half block for a thumb that starts or ends
// inside the row. Otherwise each row is a whole-cell glyph, with the thumb
// rounded to the nearest cell edge on each side.
func scrollbarColumn(style ScrollbarStyle, profile ColorProfile, rows, start, length int) []Cell {
	cells := make([]Cell, rows)
	halfCells := profile != ColorProfileNone && style.Track.BG != ColorReset() &&
		style.Thumb.BG != ColorReset() && style.Track.BG != style.Thumb.BG
	if !halfCells {
		first := (start + 1) / 2
		end := (start + length + 1) / 2
		for row := range cells {
			if row >= first && row < end {
				cells[row] = Cell{Text: SpeccedDefaults.Scrollbar.ForegroundChar, Style: style.Thumb}
			} else {
				cells[row] = Cell{Text: SpeccedDefaults.Scrollbar.BackgroundChar, Style: style.Track}
			}
		}
		return cells
	}
	for row := range cells {
		lo := row * scrollbarUnitsPerCell
		upper := start <= lo && start+length > lo
		lower := start <= lo+1 && start+length > lo+1
		switch {
		case upper && lower:
			cells[row] = Cell{Text: " ", Style: style.Thumb}
		case lower:
			cells[row] = Cell{Text: "▄", Style: Style{FG: style.Thumb.BG, BG: style.Track.BG}}
		case upper:
			cells[row] = Cell{Text: "▄", Style: Style{FG: style.Track.BG, BG: style.Thumb.BG}}
		default:
			cells[row] = Cell{Text: " ", Style: style.Track}
		}
	}
	return cells
}

func (d *scrollbarDrag) press(pointer, thumbStart, offset int) {
	d.active = true
	d.start = offset
	d.initial = offset
	d.grab = pointer - thumbStart
}

func (d *scrollbarDrag) cancel() {
	d.active = false
}
