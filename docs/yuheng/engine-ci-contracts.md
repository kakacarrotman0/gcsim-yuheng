# Engine compatibility contracts in CI

Workflow: `.github/workflows/tests.yml`, job `test-engine-contracts`.

The job runs on every `push` and `pull_request` that starts `run tests`. It does not use a path filter. A filter would let a Go change skip the contracts and still report success.

`GCSIM_ALLOW_FULL_REGEN` is unset for the test step. The job does not run `task pipeline`, `pipeline run`, or any other regeneration. The pipeline package's own tests invoke those commands and require them to refuse.

## What the job executes

| Package | Existing tests | What a failure means |
| --- | --- | --- |
| `pkg/core/keys` | `TestCharacterIDsMatchProductionPin` | A character const moved, or a name was inserted ahead of the pinned tail. Zhongli is 130, Neuvillette 88, Vodyanitsa 114, Nilou 90. The tail is Nefer, Vesna, Sandrone, Zibai, then `InvalidChar`. |
| `pkg/core/attacks` | `TestICDGroupIDsAndSequenceLengths`, `TestHandMaintainedICDSequences` | An ICD group index, timer, or hand-maintained sequence changed. New groups append after the pinned tail. |
| `pipeline` | Regeneration gate, generated-file safety, and ICD order checks in `regen_gate_test.go` and `icd_test.go` | `pipeline run`, `task pipeline`, or `scripts/pipeline-clean.sh` proceeded without the closed gate, or a generated file changed while the gate was closed. |

Go is 1.27.0, the same version as `.github/actions/go-setup-and-test`. `task` is go-task **v3.44.0**, installed from the upstream Linux amd64 tarball and checked against sha256 `d6c9c0a14793659766ee0c06f9843452942ae6982a3151c6bbd78959c1682b82`. The pipeline tests call `task` on `PATH`. A missing `task` fails the job. It is not skipped.

## What this job does not cover

- The existing `test-go` job is unchanged. It still runs `go test ./backend/... ./internal/... ./pkg/gcs/... ./pkg/reactable/...` only when its path filter matches. The commented broader `./pkg/...` suite stays out of CI because that tree has known baseline failures.
- Character mechanics, hitmarks, reactions, and talent formulas. A green contract job does not show that a character behaves correctly.
- Full pipeline regeneration. `ci-pipeline.yml` remains manual `workflow_dispatch`. Its refusal check still runs before `task pipeline --`.
- UI tests, frame evidence, and simulator configs under `docs/yuheng/configs/`.
- Branch protection. This workflow starts the job. Making `test-engine-contracts` a merge blocker on a protected branch is a separate repository setting and is not changed here.

## Unvalidated frame debt

`internal/tests/frames_func` runs inside the existing `test-go` command (`./internal/...`). `TestFramesFunc` still checks every character and every transition. Four existing rows are excused only when the character, executed action, transition, animation length, queue-open frame, and cancel frame all match `knownTimingDebt`.

These rows are unvalidated timing debt. They are not measured gameplay and they are not a claim that the transition is correct.

| Character | Action | Transition | Animation | Queue opens | Recorded frame | Why it is debt |
| --- | --- | --- | --- | --- | --- | --- |
| Kachina | N4 | Charge | 51 | 16 | 500 | Sheet sentinel. The charge cancel is unavailable until N4's animation expires. |
| Kachina | Charge | Charge | 55 | 20 | 500 | Sheet sentinel. A second charge is unavailable until the charge animation expires. |
| Linnea | Skill | Attack, Aim, Skill, Burst, Walk | 14 | 13 | 33, 44, 20, 15, 32 | Draft slice marked unknown. Those cancels are longer than the placeholder animation, so they are not taken. |
| Linnea | N3 | Aim, Skill, Burst, Dash, Jump, Swap | 43 | 47 | 47 | `attack.go` still marks the frames unimplemented. The hitmark is past the N3 → N1 frame. |
| Linnea | N3 | Every other transition | 43 | 47 | 43 | Same unimplemented pair. The queue opens at the hitmark, after these frames. |

A different number, character, action, or transition fails the test. Kachina's other charge cancels, including N1 → Charge at frame 20, stay under the strict check. Linnea's burst cancel is not in this table. `TestChargeTransitionsStayUnavailableUntilTheActionExpires` and `TestBurstSkillCancelIsReadyWhenTheQueueOpens` still run.

## How a later engine batch should report tests

Paste the commands and the commit they ran against. Do not substitute a summary for the result line.

```text
HEAD: <full sha>
branch: <branch>

go test ./backend/... ./internal/... ./pkg/gcs/... ./pkg/reactable/...
go test ./pkg/core/keys/ ./pkg/core/attacks/ ./pipeline/

GCSIM_ALLOW_FULL_REGEN: unset
pipeline regeneration: not run
```

Add the character packages the batch actually touched, for example `go test ./internal/characters/dori/`. Say which GitHub Actions run finished and whether `test-engine-contracts` passed. If a contract fails, name the test and the pin it disagreed with. Do not weaken the test or renumber keys or ICD groups to make it pass.

## Licensing review flag

`LICENSE` is the GNU Affero General Public License v3.0. `CONTRIBUTING.md` states that AGPLv3 is the project license effective 2026-09-19, that released versions through v2.47.2 were under the MIT license in `LICENSE-MIT`, and that contributions are under AGPLv3.

A future public web service needs an owner-approved compliance review before it is offered. This note records that flag. It does not modify either license, does not deploy a service, and does not decide whether a deployment would comply.
