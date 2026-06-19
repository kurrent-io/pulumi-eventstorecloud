#!/usr/bin/env bash
#
# Live alias-migration test: proves the eventstorecloud -> kurrentcloud rename is
# non-destructive. It stands up a resource with the OLD provider, switches the
# program to the NEW (local) provider, and asserts the update performs NO
# replacements (only an in-place type migration via Pulumi aliases).
#
# This is the empirical complement to the offline guard
# (TestProviderAliasesInSchema in ../live). Run it manually / periodically.
#
# Prerequisites:
#   - pulumi, node, npm on PATH
#   - The NEW provider plugin on PATH: build it with `make provider` from the
#     repo root and `export PATH="$PWD/bin:$PATH"`, OR `pulumi plugin install
#     resource kurrentcloud --server github://api.github.com/kurrent-io`.
#   - The built local Node SDK at sdk/nodejs/bin (run `make build_nodejs`).
#   - Kurrent Cloud creds: ESC_TOKEN, ESC_ORG_ID (and ESC_URL if non-default).
#   - Network access so Pulumi can fetch the OLD eventstorecloud plugin for phase 1.
#
# Tunables (env):
#   OLD_PKG_VERSION   npm range for @eventstore/pulumi-eventstorecloud (default: latest)
#
set -euo pipefail

fail() { echo "ERROR: $*" >&2; exit 1; }

command -v pulumi >/dev/null || fail "pulumi not on PATH"
command -v npm >/dev/null || fail "npm not on PATH"
command -v pulumi-resource-kurrentcloud >/dev/null || \
  fail "pulumi-resource-kurrentcloud not on PATH (run 'make provider' and add ./bin to PATH)"
[ -n "${ESC_TOKEN:-}" ] && [ -n "${ESC_ORG_ID:-}" ] || fail "set ESC_TOKEN and ESC_ORG_ID"

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
LOCAL_SDK="$REPO_ROOT/sdk/nodejs/bin"
[ -d "$LOCAL_SDK" ] || fail "built local Node SDK not found at $LOCAL_SDK (run 'make build_nodejs')"

OLD_PKG_VERSION="${OLD_PKG_VERSION:-latest}"
WORKDIR="$(mktemp -d)"
STACK="alias-migration-$$"
export PULUMI_BACKEND_URL="file://$WORKDIR/state"
export PULUMI_CONFIG_PASSPHRASE="${PULUMI_CONFIG_PASSPHRASE:-}"

cleanup() {
  echo "--- cleanup ---"
  ( cd "$WORKDIR" 2>/dev/null && pulumi destroy -y -s "$STACK" >/dev/null 2>&1 || true
    cd "$WORKDIR" 2>/dev/null && pulumi stack rm -y -s "$STACK" >/dev/null 2>&1 || true )
  rm -rf "$WORKDIR"
}
trap cleanup EXIT

cd "$WORKDIR"

cat > Pulumi.yaml <<YAML
name: alias-migration
runtime: nodejs
description: eventstorecloud -> kurrentcloud alias migration test
YAML

# ---- Phase 1: create a Project with the OLD provider --------------------------
cat > package.json <<JSON
{
  "name": "alias-migration",
  "dependencies": {
    "@pulumi/pulumi": "^3.0.0",
    "@eventstore/pulumi-eventstorecloud": "$OLD_PKG_VERSION"
  }
}
JSON

cat > index.ts <<'TS'
import * as esc from "@eventstore/pulumi-eventstorecloud";
const project = new esc.Project("alias-project", { name: "alias-migration-test" });
export const projectId = project.id;
TS

echo "=== Phase 1: provisioning with OLD eventstorecloud provider ==="
npm install --silent
pulumi stack init "$STACK" >/dev/null
pulumi up -y -s "$STACK"
OLD_ID="$(pulumi stack output projectId -s "$STACK")"
echo "created project id: $OLD_ID"

# ---- Phase 2: switch to the NEW (local) provider and preview ------------------
cat > package.json <<JSON
{
  "name": "alias-migration",
  "dependencies": {
    "@pulumi/pulumi": "^3.0.0",
    "@kurrent-io/pulumi-kurrentcloud": "file:$LOCAL_SDK"
  }
}
JSON

cat > index.ts <<'TS'
import * as kurrent from "@kurrent-io/pulumi-kurrentcloud";
const project = new kurrent.Project("alias-project", { name: "alias-migration-test" });
export const projectId = project.id;
TS

echo "=== Phase 2: switching to NEW kurrentcloud provider; previewing ==="
rm -rf node_modules package-lock.json
npm install --silent

pulumi preview --json -s "$STACK" > preview.json
python3 - preview.json <<'PY'
import json, sys
data = json.load(open(sys.argv[1]))
bad = {"replace", "create-replacement", "delete-replaced", "delete"}
offending = [s for s in data.get("steps", []) if s.get("op") in bad]
print("change summary:", data.get("changeSummary", {}))
if offending:
    for s in offending:
        print("REPLACEMENT/DELETE step:", s.get("op"), s.get("urn"))
    print("\nFAIL: switching providers would replace/destroy resources — aliases are not working.")
    sys.exit(1)
print("\nPASS: no replacements/deletes — the alias migration is non-destructive.")
PY

echo "=== applying the migration to confirm the id is preserved ==="
pulumi up -y -s "$STACK"
NEW_ID="$(pulumi stack output projectId -s "$STACK")"
echo "project id after migration: $NEW_ID"
[ "$OLD_ID" = "$NEW_ID" ] || fail "project id changed across migration ($OLD_ID -> $NEW_ID)"
echo "PASS: project id preserved across the eventstorecloud -> kurrentcloud migration."
