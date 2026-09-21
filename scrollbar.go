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

func (d *scrollbarDrag) press(pointer, thumbStart, offset int) {
	d.active = true
	d.start = offset
	d.initial = offset
	d.grab = pointer - thumbStart
}

func (d *scrollbarDrag) cancel() {
	d.active = false
}
