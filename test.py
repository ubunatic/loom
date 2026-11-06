import json

with open("spec/schemas/defaults.schema.json", "r") as f:
    d = json.load(f)

session_props = d["properties"]["session"]["properties"]
session_props["mouse_mode"] = { "type": "string", "enum": ["off", "basic", "full"] }
d["properties"]["session"]["required"].insert(1, "mouse_mode")

bar_labels_props = session_props["bar_labels"]["properties"]
bar_labels_props["mouse_icon_basic"] = { "type": "string", "minLength": 1 }
bar_labels_props["mouse_icon_full"] = { "type": "string", "minLength": 1 }
bar_labels_props["mouse_icon_off"] = { "type": "string", "minLength": 1 }

session_props["bar_labels"]["required"].remove("mouse_icon")
session_props["bar_labels"]["required"].insert(4, "mouse_icon_basic")
session_props["bar_labels"]["required"].insert(5, "mouse_icon_full")
session_props["bar_labels"]["required"].insert(6, "mouse_icon_off")

del bar_labels_props["mouse_icon"]


with open("spec/schemas/defaults.schema.json", "w") as f:
    json.dump(d, f, separators=(',', ':')) # try a different formatter if they use that?
