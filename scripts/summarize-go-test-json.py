#!/usr/bin/env python3
"""Append failed Go test names and their output to a GitHub job summary."""

import json
import os
import sys
from collections import defaultdict
from pathlib import Path


def main() -> int:
    test_json = Path(sys.argv[1] if len(sys.argv) > 1 else "test.json")
    summary_path = os.environ.get("GITHUB_STEP_SUMMARY")
    if not summary_path:
        print("GITHUB_STEP_SUMMARY is not set", file=sys.stderr)
        return 2

    output: dict[tuple[str, str], list[str]] = defaultdict(list)
    failed: set[tuple[str, str]] = set()
    if not test_json.exists():
        report = "## Go race test failures\n\n`test.json` was not produced.\n"
    else:
        for line in test_json.read_text(encoding="utf-8").splitlines():
            try:
                event = json.loads(line)
            except json.JSONDecodeError:
                continue
            key = (event.get("Package", "unknown package"), event.get("Test", ""))
            if event.get("Action") == "fail":
                failed.add(key)
            if event.get("Action") == "output" and event.get("Output"):
                output[key].append(event["Output"])

        if not failed:
            report = "## Go race test failures\n\nNo failing test events were recorded.\n"
        else:
            sections = ["## Go race test failures\n"]
            for package, test in sorted(failed):
                name = f"{package} / {test}" if test else package
                log = "".join(output.get((package, test), []))
                sections.append(f"### `{name}`\n\n```text\n{log.rstrip()}\n```\n")
            report = "\n".join(sections)

    with Path(summary_path).open("a", encoding="utf-8") as summary:
        summary.write(report)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
