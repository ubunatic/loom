// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

package loom

// Paneable is a Widget that can be driven by RunPane: it reports the height it
// wants (ContentHeight), the user's choice (Selected), and any navigation signal
// raised from its command prompt (Nav). Choice and Table implement it.
type Paneable interface {
	Widget
	ContentHeighter
	// Selected returns the chosen Item and whether a selection was made.
	Selected() (Item, bool)
	// Nav returns the navigation signal set by a prompt command (e.g. :home).
	// NavNone means the user made a normal selection or cancelled.
	Nav() Nav
}

// RunPane opens an inline Pane sized to w, runs w's event loop, and always
// closes the pane before returning — so the terminal is restored (and the
// reserved region cleared) before the caller writes any result to stdout.
//
// It returns the selection, the ok flag from w.Selected (false when the user
// aborted), and the widget's navigation signal. A non-nil error means the pane
// could not be opened or the event loop failed; on error item/ok/nav are zero.
//
// RunPane is the exported form of the open/run/close dance shown in the package
// doc. Terminal restore on SIGINT/SIGTERM and panic is handled by Pane itself.
func RunPane(w Paneable) (item Item, ok bool, nav Nav, err error) {
	pane, err := New(w.ContentHeight())
	if err != nil {
		return Item{}, false, NavNone, err
	}
	err = pane.Run(w)
	pane.Close()
	if err != nil {
		return Item{}, false, NavNone, err
	}
	item, ok = w.Selected()
	return item, ok, w.Nav(), nil
}
