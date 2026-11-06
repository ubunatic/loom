//go:build !linux

package session

func getForegroundInfo(masterFd uintptr) (command, cwd string) {
	return "", ""
}
