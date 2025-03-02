#!/bin/sh

set -e

INPUT_DRY_RUN="${INPUT_DRY_RUN:-false}"

echo "Dry-run mode: ${INPUT_DRY_RUN}"

rm -rf ./issues

/app/issue-scouter

git config --global user.name "github-actions[bot]"
git config --global user.email "github-actions[bot]@users.noreply.github.com"

if git diff --quiet; then
  echo "No changes detected. Exiting."
  exit 0
fi

if [ "${INPUT_DRY_RUN}" = "true" ]; then
  echo "Dry-run mode: Changes detected but not committing."
  exit 0
fi

git add README.md ./issues
git commit -m "Update issue list"
git push origin main

echo "Changes have been pushed successfully!"
