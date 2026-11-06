import re

with open("spec/schemas/defaults.schema.json", "r") as f:
    original = f.read()

# I will just insert the new fields inline instead of parsing and dumping to avoid massive formatting changes
original = re.sub(
    r'"mouse": \{ "type": "string", "enum": \["safe"\] \},',
    '"mouse": { "type": "string", "enum": ["safe"] },\n        "mouse_mode": { "type": "string", "enum": ["off", "basic", "full"] },',
    original
)

original = re.sub(
    r'"required": \["mouse", "bar", "scrollback_lines",',
    '"required": ["mouse", "mouse_mode", "bar", "scrollback_lines",',
    original
)

original = re.sub(
    r'"required": \["tools", "help", "quit", "screenshot", "mouse_icon",',
    '"required": ["tools", "help", "quit", "screenshot", "mouse_icon_basic", "mouse_icon_full", "mouse_icon_off",',
    original
)

original = re.sub(
    r'"mouse_icon": \{ "type": "string", "minLength": 1 \},',
    '"mouse_icon_basic": { "type": "string", "minLength": 1 },\n            "mouse_icon_full": { "type": "string", "minLength": 1 },\n            "mouse_icon_off": { "type": "string", "minLength": 1 },',
    original
)

with open("spec/schemas/defaults.schema.json", "w") as f:
    f.write(original)
