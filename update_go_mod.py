#!/usr/bin/env python3

from __future__ import annotations

import json
import os
import re
import sys
import urllib.request
from pathlib import Path


GO_LINE_RE = re.compile(r"^(?P<indent>\s*)go\s+(?P<version>\S+)\s*$")
TOOLCHAIN_LINE_RE = re.compile(r"^(?P<indent>\s*)toolchain\s+go(?P<version>\S+)\s*$")
GO_RELEASES_URL = "https://go.dev/dl/?mode=json"


def parse_bool(value: str) -> bool:
    return value.strip().lower() in {"1", "true", "yes", "on"}


def parse_version(version: str) -> tuple[int, ...]:
    cleaned = version.strip()
    if cleaned.startswith("go"):
        cleaned = cleaned[2:]
    parts = cleaned.split(".")
    return tuple(int(part) for part in parts)


def load_release_data() -> list[dict]:
    inline = os.environ.get("GO_RELEASES_JSON", "").strip()
    if inline:
        return json.loads(inline)

    with urllib.request.urlopen(GO_RELEASES_URL, timeout=30) as response:
        return json.loads(response.read().decode("utf-8"))


def latest_stable_version() -> str:
    releases = load_release_data()
    stable_versions = []
    for release in releases:
        version = release.get("version", "")
        stable = release.get("stable", False)
        if not stable or not version.startswith("go"):
            continue
        stable_versions.append(version[2:])

    if not stable_versions:
        raise RuntimeError("No stable Go releases found from go.dev")

    return max(stable_versions, key=parse_version)


def set_output(name: str, value: str) -> None:
    output_file = os.environ.get("GITHUB_OUTPUT")
    if not output_file:
        return

    with open(output_file, "a", encoding="utf-8") as handle:
        handle.write(f"{name}={value}\n")


def main() -> int:
    go_mod_path = Path(os.environ.get("INPUT_GO_MOD_PATH", "go.mod"))
    update_toolchain = parse_bool(os.environ.get("INPUT_UPDATE_TOOLCHAIN", "true"))

    if not go_mod_path.exists():
        raise FileNotFoundError(f"go.mod path does not exist: {go_mod_path}")

    original = go_mod_path.read_text(encoding="utf-8")
    lines = original.splitlines()

    go_line_index = None
    previous_version = None
    toolchain_line_index = None

    for index, line in enumerate(lines):
        if go_line_index is None:
            match = GO_LINE_RE.match(line)
            if match:
                go_line_index = index
                previous_version = match.group("version")
                continue

        if toolchain_line_index is None and TOOLCHAIN_LINE_RE.match(line):
            toolchain_line_index = index

    if go_line_index is None or previous_version is None:
        raise RuntimeError(f'Could not find a "go" directive in {go_mod_path}')

    latest_version = latest_stable_version()
    changed = parse_version(previous_version) < parse_version(latest_version)

    if changed:
        go_match = GO_LINE_RE.match(lines[go_line_index])
        assert go_match is not None
        lines[go_line_index] = f'{go_match.group("indent")}go {latest_version}'

        if update_toolchain and toolchain_line_index is not None:
            toolchain_match = TOOLCHAIN_LINE_RE.match(lines[toolchain_line_index])
            assert toolchain_match is not None
            lines[toolchain_line_index] = (
                f'{toolchain_match.group("indent")}toolchain go{latest_version}'
            )

        updated = "\n".join(lines) + ("\n" if original.endswith("\n") else "")
        go_mod_path.write_text(updated, encoding="utf-8")
        print(f"Updated {go_mod_path} from Go {previous_version} to Go {latest_version}.")
    else:
        print(f"{go_mod_path} is already on the latest stable Go version: {previous_version}.")

    current_version = latest_version if changed else previous_version
    set_output("changed", "true" if changed else "false")
    set_output("previous-version", previous_version)
    set_output("current-version", current_version)
    set_output("latest-version", latest_version)
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f"error: {exc}", file=sys.stderr)
        raise SystemExit(1)
