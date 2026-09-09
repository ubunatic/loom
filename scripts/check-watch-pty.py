#!/usr/bin/env python3
"""Bounded Linux PTY probe/smoke test. Pass a command after the script name.

Without arguments, probe raw input and restoration with a tiny Python child.
With a monitor binary, observe idle redraws, answer DSR, resize, then send q.
"""
import fcntl
import os
import pty
import select
import signal
import struct
import subprocess
import sys
import termios
import time

master, slave = pty.openpty()
before = termios.tcgetattr(slave)
fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 24, 80, 0, 0))

def attach():
    os.setsid()
    fcntl.ioctl(0, termios.TIOCSCTTY, 0)

probe = not sys.argv[1:]
command = sys.argv[1:] or [sys.executable, "-c", "import os,tty,termios; old=termios.tcgetattr(0); tty.setraw(0); print('ready',flush=True); os.read(0,1); termios.tcsetattr(0,termios.TCSANOW,old)"]
child = subprocess.Popen(command, stdin=slave, stdout=slave, stderr=slave, preexec_fn=attach)
output = bytearray()
started = time.monotonic()
sent = resized = False
try:
    while child.poll() is None and time.monotonic() - started < 6:
        if select.select([master], [], [], 0.05)[0]:
            chunk = os.read(master, 65536)
            output.extend(chunk)
            if b"\x1b[6n" in chunk and not os.environ.get("LOOM_TEST_NO_DSR"):
                os.write(master, b"\x1b[1;1R")
        elapsed = time.monotonic() - started
        if not resized and elapsed > 1:
            fcntl.ioctl(slave, termios.TIOCSWINSZ, struct.pack("HHHH", 20, 60, 0, 0))
            os.kill(child.pid, signal.SIGWINCH)
            resized = True
        if not sent and elapsed > (0.2 if probe else 2.2):
            if os.environ.get("LOOM_TEST_SIGNAL"):
                os.kill(child.pid, signal.SIGTERM)
            else:
                os.write(master, b"q")
            sent = True
    if child.poll() is None:
        raise RuntimeError("child did not quit within six seconds")
    if child.returncode != 0:
        raise RuntimeError(f"child exited {child.returncode}: {output!r}")
    if termios.tcgetattr(slave) != before:
        raise RuntimeError("terminal mode was not restored")
    if not probe and output.count(b"\x1b[?25l") < 20:
        raise RuntimeError("too few idle redraws")
    print(f"PTY {'probe' if probe else 'watch'} passed; {len(output)} bytes; terminal restored")
finally:
    if child.poll() is None:
        child.kill()
        child.wait()
    os.close(master)
    os.close(slave)
