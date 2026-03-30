#!/usr/bin/env python3
"""
Merge a new Helm chart index.yaml from a release into the helmcharts repo index.yaml.

Usage:
    merge-helm-index.py <release-index.yaml> <helmcharts-index.yaml> <output-index.yaml>

New chart versions from the release index are appended to the helmcharts index.
Existing versions are left unchanged (operation is idempotent).
If the helmcharts index does not exist yet, it is created from scratch.
"""
import os
import sys
from io import StringIO
from datetime import datetime, timezone
from ruamel.yaml import YAML


EMPTY_INDEX = """\
apiVersion: v1
entries: {}
generated: ""
"""


def merge_indexes(release_index_path, helmcharts_index_path, output_path):
    yaml = YAML()
    yaml.preserve_quotes = True
    yaml.width = 4096  # prevent line wrapping on long lines

    with open(release_index_path) as f:
        release_idx = yaml.load(f)

    if os.path.exists(helmcharts_index_path):
        with open(helmcharts_index_path) as f:
            helmcharts_idx = yaml.load(f)
    else:
        print(f"{helmcharts_index_path} not found, starting from an empty index.")
        helmcharts_idx = yaml.load(StringIO(EMPTY_INDEX))

    for chart_name, versions in release_idx.get("entries", {}).items():
        if chart_name not in helmcharts_idx["entries"]:
            helmcharts_idx["entries"][chart_name] = []
        existing_versions = {v["version"] for v in helmcharts_idx["entries"][chart_name]}
        for v in versions:
            if v["version"] not in existing_versions:
                helmcharts_idx["entries"][chart_name].append(v)

    now = datetime.now(timezone.utc)
    helmcharts_idx["generated"] = now.strftime("%Y-%m-%dT%H:%M:%S.") + f"{now.microsecond // 1000:03d}Z"

    with open(output_path, "w") as f:
        yaml.dump(helmcharts_idx, f)

    print(f"Merged index written to {output_path}")


if __name__ == "__main__":
    if len(sys.argv) != 4:
        print(__doc__)
        sys.exit(1)
    merge_indexes(sys.argv[1], sys.argv[2], sys.argv[3])
