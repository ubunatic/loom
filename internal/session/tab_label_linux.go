package session

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

func getForegroundInfo(masterFd uintptr) (command, cwd string) {
	pgid, err := unix.IoctlGetInt(int(masterFd), unix.TIOCGPGRP)
	if err != nil {
		return "", ""
	}

	commBytes, err := os.ReadFile(filepath.Join("/proc", formatInt(pgid), "comm"))
	if err != nil {
		return "", ""
	}
	command = strings.TrimSpace(string(commBytes))

	cwd, _ = os.Readlink(filepath.Join("/proc", formatInt(pgid), "cwd"))
	return command, cwd
}

func formatInt(i int) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
