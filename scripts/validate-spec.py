#!/usr/bin/env python3
"""Validate authored specs with standard JSON Schema; no Go runtime dependency.

Prerequisites: Python 3, PyYAML, jsonschema (system packages are sufficient).
Run from any directory. Unknown IDs and padding-fit rules are checked by Go.
"""

import copy
import json
from pathlib import Path

import jsonschema
import yaml


def main():
    root = Path(__file__).resolve().parent.parent
    pairs = (
        ("spec/box.yaml", "spec/schemas/box.schema.json"),
        ("examples/monitor/spec/watch.yaml", "spec/schemas/watch.schema.json"),
        ("examples/monitor/spec/monitor.yaml", "spec/schemas/monitor.schema.json"),
    )
    for document_path, schema_path in pairs:
        schema = json.loads((root / schema_path).read_text())
        jsonschema.Draft7Validator.check_schema(schema)
        validator = jsonschema.Draft7Validator(schema)
        document = yaml.safe_load((root / document_path).read_text())
        validator.validate(document)
        # Negative control: a typo must not silently pass schema validation.
        invalid = copy.deepcopy(document)
        invalid["misspelled_property"] = True
        if validator.is_valid(invalid):
            raise RuntimeError(f"{schema_path}: accepts unknown properties")
        if "view" in document:
            for change in ("negative_width", "unknown_box_field", "missing_id", "ambiguous_root"):
                invalid = copy.deepcopy(document)
                box = invalid["view"]["frame"]["boxes"][0]
                if change == "negative_width":
                    box["width"] = -1
                elif change == "unknown_box_field":
                    box["widht"] = 31
                elif change == "missing_id":
                    del box["id"]
                else:
                    invalid["pane"] = invalid["view"]
                if validator.is_valid(invalid):
                    raise RuntimeError(f"{schema_path}: accepts {change}")
        print(f"validated {document_path} (including negative controls)")


if __name__ == "__main__":
    main()
