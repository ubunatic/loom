// SPDX-FileCopyrightText: 2026 Uwe Jugel
// SPDX-License-Identifier: AGPL-3.0-or-later

//go:build windows

package ansiviewer

import (
	"context"
	"errors"
	"io"
	"time"
)

func RecordCommand(context.Context, io.Writer, time.Duration, string, ...string) error {
	return errors.New("ansiviewer: command recording with a pseudo-terminal is unsupported on Windows")
}
