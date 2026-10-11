#!/usr/bin/env python3
"""Check the required governance fields in newly submitted commit messages."""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys
from pathlib import Path


PERCENT_FIELDS = (
    (
        "Quanto falta da tarefa",
        re.compile(
            r"^\s*Quanto falta da tarefa[^:]*:\s*"
            r"([0-9]+(?:[.,][0-9]+)?)\s*%\.?\s*$",
            re.MULTILINE,
        ),
    ),
    (
        "Quanto falta do projeto",
        re.compile(
            r"^\s*Quanto falta (?:da meta )?do projeto:\s*"
            r"([0-9]+(?:[.,][0-9]+)?)\s*%\.?\s*$",
            re.MULTILINE,
        ),
    ),
)
TEXT_FIELDS = ("O que é", "O que foi feito", "Próximo passo")


def git(*args: str) -> str:
    result = subprocess.run(
        ["git", *args], check=True, capture_output=True, text=True, encoding="utf-8"
    )
    return result.stdout.strip()


def without_merge_commits(commits: list[str]) -> list[str]:
    result = []
    for commit in commits:
        parents = git("rev-list", "--parents", "-n", "1", commit).split()
        if len(parents) <= 2:
            result.append(commit)
    return result


def push_event_commits() -> list[str]:
    before = os.getenv("BEFORE_SHA", "")
    head = os.getenv("HEAD_SHA", "")
    if not head or set(head) == {"0"}:
        return []
    if before and set(before) != {"0"}:
        try:
            return git("rev-list", "--reverse", "--no-merges", f"{before}..{head}").splitlines()
        except subprocess.CalledProcessError:
            # A force-push can make the previous tip unavailable in a fresh
            # checkout. Fall back to the commit list supplied with the event.
            pass

    # A new branch has no usable `before` revision. The event payload lists the
    # commits that arrived in this push without forcing validation of old history.
    event_path = os.getenv("GITHUB_EVENT_PATH", "")
    if event_path and Path(event_path).is_file():
        payload = json.loads(Path(event_path).read_text(encoding="utf-8"))
        commits = [item.get("id", "") for item in payload.get("commits", [])]
        commits = [commit for commit in commits if commit and set(commit) != {"0"}]
        commits = without_merge_commits(commits)
        if commits:
            return commits
    return [head]


def commits_for_event() -> list[str]:
    event = os.getenv("EVENT_NAME", "")
    if event == "push":
        if os.getenv("PUSH_DELETED", "").lower() == "true":
            return []
        return push_event_commits()
    if event == "pull_request":
        head = os.getenv("PR_HEAD_SHA", "")
        base = os.getenv("PR_BASE_SHA", "")
        if head and base:
            commits = git("rev-list", "--reverse", "--no-merges", f"{base}..{head}").splitlines()
            if commits:
                return [commits[-1]]
        return [head] if head else []
    head = os.getenv("HEAD_SHA", "")
    return [head] if head else [git("rev-parse", "HEAD")]


def check_message(commit: str) -> list[str]:
    # `%b` returns only the body; `%B` would also allow the subject to satisfy
    # fields that CONTRIBUTING.md explicitly requires in the commit body.
    body = git("show", "-s", "--format=%b", commit)
    missing: list[str] = []
    for field in TEXT_FIELDS:
        if not re.search(rf"^\s*{re.escape(field)}:\s*\S", body, re.MULTILINE):
            missing.append(field)
    for label, pattern in PERCENT_FIELDS:
        match = pattern.search(body)
        if not match:
            missing.append(f"{label} (percentual de 0% a 100%)")
            continue
        percentage = float(match.group(1).replace(",", "."))
        if not 0 <= percentage <= 100:
            missing.append(f"{label} (percentual fora de 0% a 100%)")
    return missing


def main() -> int:
    commits = commits_for_event()
    if not commits:
        print("Nenhum commit novo para conferir.")
        return 0

    failed = False
    for commit in commits:
        short_sha = commit[:7]
        subject = re.sub(r"[\x00-\x1f\x7f-\x9f]", " ", git("show", "-s", "--format=%s", commit))
        subject = subject[:120]
        missing = check_message(commit)
        if missing:
            failed = True
            print(f"ERRO {short_sha} ({subject}): preencher {', '.join(missing)}.")
        else:
            print(f"OK {short_sha} ({subject}): campos obrigatórios preenchidos.")

    if failed:
        print(
            "Atualize a mensagem do novo commit com os campos definidos em "
            "CONTRIBUTING.md. Não inclua credenciais ou dados pessoais."
        )
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
