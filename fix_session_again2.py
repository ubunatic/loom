import re

with open("internal/session/session.go", "r") as f:
    content = f.read()

# Fix duplicated mouse drag block
content = content.replace("""					// SGR Mouse event (\\x1b[<...M or \\x1b[<...m)
					if ev, used, ok := loom.ScanMouse(pending); ok {
						if sessionTabMouseMode(tabs[activeIndex]) == MouseModeBasic && ev.Action == loom.MouseDrag {
							pending = pending[used:]
							prefixMode = false
							sessionMu.Unlock()
							continue
						}
						if sessionTabMouseMode(tabs[activeIndex]) == MouseModeBasic && ev.Action == loom.MouseDrag {
							pending = pending[used:]
							prefixMode = false
							sessionMu.Unlock()
							continue
						}""",
"""					// SGR Mouse event (\\x1b[<...M or \\x1b[<...m)
					if ev, used, ok := loom.ScanMouse(pending); ok {
						if sessionTabMouseMode(tabs[activeIndex]) == MouseModeBasic && ev.Action == loom.MouseDrag {
							pending = pending[used:]
							prefixMode = false
							sessionMu.Unlock()
							continue
						}""")
with open("internal/session/session.go", "w") as f:
    f.write(content)
