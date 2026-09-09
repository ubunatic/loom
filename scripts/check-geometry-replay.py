#!/usr/bin/env python3
"""Probe printf replay semantics, then compare saved ANSI replay bytes exactly."""
import json
from pathlib import Path
import subprocess

# Canary: shell printf must decode the same octal ESC used by the saved replay.
probe = subprocess.check_output(["bash", "-c", r"printf '%b\n' '\033[1mX\033[0m'"])
assert probe == b"\x1b[1mX\x1b[0m\n", repr(probe)
print("printf replay canary passed")
root = Path(__file__).resolve().parent.parent
for name in ("wide", "slim"):
    path = root / "testdata" / "geometry" / name
    golden = json.loads(path.with_suffix(".json").read_text())
    actual = subprocess.check_output(["bash", str(path.with_suffix(".sh"))])
    expected = ("\n".join(golden["Rows"]) + "\n").encode()
    assert actual == expected, f"{name}: replay differs from ANSI golden"
    print(f"{name}: {golden['Width']}x{golden['Height']}, replay matches golden")
