#!/bin/sh
# Deleting generated files before validation is disabled.
# pipeline run used to be preceded by `git rm` of *.dm.* and reference docs.
# That removed hand-maintained characters before assertHandMaintainedRegen
# could refuse. Do not restore a delete step until released character data
# and a stable generated-table strategy exist.
set -eu
echo "refusing pipeline-clean: deleting generated files before a successful validation is disabled" >&2
echo "checked-in generated files were not modified" >&2
exit 1
