#!/usr/bin/env python3
"""Regression tests for push commit selection in the governance checker."""

from __future__ import annotations

import importlib.util
import json
import os
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch


SCRIPT = Path(__file__).with_name("verify-commit-messages.py")
SPEC = importlib.util.spec_from_file_location("verify_commit_messages", SCRIPT)
if SPEC is None or SPEC.loader is None:
    raise RuntimeError("could not load verify-commit-messages.py")
CHECKER = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(CHECKER)


class PushEventCommitSelectionTests(unittest.TestCase):
    def test_force_push_uses_event_commits_when_before_sha_is_unavailable(self) -> None:
        first = "a" * 40
        merge = "b" * 40
        head = "c" * 40
        event_path = Path("artifacts/cache/force-push-event.json").resolve()
        event_path.parent.mkdir(parents=True, exist_ok=True)
        event_path.write_text(
            json.dumps({"commits": [{"id": first}, {"id": merge}, {"id": head}]}),
            encoding="utf-8",
        )

        def fake_git(*args: str) -> str:
            if args[:3] == ("rev-list", "--reverse", "--no-merges"):
                raise subprocess.CalledProcessError(128, ["git", *args])
            if args[:2] == ("rev-list", "--parents"):
                commit = args[-1]
                parents = "parent-one parent-two" if commit == merge else "parent-one"
                return f"{commit} {parents}"
            raise AssertionError(f"unexpected git invocation: {args}")

        try:
            with patch.dict(
                os.environ,
                {
                    "EVENT_NAME": "push",
                    "BEFORE_SHA": "d" * 40,
                    "HEAD_SHA": head,
                    "GITHUB_EVENT_PATH": str(event_path),
                },
            ), patch.object(CHECKER, "git", side_effect=fake_git):
                self.assertEqual(CHECKER.push_event_commits(), [first, head])
        finally:
            event_path.unlink(missing_ok=True)


if __name__ == "__main__":
    unittest.main()
