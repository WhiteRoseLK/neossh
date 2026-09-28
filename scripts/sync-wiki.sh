#!/usr/bin/env bash
set -euo pipefail

TOKEN=$(gh auth token)
REPO="WhiteRoseLK/neossh"
WIKI_URL="https://x-access-token:${TOKEN}@github.com/${REPO}.wiki.git"
TEMP_DIR=$(mktemp -d)

trap 'rm -rf "$TEMP_DIR"' EXIT

echo "Cloning GitHub Wiki repository for ${REPO}..."
if ! git clone "$WIKI_URL" "$TEMP_DIR/wiki_repo" 2>/dev/null; then
    echo "Error: Wiki repository is not yet initialized on GitHub."
    echo "Please visit https://github.com/${REPO}/wiki and click 'Create the first page' (or save a page) once."
    exit 1
fi

echo "Copying wiki pages..."
find "$TEMP_DIR/wiki_repo" -maxdepth 1 ! -name '.git' ! -name 'wiki_repo' -exec rm -rf {} +
cp -r wiki/* "$TEMP_DIR/wiki_repo/"

cd "$TEMP_DIR/wiki_repo"
git config user.name "WhiteRose"
git config user.email "50756181+WhiteRoseLK@users.noreply.github.com"
git add -A

if git diff --staged --quiet; then
    echo "Wiki is already up to date."
else
    git commit -m "docs: sync wiki pages from repository"
    git push origin master 2>/dev/null || git push origin main
    echo "✅ Wiki successfully synchronized to https://github.com/${REPO}/wiki !"
fi
