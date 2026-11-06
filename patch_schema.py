import re

with open("spec/schemas/defaults.schema.json", "r") as f:
    content = f.read()

content = content.replace('"mouse": { "type": "string", "enum": ["safe"] },', '"mouse": { "type": "string", "enum": ["safe"] },\n        "mouse_mode": { "type": "string", "enum": ["off", "basic", "full"] },')
content = content.replace('"required": ["mouse", "bar", "scrollback_lines",', '"required": ["mouse", "mouse_mode", "bar", "scrollback_lines",')
content = content.replace('"required": ["tools", "help", "quit", "screenshot", "mouse_icon",', '"required": ["tools", "help", "quit", "screenshot", "mouse_icon_basic", "mouse_icon_full", "mouse_icon_off",')
content = content.replace('"mouse_icon": { "type": "string", "minLength": 1 },', '"mouse_icon_basic": { "type": "string", "minLength": 1 },\n            "mouse_icon_full": { "type": "string", "minLength": 1 },\n            "mouse_icon_off": { "type": "string", "minLength": 1 },')

with open("spec/schemas/defaults.schema.json", "w") as f:
    f.write(content)
