#!/bin/sh
# Same closed-by-default condition as fullRegenGate in pipeline/protect.go.
# task pipeline runs this before every generation command, including go run.
# Checked-in pipeline/kind_enumer.go is not regenerated here.
set -eu
if [ "${GCSIM_ALLOW_FULL_REGEN:-}" != "1" ]; then
	echo "refusing full pipeline regeneration: set GCSIM_ALLOW_FULL_REGEN=1 only when released character data and a stable generated-table strategy exist; checked-in generated files were not modified" >&2
	exit 1
fi
