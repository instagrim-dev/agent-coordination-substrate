#!/usr/bin/env bash
# publish.sh — Create the public repo and initial release.
# Run from the agent-coordination-substrate/ directory.
# Prerequisites: gh CLI authenticated, git configured.

set -euo pipefail

REPO="instagrim-dev/agent-coordination-substrate"
TAG="v0.1.0"

echo "=== Step 1: Create GitHub repository ==="
gh repo create "$REPO" \
  --public \
  --description "Specification for environment-mediated coordination between autonomous agents" \
  --source . \
  --remote origin \
  --push

echo ""
echo "=== Step 2: Set repository topics ==="
gh repo edit "$REPO" \
  --add-topic multi-agent \
  --add-topic coordination \
  --add-topic substrate \
  --add-topic protocol \
  --add-topic stigmergy \
  --add-topic enforcement \
  --add-topic agent-coordination \
  --add-topic specification

echo ""
echo "=== Step 3: Create initial release ==="
gh release create "$TAG" \
  --repo "$REPO" \
  --title "v0.1.0 — Initial Publication" \
  --notes-file .github/RELEASE-v0.1.0.md

echo ""
echo "=== Done ==="
echo "Repository: https://github.com/$REPO"
echo "Release:    https://github.com/$REPO/releases/tag/$TAG"
echo ""
echo "Next steps:"
echo "  1. Verify repo is public: gh repo view $REPO"
echo "  2. Verify topics appear in search"
echo "  3. Add cross-link from BMO docs to this spec"
