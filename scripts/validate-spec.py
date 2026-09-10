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
        ("testdata/fixtures/empty-shell.yaml", "spec/schemas/monitor.schema.json"),
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
            dynamic = copy.deepcopy(document)
            dynamic_box = dynamic["view"]["frame"]["boxes"][0]
            dynamic_box["dynamic"] = True
            dynamic_box.pop("width", None)
            dynamic_box.pop("height", None)
            validator.validate(dynamic)
            box = document["view"]["frame"]["boxes"][0]
            if "rows" in box:
                for change in (
                    "invalid_align",
                    "zero_column_width",
                    "negative_column_width",
                    "negative_gap",
                    "unknown_rows_field",
                    "unknown_column_field",
                    "non_string_value",
                ):
                    invalid = copy.deepcopy(document)
                    rbox = invalid["view"]["frame"]["boxes"][0]
                    rows = rbox["rows"]
                    if change == "invalid_align":
                        rows["columns"][0]["align"] = "center"
                    elif change == "zero_column_width":
                        rows["columns"][0]["width"] = 0
                    elif change == "negative_column_width":
                        rows["columns"][0]["width"] = -1
                    elif change == "negative_gap":
                        rows["gap"] = -1
                    elif change == "unknown_rows_field":
                        rows["unknown_field"] = True
                    elif change == "unknown_column_field":
                        rows["columns"][0]["unknown_field"] = True
                    elif change == "non_string_value":
                        rows["values"][0][0] = 123
                    if validator.is_valid(invalid):
                        raise RuntimeError(f"{schema_path}: accepts {change}")
        print(f"validated {document_path} (including negative controls)")


if __name__ == "__main__":
    main()
