# Batch 4 engine integration

Integration branch `cursor/batch4-engine-integration`, cut from `yuheng-production` at `7b58cbadbaa32086e5cc7a5ea1b01969f6e98d69`.

The three source drafts were merged with `--no-ff` and were not modified:

| PR | Branch | HEAD | Merge |
| --- | --- | --- | --- |
| #4 Nefer | `cursor/nefer-gcsim-5552` | `cc298e5e23a44a0502d6839ab3395fde924927fd` | `merge(nefer)` |
| #5 Vesna | `cursor/vesna-gcsim-5552` | `df1d88356a3821fb3b958d09456c54247b50ba65` | `merge(vesna)` |
| #6 Sandrone | `cursor/sandrone-gcsim-batch4` | `a7b76c665a789edfb6787401b268f99a9f967bd8` | `merge(sandrone)` |

## Key order

Nefer's branch inserts `keys.Nefer` between Navia and Neuvillette. Vesna's branch inserts `keys.Vesna` between Venti and Vodyanitsa. Both shifts would renumber every later production character. The merges move those consts, with Sandrone, to the tail:

`Zhongli`, `Nefer`, `Vesna`, `Sandrone`, `InvalidChar`.

Zhongli stays 130. Neuvillette stays 88. Vodyanitsa stays 114. Nilou stays 90. `InvalidChar` moves from 131 to 134 because it is the sentinel after the new keys.

Shortcut, import, asset, and catalog entries stay in alphabetical map order. That does not change iota values.

## ICD order

Vesna and Sandrone both appended after `ICDGroupLohenSkill`. The resolved tail is:

1. `ICDGroupVesnaSkillStart` — 150 frames, element `1,0,0` repeated, damage all 1. Vesna's entry counter. Same shape as `default`, separate counter.
2. `ICDGroupVesnaSkill` — 120 frames, element `1,0`, damage `1`. Vesna's Windborne/Pinion counter.
3. `ICDGroupSandroneMachineGun` — 150 frames, element `1,0,0` repeated, damage all 1. Provisional. Bullet `41133002` is still one attach hit.

`ICDGroupLohenSkill` stays 65. `ICDGroupAlyoshaElementalBurst` stays 64. Every production group before that tail keeps its index, timer, and sequence length.

## Pipeline

`pipeline/config.yml` now fully specifies Alyosha, Lohen, both Vesna groups, and Sandrone's machine-gun group, so those rows are not looked up in excel. `compareICDGroup` forces that tail after the alphabetical block.

`task pipeline` runs `scripts/pipeline-preflight.sh` before `go run ./pipeline`. The preflight returns unless `GCSIM_ALLOW_FULL_REGEN=1`, so a closed gate never reaches `go generate` or `go run`. `pipeline/kind_enumer.go` stays the checked-in enumer output. `task pipeline-enumer` is an unsupported maintenance command that rewrites that file, and only after the same preflight. It is not part of `task pipeline` or `task build`.

`pipeline run` also returns at `fullRegenGate` unless `GCSIM_ALLOW_FULL_REGEN=1`. That check happens before fetch and compile. When the gate is open, `assertHandMaintainedRegen` still runs before any generator writes. `GenerateCharacters` and `GenerateICDGroup` repeat the character and ICD checks. A character list or ICD order that drops a pinned name, inserts a name in the middle, or shortens a hand-maintained sequence returns an error and does not write the tables.

`scripts/pipeline-clean.sh` always refuses. It does not `git rm` generated files, including when `GCSIM_ALLOW_FULL_REGEN=1`. `.github/workflows/ci-pipeline.yml` does not call `pipeline-clean`. A manual dispatch with `allow_full_regen` left false exits 1 after checking that the worktree is clean. Opting in runs `task pipeline` without a preceding delete. The in-process character and ICD guards still fail closed.

`task build` compiles `./cmd/gcsim` from the checked-in generated data. `task build-generated` is the old protos, pipeline, and editor-keys sequence. It is not a supported build.

Nefer, Vesna, and Sandrone have no excel avatar in this module, so a full `pipeline run` cannot satisfy the character pin even after the env gate opens. That refusal is intentional. Unattended full regeneration remains a release blocker: there is no path that refreshes official generated characters while leaving these three kits untouched, because generation rewrites the shared tables in one pass. Deleting stale generated files stays disabled until that strategy exists.

`go run ./pipeline run` was not executed with `GCSIM_ALLOW_FULL_REGEN=1` for this correction. The closed gate was executed, and it did not modify tracked files.

## Nefer premium team, stage 1

`docs/yuheng/nefer-premium-stage1.md` records a resource-natural execution check of Nefer / Lauma / Columbina / Sucrose. The four configs under `docs/yuheng/configs/nefer-premium-*.txt` are experimental. They are not a damage benchmark and they are not a backend rotation profile. No character mechanic was changed for that check.

## Zibai

`docs/yuheng/zibai.md` records the append-only port of upstream Zibai. The engine key sits after Sandrone. No ICD group was added. The config under `docs/yuheng/configs/zibai-c0.txt` is a registration check, not a rotation.
