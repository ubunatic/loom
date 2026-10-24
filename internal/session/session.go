// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build !windows

package session

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
	"ubunatic.com/loom"
	"ubunatic.com/loom/terminal"
	"ubunatic.com/loom/vt"
)

// ExitError reports the command's exit status after a child process exits.
type ExitError struct {
	code int
	err  error
}

func (e ExitError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return ""
}

func (e ExitError) Unwrap() error { return e.err }

// ExitCode returns the wrapped command's exit status.
func (e ExitError) ExitCode() int { return e.code }

// LongHelp returns the session command's detailed hotkey help.
func LongHelp() string {
	sessionDefs := loom.SpeccedDefaults.Session
	shotKey := strings.Join(sessionDefs.Bindings.Screenshot, ", ")
	quitKey := strings.Join(sessionDefs.Bindings.Quit, ", ")
	newTabKey := strings.Join(sessionDefs.Bindings.NewTab, ", ")
	searchKey := strings.Join(sessionDefs.Bindings.Search, ", ")
	prefixShot := strings.Join(sessionDefs.PrefixBindings.Screenshot, ", ")
	prefixQuit := strings.Join(sessionDefs.PrefixBindings.Quit, ", ")
	sendPrefix := strings.Join(sessionDefs.PrefixBindings.SendPrefix, ", ")
	prefixNewTab := strings.Join(sessionDefs.PrefixBindings.NewTab, ", ")
	nextTab := strings.Join(sessionDefs.PrefixBindings.NextTab, ", ")
	prevTab := strings.Join(sessionDefs.PrefixBindings.PreviousTab, ", ")
	closeTab := strings.Join(sessionDefs.PrefixBindings.CloseTab, ", ")
	longHelp := fmt.Sprintf(
		"Run the command in a pseudo-terminal, passing all input and output through unchanged. Hotkeys: %s saves an ANSI screenshot, %s quits, %s opens a new tab, %s opens scrollback search. Prefix %s followed by %s saves a screenshot, %s quits, and %s forwards %s to the child. Prefix %s opens a new tab, %s/%s switches tabs, %s closes a tab.",
		shotKey, quitKey, newTabKey, searchKey, sessionDefs.Prefix, prefixShot, prefixQuit, sendPrefix, sessionDefs.Prefix, prefixNewTab, nextTab, prevTab, closeTab,
	)
	return longHelp
}

// Run runs a command in a pseudo-terminal, forwarding input, output and hotkeys.
func Run(ctx context.Context, in io.Reader, out, errOut io.Writer, command string, args ...string) error {
	return RunWithOptions(ctx, in, out, errOut, Options{}, command, args...)
}

// Options configures optional session behavior.
type Options struct {
	Debug    bool
	DebugDir string
}

// RunWithOptions runs a command in a pseudo-terminal with optional features.
func RunWithOptions(ctx context.Context, in io.Reader, out, errOut io.Writer, options Options, command string, args ...string) error {
	if name := os.Getenv("LOOM_SESSION_SERVER_NAME"); name != "" {
		_ = os.Unsetenv("LOOM_SESSION_SERVER_NAME")
		return serveSession(ctx, name, options, command, args...)
	}
	return runSession(ctx, in, out, errOut, options, command, args...)
}

func openSessionPTY(cols, rows int) (master, slave *os.File, err error) {
	master, err = os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("open /dev/ptmx: %w", err)
	}
	fd := int(master.Fd())
	if err := unix.IoctlSetPointerInt(fd, unix.TIOCSPTLCK, 0); err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("unlock pty: %w", err)
	}
	n, err := unix.IoctlGetInt(fd, unix.TIOCGPTN)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("pty number: %w", err)
	}
	slave, err = os.OpenFile(fmt.Sprintf("/dev/pts/%d", n), os.O_RDWR, 0)
	if err != nil {
		master.Close()
		return nil, nil, fmt.Errorf("open slave: %w", err)
	}
	setPTYSize(master, cols, rows)
	return master, slave, nil
}

func setPTYSize(f *os.File, cols, rows int) {
	ws := &unix.Winsize{Col: uint16(cols), Row: uint16(rows)}
	_ = unix.IoctlSetWinsize(int(f.Fd()), unix.TIOCSWINSZ, ws)
}

type sessionTab struct {
	id            int
	title         string
	command       string
	args          []string
	master        *os.File
	child         *exec.Cmd
	vt            *vt.VT
	exitCode      int
	waitErr       error
	mouseOverride int // -1 disabled, 0 follows child, 1 enabled
	done          chan struct{}
	debugRing     *debugRing
	debugDetector *debugDetector
	debugAlert    bool
}

func drawSessionTabBar(out io.Writer, tabs []*sessionTab, activeIndex, cols, rows int) {
	defs := loom.SpeccedDefaults.Session
	if !sessionBarVisible(defs, len(tabs)) || cols <= 0 {
		return
	}
	var b strings.Builder
	b.WriteString("\x1b7") // DEC save cursor & attributes
	fmt.Fprintf(&b, "\x1b[%d;1H", rows)

	barCols := cols - 1 // avoid terminal's pending-wrap behavior on the final cell
	if barCols < 0 {
		barCols = 0
	}
	bar, _, _ := sessionBarContent(tabs, activeIndex, barCols, activeIndex >= 0 && activeIndex < len(tabs) && sessionTabMouseOn(tabs[activeIndex]))
	b.WriteString(bar)
	b.WriteString("\x1b[0m\x1b8") // reset style and restore cursor
	_, _ = out.Write([]byte(b.String()))
}

func sessionBarVisible(defs loom.SessionDefaults, tabCount int) bool {
	return defs.Bar == "always" || (defs.Bar == "tabs" && tabCount > 1)
}

func sessionChildRows(defs loom.SessionDefaults, tabCount, rows int) int {
	if sessionBarVisible(defs, tabCount) && rows > 1 {
		return rows - 1
	}
	return rows
}

func enterSessionOuterAlt(out io.Writer, rows int) {
	_, _ = fmt.Fprintf(out, "\x1b[r\x1b[%d;1H\x1b[2K\x1b[?1049h\x1b[2J\x1b[H", rows)
}

type sessionBarTabHit struct{ start, end, index int }

func sessionBarContent(tabs []*sessionTab, activeIndex, cols int, mouseOn bool) (string, int, []sessionBarTabHit) {
	defs := loom.SpeccedDefaults.Session
	labels := defs.BarLabels
	hintStyle := loom.DefaultTheme().HintBarStyle()
	capStyle, labelStyle := hintStyle.Cap.ANSI(), hintStyle.Label.ANSI()
	mouseStyle := labelStyle
	if mouseOn {
		mouseStyle = capStyle
	}
	icon := labels.Separator + labels.MouseIcon
	iconWidth := loom.StringWidth(icon)
	alert := activeIndex >= 0 && activeIndex < len(tabs) && tabs[activeIndex].debugAlert
	alertWidth := 0
	if alert {
		alertWidth = loom.StringWidth("[!] ")
	}
	searchKey := loom.KeyCap(defs.Bindings.Search[0])
	widths := []int{loom.StringWidth(loom.KeyCap(defs.Prefix)+"…") + 4 + loom.StringWidth(labels.Tools), loom.StringWidth(loom.KeyCap(defs.PrefixBindings.Help[0])) + 4 + loom.StringWidth(labels.Help), loom.StringWidth(loom.KeyCap(defs.Bindings.Screenshot[0])) + 4 + loom.StringWidth(labels.Screenshot), loom.StringWidth(loom.KeyCap(defs.Bindings.Quit[0])) + 4 + loom.StringWidth(labels.Quit), loom.StringWidth(searchKey) + 4 + loom.StringWidth(labels.Search)}
	hints := []string{sessionHint(capStyle, labelStyle, loom.KeyCap(defs.Prefix)+"…", labels.Tools), sessionHint(capStyle, labelStyle, loom.KeyCap(defs.PrefixBindings.Help[0]), labels.Help), sessionHint(capStyle, labelStyle, loom.KeyCap(defs.Bindings.Screenshot[0]), labels.Screenshot), sessionHint(capStyle, labelStyle, loom.KeyCap(defs.Bindings.Quit[0]), labels.Quit), sessionHint(capStyle, labelStyle, searchKey, labels.Search)}
	visible := []bool{true, true, true, true, true}
	tabWidth := 0
	for i, tab := range tabs {
		sep := 0
		if i > 0 {
			sep = 1
		}
		tabWidth += sep + loom.StringWidth(fmt.Sprintf(" %d: %s ", i+1, tab.title))
	}
	dropIndices := map[string]int{"tools": 0, "help": 1, "screenshot": 2, "quit": 3, "search": 4}
	for _, name := range labels.DropOrder {
		i := dropIndices[name]
		need := iconWidth + alertWidth
		for j, w := range widths {
			if visible[j] {
				need += w
			}
		}
		if tabWidth+need <= cols {
			break
		}
		visible[i] = false
	}
	usedHints := 0
	for i, w := range widths {
		if visible[i] {
			usedHints += w
		}
	}
	if usedHints > 0 {
		usedHints += 1
	}
	remainingTabs := cols - iconWidth - alertWidth - usedHints
	var b strings.Builder
	style := loom.DefaultTabsStyle()
	hits := []sessionBarTabHit{}
	x := 0
	for i, tab := range tabs {
		if i > 0 {
			if x >= remainingTabs {
				break
			}
			b.WriteString(" ")
			x++
		}
		text := fmt.Sprintf(" %d: %s ", i+1, tab.title)
		if x+loom.StringWidth(text) > remainingTabs {
			text = loom.TruncateText(text, remainingTabs-x, "…")
		}
		if text == "" {
			break
		}
		start := x
		tabStyle := style.Inactive
		if i == activeIndex {
			tabStyle = style.Active
		}
		b.WriteString(tabStyle.ANSI())
		b.WriteString(text)
		x += loom.StringWidth(text)
		hits = append(hits, sessionBarTabHit{start, x, i})
	}
	if x > 0 && usedHints > 0 {
		b.WriteString(strings.Repeat(" ", remainingTabs-x))
		x = remainingTabs
	}
	mouseStart := cols - iconWidth
	for i, hint := range hints {
		if visible[i] {
			b.WriteString(hint)
		}
	}
	if alert {
		b.WriteString(capStyle)
		b.WriteString("[!] ")
	}
	b.WriteString(mouseStyle)
	b.WriteString(icon)
	width := x + usedHints + iconWidth + alertWidth
	if width < cols {
		b.WriteString(labelStyle)
		b.WriteString(strings.Repeat(" ", cols-width))
	}
	return b.String(), mouseStart, hits
}

func sessionHint(capStyle, labelStyle, key, label string) string {
	return capStyle + " " + key + " " + labelStyle + " " + label + " "
}

func sessionTabMouseOn(tab *sessionTab) bool {
	return tab != nil && (tab.mouseOverride > 0 || (tab.mouseOverride == 0 && tab.vt.MouseMode != 0))
}

func syncSessionMouse(out io.Writer, tab *sessionTab) {
	if tab == nil || tab.mouseOverride == 0 {
		return
	}
	if tab.mouseOverride < 0 {
		_, _ = io.WriteString(out, "\x1b[?9l\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1005l\x1b[?1006l\x1b[?1015l")
		return
	}
	_, _ = io.WriteString(out, "\x1b[?9l\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1005l\x1b[?1006l\x1b[?1015l\x1b[?1000h\x1b[?1006h")
}

func sessionMouseIconRange(cols int) (int, int) {
	icon := loom.SpeccedDefaults.Session.BarLabels.Separator + loom.SpeccedDefaults.Session.BarLabels.MouseIcon
	return cols - 1 - loom.StringWidth(icon), cols - 1
}

func drawSessionHelp(out io.Writer, cols, rows, page int) int {
	d := loom.SpeccedDefaults.Session
	style := loom.DefaultTheme().HintBarStyle()
	entries := []struct {
		keys  []string
		label string
	}{
		{d.Bindings.Screenshot, d.BarLabels.Screenshot}, {d.Bindings.Quit, d.BarLabels.Quit}, {d.Bindings.NewTab, d.BarLabels.NewTab},
		{d.Bindings.Search, d.BarLabels.Search},
		{d.PrefixBindings.Help, d.BarLabels.Help}, {d.PrefixBindings.ToggleMouse, d.BarLabels.ToggleMouse}, {d.PrefixBindings.Screenshot, d.BarLabels.Screenshot},
		{d.PrefixBindings.Quit, d.BarLabels.Quit}, {d.PrefixBindings.SendPrefix, d.BarLabels.SendPrefix}, {d.PrefixBindings.NewTab, d.BarLabels.NewTab},
		{d.PrefixBindings.NextTab, d.BarLabels.NextTab}, {d.PrefixBindings.PreviousTab, d.BarLabels.PreviousTab}, {d.PrefixBindings.CloseTab, d.BarLabels.CloseTab},
		{d.PrefixBindings.Search, d.BarLabels.Search},
	}
	if rows <= 0 || cols <= 0 {
		return 1
	}
	labelText := func(entry struct {
		keys  []string
		label string
	}) (string, string) {
		keys := make([]string, len(entry.keys))
		for i, key := range entry.keys {
			keys[i] = loom.KeyCap(key)
		}
		return strings.Join(keys, ", "), entry.label
	}
	entryWidth := func(entry struct {
		keys  []string
		label string
	}) int {
		keys, label := labelText(entry)
		return loom.StringWidth(keys) + 1 + loom.StringWidth(label)
	}
	leftCount := (len(entries) + 1) / 2
	leftWidth, rightWidth := 0, 0
	for i, entry := range entries {
		width := entryWidth(entry)
		if i < leftCount {
			leftWidth = max(leftWidth, width)
		} else {
			rightWidth = max(rightWidth, width)
		}
	}
	twoColumns := leftWidth+rightWidth+3 <= cols && leftCount+4 <= rows
	var b strings.Builder
	b.WriteString("\x1b[0m\x1b[2J\x1b[H")
	fmt.Fprintf(&b, "%sSession tools%s\r\n\r\n", style.Cap.ANSI(), style.Label.ANSI())
	renderEntry := func(entry struct {
		keys  []string
		label string
	}) string {
		keys, label := labelText(entry)
		return fmt.Sprintf("%s%s%s %s%s", style.Cap.ANSI(), keys, style.Label.ANSI(), label, "\x1b[0m")
	}
	pages := 1
	if twoColumns {
		for i := 0; i < leftCount; i++ {
			left := entries[i]
			leftKeys, leftLabel := labelText(left)
			fmt.Fprintf(&b, "%s%s%s %s%s", style.Cap.ANSI(), leftKeys, style.Label.ANSI(), leftLabel, "\x1b[0m")
			if rightIndex := i + leftCount; rightIndex < len(entries) {
				padding := leftWidth - entryWidth(left) + 3
				b.WriteString(strings.Repeat(" ", padding))
				b.WriteString(renderEntry(entries[rightIndex]))
			}
			b.WriteString("\r\n")
		}
	} else {
		perPage := max(1, rows-4)
		pages = (len(entries) + perPage - 1) / perPage
		page = (page%pages + pages) % pages
		start := page * perPage
		end := min(len(entries), start+perPage)
		for _, entry := range entries[start:end] {
			fmt.Fprintf(&b, "%s%s\r\n", renderEntry(entry), "\x1b[0m")
		}
	}
	footer := "Press any key to return"
	if pages > 1 {
		footer = fmt.Sprintf("Any key: next · Esc: return (%d/%d)", page+1, pages)
	}
	fmt.Fprintf(&b, "\r\n%s%s%s\r\n", style.Label.ANSI(), footer, "\x1b[0m")
	_, _ = out.Write([]byte(b.String()))
	return pages
}

func tabIndexAtCol(tabs []*sessionTab, col, cols int) int {
	barCols := cols - 1
	_, _, hits := sessionBarContent(tabs, -1, barCols, false)
	for _, hit := range hits {
		if col >= hit.start && col < hit.end {
			return hit.index
		}
	}
	return -1
}

func applyTabModes(out io.Writer, oldTab, newTab *sessionTab) {
	if out == nil || newTab == nil {
		return
	}
	var b strings.Builder

	oldMouse := 0
	oldSGR := false
	oldUTF8 := false
	oldURXVT := false
	if oldTab != nil {
		oldMouse = oldTab.vt.MouseMode
		oldSGR = oldTab.vt.MouseSGR
		oldUTF8 = oldTab.vt.MouseUTF8
		oldURXVT = oldTab.vt.MouseURXVT
	}
	newMouse := newTab.vt.MouseMode
	newSGR := newTab.vt.MouseSGR
	newUTF8 := newTab.vt.MouseUTF8
	newURXVT := newTab.vt.MouseURXVT

	if oldMouse != newMouse || oldSGR != newSGR || oldUTF8 != newUTF8 || oldURXVT != newURXVT {
		if oldMouse != 0 || oldSGR || oldUTF8 || oldURXVT {
			b.WriteString("\x1b[?9l\x1b[?1000l\x1b[?1002l\x1b[?1003l\x1b[?1005l\x1b[?1006l\x1b[?1015l")
		}
		if newMouse != 0 {
			fmt.Fprintf(&b, "\x1b[?%dh", newMouse)
			if newSGR {
				b.WriteString("\x1b[?1006h")
			}
			if newUTF8 {
				b.WriteString("\x1b[?1005h")
			}
			if newURXVT {
				b.WriteString("\x1b[?1015h")
			}
		}
	}

	oldFocus := false
	if oldTab != nil {
		oldFocus = oldTab.vt.FocusEvents
	}
	if oldFocus != newTab.vt.FocusEvents {
		if newTab.vt.FocusEvents {
			b.WriteString("\x1b[?1004h")
		} else {
			b.WriteString("\x1b[?1004l")
		}
	}

	oldPaste := false
	if oldTab != nil {
		oldPaste = oldTab.vt.BracketedPaste
	}
	if oldPaste != newTab.vt.BracketedPaste {
		if newTab.vt.BracketedPaste {
			b.WriteString("\x1b[?2004h")
		} else {
			b.WriteString("\x1b[?2004l")
		}
	}

	oldCK := false
	if oldTab != nil {
		oldCK = oldTab.vt.CursorKeysApp
	}
	if oldCK != newTab.vt.CursorKeysApp {
		if newTab.vt.CursorKeysApp {
			b.WriteString("\x1b[?1h")
		} else {
			b.WriteString("\x1b[?1l")
		}
	}

	oldKeypad := false
	if oldTab != nil {
		oldKeypad = oldTab.vt.KeypadApp
	}
	if oldKeypad != newTab.vt.KeypadApp {
		if newTab.vt.KeypadApp {
			b.WriteString("\x1b=")
		} else {
			b.WriteString("\x1b>")
		}
	}

	oldKitty := false
	if oldTab != nil {
		oldKitty = oldTab.vt.KittyKeyboard
	}
	if oldKitty != newTab.vt.KittyKeyboard {
		if newTab.vt.KittyKeyboard {
			b.WriteString("\x1b[>1u")
		} else {
			b.WriteString("\x1b[<u")
		}
	}

	if b.Len() > 0 {
		_, _ = out.Write([]byte(b.String()))
	}
	syncSessionMouse(out, newTab)
}

func repaintActiveTab(out io.Writer, tabs []*sessionTab, activeIndex, cols, rows int, outerAlt bool) {
	if activeIndex < 0 || activeIndex >= len(tabs) {
		return
	}
	tab := tabs[activeIndex]
	repaint := tab.vt.Repaint()
	if outerAlt {
		// When outer alternate screen is active, keep the terminal in alternate screen
		repaint = strings.ReplaceAll(repaint, "\x1b[?1049l", "")
	}
	_, _ = out.Write([]byte(repaint))

	if sessionBarVisible(loom.SpeccedDefaults.Session, len(tabs)) {
		fmt.Fprintf(out, "\x1b[1;%dr", rows-1)
		drawSessionTabBar(out, tabs, activeIndex, cols, rows)
		x, y := tab.vt.Cursor()
		fmt.Fprintf(out, "\x1b[%d;%dH", y+1, x+1)
		if tab.vt.CursorVisible {
			_, _ = out.Write([]byte("\x1b[?25h"))
		} else {
			_, _ = out.Write([]byte("\x1b[?25l"))
		}
	} else {
		_, _ = out.Write([]byte("\x1b[r"))
	}
}

func switchActiveTab(out io.Writer, oldTab, newTab *sessionTab, tabs []*sessionTab, activeIndex, cols, rows int, outerAlt bool) {
	applyTabModes(out, oldTab, newTab)
	repaintActiveTab(out, tabs, activeIndex, cols, rows, outerAlt)
}

func runSession(ctx context.Context, in io.Reader, out io.Writer, errOut io.Writer, options Options, command string, args ...string) error {
	return runSessionCore(ctx, in, out, errOut, options, nil, command, args...)
}

func runSessionWithServer(ctx context.Context, server *sessionServer, options Options, command string, args ...string) error {
	err := runSessionCore(ctx, server.inputR, server.output, io.Discard, options, server, command, args...)
	var exitCode int
	var exitErr ExitError
	if errors.As(err, &exitErr) {
		exitCode = exitErr.ExitCode()
	}
	_ = server.writeMessage(ipcMessage{Type: msgExit, Code: exitCode})
	return err
}

func runSessionCore(ctx context.Context, in io.Reader, out io.Writer, errOut io.Writer, options Options, server *sessionServer, command string, args ...string) error {
	if command == "" {
		command = os.Getenv("SHELL")
		if command == "" {
			command = "/bin/sh"
		}
	}

	cols, rows := 80, 24
	var oldState *term.State
	if inFile, ok := in.(*os.File); ok && term.IsTerminal(int(inFile.Fd())) {
		if w, h, err := term.GetSize(int(inFile.Fd())); err == nil && w > 0 && h > 0 {
			cols, rows = w, h
		}
		state, err := term.MakeRaw(int(inFile.Fd()))
		if err == nil {
			oldState = state
		}
	} else if outFile, ok := out.(*os.File); ok && term.IsTerminal(int(outFile.Fd())) {
		if w, h, err := term.GetSize(int(outFile.Fd())); err == nil && w > 0 && h > 0 {
			cols, rows = w, h
		}
	}

	inFD := -1
	if inFile, ok := in.(*os.File); ok {
		inFD = int(inFile.Fd())
	}
	restorer := terminal.NewRestorer(inFD, oldState, out)
	defer restorer.Restore()
	defer func() {
		if r := recover(); r != nil {
			restorer.Restore()
			panic(r)
		}
	}()

	outerAltActive := false
	helpActive := false
	helpPage := 0
	helpPages := 1
	searchActive := false
	searchBar := loom.NewSearchBar()
	searchBar.Prompt = loom.SpeccedDefaults.Session.Search.Prompt
	searchBar.Placeholder = loom.SpeccedDefaults.Session.Search.Placeholder
	searchBar.SetFocused(true)
	var searchRows [][]vt.Cell
	var searchMatches []sessionMatch
	searchMatchIndex := -1
	var sessionMu sync.Mutex
	var tabs []*sessionTab
	activeIndex := 0
	barDrawn := false
	barMouseState := false
	barAlertState := false
	nextTabID := 1
	lastDebugAction := time.Time{}

	tabExitChan := make(chan *sessionTab, 16)
	sessionQuitChan := make(chan struct{})

	triggerDebugRepair := func(t *sessionTab, anomaly debugAnomaly) {
		if !options.Debug || t == nil || t.debugRing == nil || t.debugDetector == nil {
			return
		}
		now := time.Now()
		if !lastDebugAction.IsZero() && now.Sub(lastDebugAction) < debugRepairCooldown {
			return
		}
		lastDebugAction = now
		before := t.vt.ANSI()
		t.vt.ResetParser()
		t.debugDetector.pending = nil
		t.debugDetector.pendingAt = time.Time{}
		t.debugAlert = true
		childRows := t.vt.Rows
		toggleRows := childRows + 1
		if childRows > 1 {
			toggleRows = childRows - 1
		}
		setPTYSize(t.master, cols, toggleRows)
		setPTYSize(t.master, cols, childRows)
		if len(tabs) > 0 && activeIndex < len(tabs) && tabs[activeIndex] == t && sessionBarVisible(loom.SpeccedDefaults.Session, len(tabs)) {
			drawSessionTabBar(out, tabs, activeIndex, cols, rows)
			barDrawn = true
			barAlertState = true
		}
		go func(tab *sessionTab, reason debugAnomaly, captured string, started time.Time) {
			time.Sleep(debugCaptureDelay)
			sessionMu.Lock()
			after := tab.vt.ANSI()
			ring := tab.debugRing.Snapshot()
			sessionMu.Unlock()
			dir, dirErr := sessionDebugDirectory(options.DebugDir)
			if dirErr == nil {
				dirErr = writeDebugBundle(dir, tab.id, string(reason), ring, captured, after, started)
			}
			if dirErr != nil {
				fmt.Fprintf(errOut, "session debug: %v\n", dirErr)
			}
		}(t, anomaly, before, now)
	}

	restorer.OnRestore(func() {
		sessionMu.Lock()
		for _, t := range tabs {
			if t.vt != nil && t.vt.AltScreen() {
				restorer.SetAltScreen(true)
				break
			}
		}
		sessionMu.Unlock()
		eraseBar := fmt.Sprintf("\x1b7\x1b[%d;1H\x1b[2K\x1b8", rows)
		_, _ = out.Write([]byte("\x1b[r" + eraseBar))
	})

	spawnTab := func(cmdName string, cmdArgs ...string) (*sessionTab, error) {
		childRows := sessionChildRows(loom.SpeccedDefaults.Session, len(tabs)+1, rows)
		master, slave, err := openSessionPTY(cols, childRows)
		if err != nil {
			return nil, err
		}
		child := exec.CommandContext(ctx, cmdName, cmdArgs...)
		child.Stdin = slave
		child.Stdout = slave
		child.Stderr = slave
		child.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true, Ctty: 0}
		if err := child.Start(); err != nil {
			master.Close()
			slave.Close()
			return nil, fmt.Errorf("session start %s: %w", cmdName, err)
		}
		slave.Close()

		baseCmd := filepath.Base(cmdName)
		tab := &sessionTab{
			id:      nextTabID,
			title:   baseCmd,
			command: cmdName,
			args:    cmdArgs,
			master:  master,
			child:   child,
			vt:      vt.NewVT(cols, childRows),
			done:    make(chan struct{}),
		}
		if options.Debug {
			tab.debugRing = newDebugRing()
			tab.debugDetector = &debugDetector{}
		}
		tab.vt.SetScrollbackLimit(loom.SpeccedDefaults.Session.ScrollbackLines)
		nextTabID++

		go func(t *sessionTab) {
			defer close(t.done)
			buf := make([]byte, 4096)
			for {
				n, readErr := t.master.Read(buf)
				if n > 0 {
					sessionMu.Lock()
					var anomaly debugAnomaly
					if t.debugRing != nil {
						t.debugRing.Write(buf[:n])
						anomaly = t.debugDetector.Observe(buf[:n], time.Now())
					}
					frame, _ := t.vt.WriteFramed(buf[:n])
					if anomaly == "" && t.debugDetector != nil && suspiciousDebugText(t.vt.Text()) {
						anomaly = debugLeak
					}
					if anomaly != "" {
						triggerDebugRepair(t, anomaly)
					}
					if t.vt.KittyKeyboard {
						restorer.SetKittyKeyboard(true)
					}
					if t.vt.AltScreen() {
						restorer.SetAltScreen(true)
					}
					if len(tabs) > 0 && activeIndex < len(tabs) && tabs[activeIndex] == t && !helpActive && !searchActive {
						_, _ = out.Write(frame)
						syncSessionMouse(out, t)
						if sessionBarVisible(loom.SpeccedDefaults.Session, len(tabs)) {
							mouseState := sessionTabMouseOn(t)
							if len(tabs) > 1 || !barDrawn || mouseState != barMouseState || t.debugAlert != barAlertState {
								drawSessionTabBar(out, tabs, activeIndex, cols, rows)
								barDrawn = true
								barMouseState = mouseState
								barAlertState = t.debugAlert
							}
						}
					}
					sessionMu.Unlock()
				}
				if readErr != nil {
					break
				}
			}
			waitErr := t.child.Wait()
			var exitCode int
			if waitErr != nil {
				var exitErr *exec.ExitError
				if errors.As(waitErr, &exitErr) {
					exitCode = exitErr.ExitCode()
				}
			}
			t.exitCode = exitCode
			t.waitErr = waitErr
			tabExitChan <- t
		}(tab)

		return tab, nil
	}

	initialTab, err := spawnTab(command, args...)
	if err != nil {
		return err
	}
	tabs = append(tabs, initialTab)
	if sessionBarVisible(loom.SpeccedDefaults.Session, len(tabs)) {
		fmt.Fprintf(out, "\x1b[1;%dr", rows-1)
		drawSessionTabBar(out, tabs, activeIndex, cols, rows)
	}

	sigWinch := make(chan os.Signal, 1)
	signal.Notify(sigWinch, syscall.SIGWINCH)
	defer signal.Stop(sigWinch)

	sigTerm := make(chan os.Signal, 2)
	signal.Notify(sigTerm, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(sigTerm)
	go func() {
		for range sigWinch {
			sessionMu.Lock()
			if inFile, ok := in.(*os.File); ok && term.IsTerminal(int(inFile.Fd())) {
				if w, h, err := term.GetSize(int(inFile.Fd())); err == nil && w > 0 && h > 0 {
					cols, rows = w, h
					childRows := sessionChildRows(loom.SpeccedDefaults.Session, len(tabs), rows)
					for _, t := range tabs {
						setPTYSize(t.master, cols, childRows)
						t.vt.Resize(cols, childRows)
					}
					if helpActive {
						helpPages = drawSessionHelp(out, cols, rows, helpPage)
					} else if searchActive && activeIndex >= 0 && activeIndex < len(tabs) {
						var selected *sessionMatch
						if searchMatchIndex >= 0 && searchMatchIndex < len(searchMatches) {
							selected = &searchMatches[searchMatchIndex]
						}
						drawSessionSearchView(out, searchRows, tabs[activeIndex].vt.Rows, cols, selected)
						drawSessionSearchBar(out, cols, rows, searchBar, sessionSearchControls(loom.SpeccedDefaults.Session, searchBar.Query, searchMatchIndex, len(searchMatches)))
					} else {
						repaintActiveTab(out, tabs, activeIndex, cols, rows, outerAltActive)
					}
				}
			}
			sessionMu.Unlock()
		}
	}()

	var savedShots []string
	var shotErrors []error
	var shotMu sync.Mutex
	saveScreenshot := func() error {
		sessionMu.Lock()
		if len(tabs) == 0 || activeIndex >= len(tabs) {
			sessionMu.Unlock()
			return errors.New("no active tab for screenshot")
		}
		curTab := tabs[activeIndex]
		content := curTab.vt.ANSI()
		cmdName := curTab.command
		sessionMu.Unlock()

		baseCmd := filepath.Base(cmdName)
		safeCmd := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				return r
			}
			return '-'
		}, baseCmd)
		if safeCmd == "" || safeCmd == "-" {
			safeCmd = "session"
		}

		timestamp := time.Now().Format("20060102-150405")
		path := fmt.Sprintf("./loom-%s-%s.ansi", safeCmd, timestamp)
		for i := 1; ; i++ {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
			if err == nil {
				_, writeErr := f.WriteString(content)
				closeErr := f.Close()
				if writeErr != nil {
					return writeErr
				}
				if closeErr != nil {
					return closeErr
				}
				shotMu.Lock()
				savedShots = append(savedShots, path)
				shotMu.Unlock()
				return nil
			}
			if os.IsExist(err) {
				path = fmt.Sprintf("./loom-%s-%s-%d.ansi", safeCmd, timestamp, i)
				continue
			}
			return err
		}
	}

	drawSearch := func() {
		if activeIndex < 0 || activeIndex >= len(tabs) {
			return
		}
		var selected *sessionMatch
		if searchMatchIndex >= 0 && searchMatchIndex < len(searchMatches) {
			selected = &searchMatches[searchMatchIndex]
		}
		drawSessionSearchView(out, searchRows, tabs[activeIndex].vt.Rows, cols, selected)
		drawSessionSearchBar(out, cols, rows, searchBar, sessionSearchControls(loom.SpeccedDefaults.Session, searchBar.Query, searchMatchIndex, len(searchMatches)))
	}

	updateSearch := func() {
		searchMatches = findSessionMatches(searchRows, searchBar.Query)
		searchMatchIndex = -1
		if len(searchMatches) > 0 {
			searchMatchIndex = len(searchMatches) - 1
		}
		drawSearch()
	}

	openSearch := func() {
		searchActive = true
		searchBar.Clear()
		searchBar.SetFocused(true)
		searchRows = tabs[activeIndex].vt.ScrollbackCells()
		searchMatches = nil
		searchMatchIndex = -1
		drawSearch()
	}

	stdinDone := make(chan struct{})
	go func() {
		defer close(stdinDone)
		buf := make([]byte, 2048)
		var pending []byte
		prefixMode := false

		sessionDefs := loom.SpeccedDefaults.Session

		for {
			n, err := in.Read(buf)
			if n > 0 {
				pending = append(pending, buf[:n]...)

				for len(pending) > 0 {
					sessionMu.Lock()
					if len(tabs) == 0 {
						sessionMu.Unlock()
						return
					}
					activeMaster := tabs[activeIndex].master

					// Lone Esc
					if len(pending) == 1 && pending[0] == 0x1b {
						prefixMode = false
						if searchActive {
							searchActive = false
							searchBar.SetFocused(false)
							repaintActiveTab(out, tabs, activeIndex, cols, rows, outerAltActive)
						} else {
							_, _ = activeMaster.Write(pending)
						}
						pending = nil
						sessionMu.Unlock()
						break
					}

					// Bracketed paste
					if paste, used, ok := loom.DecodePaste(pending); ok {
						if searchActive {
							searchBar.SetQuery(searchBar.Query + paste.Text)
							updateSearch()
						} else {
							_, _ = activeMaster.Write(pending[:used])
						}
						pending = pending[used:]
						prefixMode = false
						sessionMu.Unlock()
						continue
					}

					// SGR Mouse event (\x1b[<...M or \x1b[<...m)
					if ev, used, ok := loom.ScanMouse(pending); ok {
						if searchActive {
							pending = pending[used:]
							sessionMu.Unlock()
							continue
						}
						if sessionBarVisible(sessionDefs, len(tabs)) && ev.Y == rows {
							if ev.Action == loom.MousePress && ev.Button == loom.MouseLeft {
								iconStart, iconEnd := sessionMouseIconRange(cols)
								if sessionTabMouseOn(tabs[activeIndex]) && ev.X-1 >= iconStart && ev.X-1 < iconEnd {
									tabs[activeIndex].mouseOverride = -1
									syncSessionMouse(out, tabs[activeIndex])
									drawSessionTabBar(out, tabs, activeIndex, cols, rows)
								} else {
									target := tabIndexAtCol(tabs, ev.X-1, cols)
									if target >= 0 && target < len(tabs) && target != activeIndex {
										oldTab := tabs[activeIndex]
										activeIndex = target
										newTab := tabs[activeIndex]
										switchActiveTab(out, oldTab, newTab, tabs, activeIndex, cols, rows, outerAltActive)
									}
								}
							}
							pending = pending[used:]
							prefixMode = false
							sessionMu.Unlock()
							continue
						}
						_, _ = activeMaster.Write(pending[:used])
						pending = pending[used:]
						prefixMode = false
						sessionMu.Unlock()
						continue
					}

					// Standard X10 mouse event (\x1b[M + 3 bytes)
					if len(pending) >= 6 && pending[0] == 0x1b && pending[1] == '[' && pending[2] == 'M' {
						if searchActive {
							pending = pending[6:]
							sessionMu.Unlock()
							continue
						}
						col := int(pending[4]) - 32
						row := int(pending[5]) - 32
						btn := int(pending[3]) - 32
						if sessionBarVisible(sessionDefs, len(tabs)) && row == rows {
							if btn&3 == 0 {
								iconStart, iconEnd := sessionMouseIconRange(cols)
								if sessionTabMouseOn(tabs[activeIndex]) && col-1 >= iconStart && col-1 < iconEnd {
									tabs[activeIndex].mouseOverride = -1
									syncSessionMouse(out, tabs[activeIndex])
									drawSessionTabBar(out, tabs, activeIndex, cols, rows)
								} else {
									target := tabIndexAtCol(tabs, col-1, cols)
									if target >= 0 && target < len(tabs) && target != activeIndex {
										oldTab := tabs[activeIndex]
										activeIndex = target
										newTab := tabs[activeIndex]
										switchActiveTab(out, oldTab, newTab, tabs, activeIndex, cols, rows, outerAltActive)
									}
								}
							}
							pending = pending[6:]
							prefixMode = false
							sessionMu.Unlock()
							continue
						}
						_, _ = activeMaster.Write(pending[:6])
						pending = pending[6:]
						prefixMode = false
						sessionMu.Unlock()
						continue
					}

					key, used, ok := loom.ScanKey(pending)
					if !ok {
						sessionMu.Unlock()
						break
					}

					if searchActive {
						if key.Is(sessionDefs.Bindings.Quit...) {
							sessionMu.Unlock()
							close(sessionQuitChan)
							return
						}
						switch {
						case key.Is(sessionDefs.Search.Bindings.Close...):
							searchActive = false
							searchBar.SetFocused(false)
							repaintActiveTab(out, tabs, activeIndex, cols, rows, outerAltActive)
						case key.Is(sessionDefs.Search.Bindings.Next...):
							if len(searchMatches) > 0 {
								searchMatchIndex = (searchMatchIndex - 1 + len(searchMatches)) % len(searchMatches)
							}
							drawSearch()
						case key.Is(sessionDefs.Search.Bindings.Previous...):
							if len(searchMatches) > 0 {
								searchMatchIndex = (searchMatchIndex + 1) % len(searchMatches)
							}
							drawSearch()
						default:
							searchBar.ConsumeKey(key)
							updateSearch()
						}
						prefixMode = false
						sessionMu.Unlock()
						pending = pending[used:]
						continue
					}

					if prefixMode {
						switch {
						case key.Is(sessionDefs.PrefixBindings.ToggleMouse...):
							tab := tabs[activeIndex]
							if sessionTabMouseOn(tab) {
								tab.mouseOverride = -1
							} else {
								tab.mouseOverride = 1
							}
							syncSessionMouse(out, tab)
							drawSessionTabBar(out, tabs, activeIndex, cols, rows)
							prefixMode = false
							sessionMu.Unlock()
						case key.Is(sessionDefs.PrefixBindings.Search...):
							prefixMode = false
							openSearch()
							sessionMu.Unlock()
						case key.Is(sessionDefs.PrefixBindings.Help...):
							helpActive = true
							prefixMode = false
							helpPage = 0
							helpPages = drawSessionHelp(out, cols, rows, helpPage)
							sessionMu.Unlock()
						case key.Is(sessionDefs.PrefixBindings.Screenshot...):
							prefixMode = false
							sessionMu.Unlock()
							if err := saveScreenshot(); err != nil {
								shotMu.Lock()
								shotErrors = append(shotErrors, err)
								shotMu.Unlock()
							}
						case key.Is(sessionDefs.PrefixBindings.Quit...):
							sessionMu.Unlock()
							close(sessionQuitChan)
							return
						case key.Is(sessionDefs.PrefixBindings.Detach...):
							prefixMode = false
							sessionMu.Unlock()
							if server != nil {
								server.detach(true)
							}
						case key.Is(sessionDefs.PrefixBindings.SendPrefix...):
							_, _ = activeMaster.Write(pending[:used])
							prefixMode = false
							sessionMu.Unlock()
						case key.Is(sessionDefs.PrefixBindings.NewTab...):
							shell := os.Getenv("SHELL")
							if shell == "" {
								shell = "/bin/sh"
							}
							if len(tabs) == 1 {
								if !outerAltActive {
									outerAltActive = true
									restorer.SetAltScreen(true)
									enterSessionOuterAlt(out, rows)
								}
								childRows := sessionChildRows(sessionDefs, len(tabs)+1, rows)
								setPTYSize(tabs[0].master, cols, childRows)
								tabs[0].vt.Resize(cols, childRows)
							}
							oldTab := tabs[activeIndex]
							newTab, err := spawnTab(shell)
							if err == nil {
								tabs = append(tabs, newTab)
								activeIndex = len(tabs) - 1
								switchActiveTab(out, oldTab, newTab, tabs, activeIndex, cols, rows, outerAltActive)
							}
							prefixMode = false
							sessionMu.Unlock()
						case key.Is(sessionDefs.PrefixBindings.NextTab...):
							if sessionBarVisible(sessionDefs, len(tabs)) {
								oldTab := tabs[activeIndex]
								activeIndex = (activeIndex + 1) % len(tabs)
								newTab := tabs[activeIndex]
								switchActiveTab(out, oldTab, newTab, tabs, activeIndex, cols, rows, outerAltActive)
							}
							prefixMode = false
							sessionMu.Unlock()
						case key.Is(sessionDefs.PrefixBindings.PreviousTab...):
							if sessionBarVisible(sessionDefs, len(tabs)) {
								oldTab := tabs[activeIndex]
								activeIndex = (activeIndex - 1 + len(tabs)) % len(tabs)
								newTab := tabs[activeIndex]
								switchActiveTab(out, oldTab, newTab, tabs, activeIndex, cols, rows, outerAltActive)
							}
							prefixMode = false
							sessionMu.Unlock()
						case key.Is(sessionDefs.PrefixBindings.CloseTab...):
							if len(tabs) > 0 {
								t := tabs[activeIndex]
								killProcessGroup(t.child.Process.Pid)
								_ = t.master.Close()
							}
							prefixMode = false
							sessionMu.Unlock()
						default:
							// Check select_tab bindings by index
							selected := -1
							for idx, binding := range sessionDefs.PrefixBindings.SelectTab {
								if key.Is(binding) {
									selected = idx
									break
								}
							}
							if selected >= 0 {
								if selected < len(tabs) && selected != activeIndex {
									oldTab := tabs[activeIndex]
									activeIndex = selected
									newTab := tabs[activeIndex]
									switchActiveTab(out, oldTab, newTab, tabs, activeIndex, cols, rows, outerAltActive)
								}
							} else {
								_, _ = activeMaster.Write(pending[:used])
							}
							prefixMode = false
							sessionMu.Unlock()
						}
						pending = pending[used:]
						continue
					}
					if helpActive {
						prefixMode = false
						if key.Is("esc") {
							helpActive = false
							repaintActiveTab(out, tabs, activeIndex, cols, rows, outerAltActive)
						} else if helpPages > 1 {
							helpPage = (helpPage + 1) % helpPages
							helpPages = drawSessionHelp(out, cols, rows, helpPage)
						} else {
							helpActive = false
							repaintActiveTab(out, tabs, activeIndex, cols, rows, outerAltActive)
						}
						sessionMu.Unlock()
						pending = pending[used:]
						continue
					}

					// Direct hotkey matching
					selectIdx := -1
					for idx, binding := range sessionDefs.Bindings.SelectTab {
						if key.Is(binding) {
							selectIdx = idx
							break
						}
					}
					if selectIdx >= 0 {
						if selectIdx < len(tabs) && selectIdx != activeIndex {
							oldTab := tabs[activeIndex]
							activeIndex = selectIdx
							newTab := tabs[activeIndex]
							switchActiveTab(out, oldTab, newTab, tabs, activeIndex, cols, rows, outerAltActive)
						}
						sessionMu.Unlock()
						pending = pending[used:]
						continue
					}

					switch {
					case key.Is(sessionDefs.Bindings.Search...):
						openSearch()
						sessionMu.Unlock()
					case key.Is(sessionDefs.Bindings.Screenshot...):
						sessionMu.Unlock()
						if err := saveScreenshot(); err != nil {
							shotMu.Lock()
							shotErrors = append(shotErrors, err)
							shotMu.Unlock()
						}
					case key.Is(sessionDefs.Bindings.Quit...):
						sessionMu.Unlock()
						close(sessionQuitChan)
						return
					case key.Is(sessionDefs.Bindings.NewTab...):
						shell := os.Getenv("SHELL")
						if shell == "" {
							shell = "/bin/sh"
						}
						if len(tabs) == 1 {
							if !outerAltActive {
								outerAltActive = true
								restorer.SetAltScreen(true)
								enterSessionOuterAlt(out, rows)
							}
							childRows := sessionChildRows(sessionDefs, len(tabs)+1, rows)
							setPTYSize(tabs[0].master, cols, childRows)
							tabs[0].vt.Resize(cols, childRows)
						}
						oldTab := tabs[activeIndex]
						newTab, err := spawnTab(shell)
						if err == nil {
							tabs = append(tabs, newTab)
							activeIndex = len(tabs) - 1
							switchActiveTab(out, oldTab, newTab, tabs, activeIndex, cols, rows, outerAltActive)
						}
						sessionMu.Unlock()
					case key.Is(sessionDefs.Prefix):
						prefixMode = true
						sessionMu.Unlock()
					default:
						_, _ = activeMaster.Write(pending[:used])
						sessionMu.Unlock()
					}
					pending = pending[used:]
				}
			}
			if err != nil {
				return
			}
		}
	}()

	debugTick := time.NewTicker(100 * time.Millisecond)
	defer debugTick.Stop()
	var finalExitCode int
	var finalWaitErr error

	for {
		select {
		case <-ctx.Done():
			sessionMu.Lock()
			for _, t := range tabs {
				killProcessGroup(t.child.Process.Pid)
			}
			sessionMu.Unlock()
			return ctx.Err()
		case <-sessionQuitChan:
			sessionMu.Lock()
			for _, t := range tabs {
				killProcessGroup(t.child.Process.Pid)
			}
			sessionMu.Unlock()
			goto done
		case <-serverKillEvents(server):
			sessionMu.Lock()
			for _, t := range tabs {
				killProcessGroup(t.child.Process.Pid)
			}
			sessionMu.Unlock()
			goto done
		case attached := <-serverAttachEvents(server):
			sessionMu.Lock()
			if attached.cols > 0 && attached.rows > 0 && (attached.cols != cols || attached.rows != rows) {
				cols, rows = attached.cols, attached.rows
				childRows := sessionChildRows(loom.SpeccedDefaults.Session, len(tabs), rows)
				for _, tab := range tabs {
					setPTYSize(tab.master, cols, childRows)
					tab.vt.Resize(cols, childRows)
				}
			}
			if activeIndex >= 0 && activeIndex < len(tabs) {
				active := tabs[activeIndex]
				if server != nil && attached.attach {
					_ = server.writeMessage(ipcMessage{Type: msgStatus, Status: "modes", AltScreen: outerAltActive, KittyKeyboard: active.vt.KittyKeyboard})
				}
				repaintOut := out
				if server != nil && attached.attach {
					repaintOut = &serverOutput{server: server, direct: true}
				}
				repaintActiveTab(repaintOut, tabs, activeIndex, cols, rows, outerAltActive)
				if server != nil && attached.attach {
					server.markReady()
				}
			}
			sessionMu.Unlock()
		case now := <-debugTick.C:
			sessionMu.Lock()
			for _, tab := range tabs {
				if tab.debugDetector == nil {
					continue
				}
				if anomaly := tab.debugDetector.Stalled(now); anomaly != "" {
					triggerDebugRepair(tab, anomaly)
				}
				if tab.debugAlert && !lastDebugAction.IsZero() && now.Sub(lastDebugAction) >= debugRepairCooldown {
					tab.debugAlert = false
					if len(tabs) > 0 && tabs[activeIndex] == tab && sessionBarVisible(loom.SpeccedDefaults.Session, len(tabs)) {
						drawSessionTabBar(out, tabs, activeIndex, cols, rows)
						barAlertState = false
					}
				}
			}
			sessionMu.Unlock()
		case sig := <-sigTerm:
			sessionMu.Lock()
			for _, t := range tabs {
				killProcessGroup(t.child.Process.Pid)
			}
			sessionMu.Unlock()
			if unixSig, ok := sig.(syscall.Signal); ok {
				finalExitCode = 128 + int(unixSig)
			} else {
				finalExitCode = 143
			}
			goto done
		case exitedTab := <-tabExitChan:
			sessionMu.Lock()
			idx := -1
			for i, t := range tabs {
				if t == exitedTab {
					idx = i
					break
				}
			}
			if idx >= 0 {
				exitedTab.master.Close()
				tabs = append(tabs[:idx], tabs[idx+1:]...)
				if len(tabs) == 0 {
					finalExitCode = exitedTab.exitCode
					finalWaitErr = exitedTab.waitErr
					sessionMu.Unlock()
					goto done
				}
				if activeIndex >= len(tabs) {
					activeIndex = len(tabs) - 1
				}
				if len(tabs) == 1 {
					childRows := sessionChildRows(loom.SpeccedDefaults.Session, len(tabs), rows)
					setPTYSize(tabs[0].master, cols, childRows)
					tabs[0].vt.Resize(cols, childRows)
					eraseBar := fmt.Sprintf("\x1b7\x1b[%d;1H\x1b[2K\x1b8", rows)
					_, _ = out.Write([]byte("\x1b[r" + eraseBar))
				}
				newTab := tabs[activeIndex]
				switchActiveTab(out, exitedTab, newTab, tabs, activeIndex, cols, rows, outerAltActive)
			}
			sessionMu.Unlock()
		}
	}

done:
	// Restore outer terminal cooked mode before printing status messages to errOut
	if inFD >= 0 && oldState != nil {
		_ = term.Restore(inFD, oldState)
	}
	if sessionBarVisible(loom.SpeccedDefaults.Session, len(tabs)) {
		fmt.Fprintf(out, "\x1b[r\x1b[%d;1H\x1b[2K\r\n", rows)
	}

	shotMu.Lock()
	for _, p := range savedShots {
		fmt.Fprintf(errOut, "Saved screenshot: %s\n", p)
	}
	for _, err := range shotErrors {
		fmt.Fprintf(errOut, "Error saving screenshot: %v\n", err)
	}
	shotMu.Unlock()

	if finalWaitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(finalWaitErr, &exitErr) {
			finalExitCode = exitErr.ExitCode()
		} else {
			return finalWaitErr
		}
	}

	if finalExitCode != 0 {
		return ExitError{code: finalExitCode}
	}
	return nil
}

func killProcessGroup(pid int) {
	_ = syscall.Kill(-pid, syscall.SIGHUP)
	_ = syscall.Kill(pid, syscall.SIGHUP)
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	_ = syscall.Kill(pid, syscall.SIGTERM)
}
