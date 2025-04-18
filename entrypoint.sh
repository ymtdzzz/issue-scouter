#!/bin/sh

set -e

INPUT_DRY_RUN="${INPUT_DRY_RUN:-false}"

echo "Dry-run mode: ${INPUT_DRY_RUN}"

/app/issue-scouter

git config --global --add safe.directory /github/workspace
git config --global user.name "github-actions[bot]"
git config --global user.email "github-actions[bot]@users.noreply.github.com"

git add --all
git commit -m "Update issue list"

if [ "${INPUT_DRY_RUN}" = "true" ]; then
  echo "Dry-run mode: Commit completed but not pushing."
  git status
  exit 0
fi

git push origin main

echo "Changes have been pushed successfully!"
