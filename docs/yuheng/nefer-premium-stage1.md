# Nefer premium team, stage 1

Investigation only. The runs below used the engine at `825860a745113490b2cf3a8fcaff6bcddd650225`. No character mechanic was changed. These configs are not a damage benchmark and not a backend rotation profile.

Team: Nefer / Lauma / Columbina / Sucrose. All C0, level 90/90. One level-100 target with resist 0.1. `iteration=3`. `swap_delay=12` is a Yuheng validation assumption. Energy starts full because that is the engine default, not a KQM energy calculation. Talent levels and `add stats` lines are assumptions. No `set_starting_verdant_dew` and no other seeded combat resource.

`nefer-natural.txt` and `nefer-lunar.txt` were not edited.

## Baseline

The action lists follow the KQM Lauma Quick Guide sample rotations, not the Sucrose Quick Guide.

Lauma guide, Deepwood Sucrose:

`(Nefer E)¹ > Columbina E Q > Lauma (N1)¹ Q hE > Sucrose ED > Nefer combo > Sucrose ED`

Lauma guide, Instructor Sucrose:

`(Nefer E)¹ > Columbina E N1 > Sucrose ED > Lauma (N1)¹ Q hE > Columbina Q N1 > Sucrose ED² > Nefer E 3[CD] E 3[CD]`

Footnotes kept from that guide: the opening Nefer Skill and Lauma N1 are first-rotation only. The Instructor guide allows the second Sucrose `ED` to be replaced by `N1 Q` when Burst is available. These files use `ED`.

Sucrose's own guide differs:

- Deepwood: both Sucrose casts are `E`, not `ED`.
- Instructor: a longer two-cycle string, `Lauma Q hE > Columbina Q E > Sucrose N1 E > Nefer combo > Sucrose E/Q > Lauma Q hE > Columbina C Q E > Sucrose N1 E > Nefer combo`, with a note to wait before Columbina's Burst if Hydro Swirl is inconsistent.

The Lauma guide is the baseline because it is the rotation quoted for this team. gcsim has no `3[CD]` bracket syntax, so Nefer's preferred dash-cancel combo is written `skill, charge, dash` three times, then another `skill` and three `charge, dash`. Naming another character swaps. `skill[hold=1]` is Lauma's Hold. `attack` is N1.

Nefer, Lauma, and Columbina cancel frames used by these strings are placeholders. `docs/yuheng/nefer.md` records that no frame sheet was found. A `charge, dash` does not skip Phantasm's second Nefer hit: that hit is frame 45 and the dash cancel is frame 48.

## Builds

Shared:

| Character | Weapon | Set | Talents | `add stats` |
| --- | --- | --- | --- | --- |
| Nefer | Reliquary of Truth R1 | 4pc Night of the Sky's Unveiling | 9/9/9 | `hp=4780 atk=311 em=374 cr=0.20 cd=0.622 er=0.10` |
| Columbina | Prototype Amber R5 | 4pc Silken Moon's Serenade | 9/9/9 | `hp=4780 hp%=0.466 atk=311 cr=0.311 cd=0.20 er=0.518 em=80` |
| Sucrose | Sacrificial Fragments R5 | see variant | 6/6/6 | `hp=4780 atk=311 em=561 er=0.20 cr=0.10 cd=0.15` |

Deepwood variant: Sucrose 4pc Deepwood Memories. Lauma R1 Etherlight Spindlelute, 2pc Gilded Dreams and 2pc Emblem of Severed Fate, `hp=4780 atk=311 em=561 er=0.30 cr=0.15 cd=0.20`.

Instructor variant: Sucrose 4pc Instructor. Lauma R1 Etherlight Spindlelute, 4pc Deepwood Memories, same Lauma stats.

Sources and assumptions:

- KQM Lauma Quick Guide assigns the sets and the Sac Fragments / Prototype Amber weapons. Prototype Amber is the guide's sustain option, not Columbina's signature. Refinement levels are not given there. R5 on Prototype Amber and Sacrificial Fragments is an assumption. Etherlight Spindlelute R1 is an assumption for Lauma's practical ER catalyst. Reliquary of Truth R1 and 4pc Night of the Sky's Unveiling are Nefer's signature weapon and 4pc set used by the KQM Nefer guide.
- "2pc mixed EM/ER" is implemented as Gilded Dreams 2pc plus Emblem of Severed Fate 2pc. That pairing is an assumption of which two 2pc bonuses.
- Talent levels, the flat `add stats` lines, one target, and `swap_delay=12` are assumptions. They are not a KQM artifact sheet.

## What was run

`go build -o /tmp/gcsim-premium ./cmd/gcsim`, then each config twice: `-out` for the 3-iteration aggregate, and a separate `-sample` for one debug log. The sample is a different seed from the aggregate. Hitlag was on. Worker count in the result was 20. DPS from those files is not reported here.

| Config | Duration | Aggregate seed | Sample seed |
| --- | --- | --- | --- |
| `nefer-premium-deepwood-first.txt` | 45s | 8013613138064098725 | 7889316721155719816 |
| `nefer-premium-deepwood-sustain.txt` | 90s | 1397861054186145605 | 1598448435266442123 |
| `nefer-premium-instructor-first.txt` | 50s | 6538712694444929678 | 12646115053570412597 |
| `nefer-premium-instructor-sustain.txt` | 100s | 7767898233175803983 | 1038268647394553286 |

Character index in these files is 0 Nefer, 1 Columbina, 2 Lauma, 3 Sucrose. Where a hit or action mean has min equal to max, every iteration matched.

`failed_actions` values are seconds the worker waited, not skipped casts. The scripted action counts still completed.

## Deepwood results

First rotation, 45s. Every iteration: Nefer 3 skills, 6 charges, 6 dashes; Columbina 1 skill and 1 burst; Lauma 1 N1, 1 burst, 1 Hold; Sucrose 2 skills and 2 dashes. Warnings: `skill_cd` only.

Nefer hits: ordinary Charged Attack mean 1, and each Phantasm component mean 5 (Nefer 1, Nefer 2, Shade 1, Shade 2, Shade 3). Five of the six charged attacks were Phantasms. Columbina Gravity Interference (Lunar-Bloom) mean 15, Skill DoT mean 12. Lauma Hold and Hold Lunar-Bloom mean 1 each, Frostgrove Sanctuary mean 8. Sucrose skill hits mean 2. Reactions: Columbina `lunarbloom` mean 11, Lauma `lunarbloom` mean 2. Sucrose has no reaction row.

Sucrose `skill_cd` mean 2.072s (min 0, max 6.217). `swap_cd` mean 0.417s. Other failed-action means are 0. Field time is not active uptime: the config ends on Sucrose, and she shows mean 30.817s of the 45s because the sim idles on the last character.

Sample seed 7889316721155719816:

- Moonridge Dew gained at frames 171 and 209, before the first Verdant Dew at frame 201. Columbina's Burst is already up in this order, so her A4 can record Moonridge as soon as Lunar-Bloom starts.
- Lauma Hold at frame 348 consumes 2 Verdant Dew and then 1 Moonridge Dew. Hold damage is frame 364. The log is `Hymn of Eternal Rest (Hold)` and `Hymn of Eternal Rest (Hold) (Lunar-Bloom)`. Pale Hymn remaining after the first consume is 35, which matches Burst's 18 stacks plus 3 Moon Song times 6, minus the one stack that hit consumed. There is no separate "moon song" log line. That stack reading is inferred from `lauma pale hymn consumed` plus `addMoonSong`.
- Deepwood logs `dm 4pc proc` on Sucrose (char 3) at frames 429 and 1338. `sucrose a4 triggered` at those frames with EM snapshot 196.3024. No `sucrose a1` and no `swirl-hydro`.
- Nefer's first four Phantasm damages are frames 509, 579, 649, and 739. The charge that deals damage at frame 819 is an ordinary Charged Attack. It started before the Verdant Dew gain at frame 792, while Dew was 0. One later Phantasm lands at frame 871. That is the 5 plus 1 split.

Two cycles, 90s. Every iteration: Nefer 5 skills, 12 charges, 12 dashes; Columbina 2 skills and 2 bursts; Lauma 1 N1, 2 bursts, 2 Holds; Sucrose 4 skills and 4 dashes. Warnings: `insufficient_energy`, `skill_cd`.

Nefer hits: ordinary Charged Attack mean 1, each Phantasm component mean 11. The miss stays in the first cycle. Cycle 2's six charges are Phantasms. Gravity Interference mean 26.667 (min 25, max 30). Columbina `lunarbloom` mean 18 (min 16, max 20). Lauma `lunarbloom` mean 3. Still no Sucrose reaction row.

Waits, in seconds: Columbina energy mean 0.828 (min 0, max 1.367) and skill CD mean 0.433 (min 0, max 1.3). Lauma energy mean 6.106 (min 2.067, max 10.633). Sucrose skill CD mean 7.356 (min 3.55, max 12.3), swap CD mean 0.833, dash CD mean 0.100 (min 0, max 0.3). The second cycle does cast. It casts after those waits. Sucrose's field-time mean of 54.917s is the idle at the end of the 90s.

Sample seed 1598448435266442123 repeats the first-cycle Dew story, including the ordinary Charged Attack at frame 819. The second Hold consumes 3 Verdant Dew at frame 1374 and its damage is frame 1390. Six Phantasm damages follow from frame 1929 through 2299. A4 snapshot stays 196.3024. No Hydro Swirl in this sample.

## Instructor results

First rotation, 50s. Every iteration: Nefer 3 skills, 6 charges, 6 dashes; Columbina 2 N1, 1 skill, 1 burst; Lauma 1 N1, 1 burst, 1 Hold; Sucrose 2 skills and 2 dashes. Warnings empty. Sucrose `swap_cd` mean 0.833s. Skill CD, energy, dash, and burst waits are 0.

Nefer hits: each Phantasm component mean 6. Ordinary Charged Attack is absent. Gravity Interference mean 15. Columbina `lunarbloom` mean 13. Lauma `lunarbloom` mean 2. No Sucrose Hydro Swirl. Field time ends on Nefer, so her 40.367s includes the idle after the combo.

Sample seed 12646115053570412597:

- First Sucrose ED triggers A4 at frame 151, EM snapshot 212.3024. That is 16 above the Deepwood snapshot. Sixteen is 20% of Instructor's 80 EM 2pc. It does not trigger A1 or Hydro Swirl.
- Hold consumes 1 Verdant Dew at frame 334. Pale Hymn remaining 23 matches 18 plus 1 times 6, minus 1. Moonridge Dew is gained only after Columbina's Burst, at frames 483, 522, and 527, so the Hold cannot see it.
- `dm 4pc proc` is on Lauma (char 2). Six Phantasm consumes follow, the last of them 1 Moonridge Dew at frame 1503. No ordinary Charged Attack.

Two cycles, 100s. Every iteration: Nefer 5 skills, 12 charges, 12 dashes; Columbina 4 N1, 2 skills, 2 bursts; Lauma 1 N1, 2 bursts, 2 Holds; Sucrose 4 skills and 4 dashes. Warnings: `insufficient_energy`, `swap_cd`, `skill_cd`.

Nefer hits: each Phantasm component mean 12. Ordinary Charged Attack is absent. Both cycles' charged attacks were Phantasms. Gravity Interference mean 28.333 (min 25, max 30). Columbina `lunarbloom` mean 20.667 (min 20, max 21). Lauma `lunarbloom` mean 4. Sucrose `swirl-hydro` damage and reaction mean 1 (min 1, max 1). One Hydro Swirl across two cycles, in every iteration, not one per ED.

Waits: Lauma energy mean 2.056s (min 1.583, max 3). Sucrose skill CD mean 4.817s and swap CD mean 1.250s, both with min equal to max. Columbina records no energy or cooldown wait. Nefer field time mean 74.511s is the ending idle.

Sample seed 1038268647394553286: the opening ED again does not Swirl. At frame 1490, during the second cycle, the log has `swirl-hydro`, `sucrose a1 triggered` with reaction `swirl-hydro` and expiry 1970, and `instructor-4pc` added on all four characters with the same expiry. A4's snapshot on that frame is 236.3024, which is 24 above 212.3024. Twenty-four is 20% of 120, and 120 is the Instructor 4pc EM. The next A4, at frame 1976, is back to 212.3024, after that mod's expiry. This sample therefore shows one 8s Instructor window, not sustained uptime. A1's +50 EM is only for the swirled element, so it lands on Columbina, not on Nefer or Lauma. `skipped swirl-hydro target` is the engine's self-target guard, not a failed cast.

Silken Moon's 4pc logs `gleaming-moon-devotion-em` at frame 54 for 8s, and Lauma's Hold Lunar-Bloom react bonus includes `gleaming-moon-devotion-reaction` 0.1 while that mod is up. That is separate from the Instructor mod.

## Dew

Implemented in `pkg/core/player/verdantdew.go` and exercised by these runs:

- Lunar-Bloom calls `OnLunarBloom`, which opens a 149-frame Verdant generation window. The counter is 146 frames per Verdant Dew, cap 3. The log `verdant dew generation stopped` appears when that window ends.
- Moonridge Dew is Columbina's A4. It is added only while `columbina-q-buff` is active, on an enemy Lunar-Bloom, outside a 3-frame ICD, and while fewer than 3 have been added in the current 18s timer. The runs show the count climb to 3 and then `moonridge dew timer reset`.
- `Dew()` is Verdant plus Moonridge. `ConsumeDew` spends Verdant first, then Moonridge. Caps are 3 and 3.
- Lauma Hold requires `Dew() > 0`. At frame 29 it consumes `min(Dew(), 3)`. The Deepwood sample spending 2 Verdant and 1 Moonridge on one Hold matches that, and it matches the KQM Columbina guide's statement that Moonridge is used once Verdant is gone. It is also why the Deepwood combo is one Dew short for the sixth Phantasm.
- Nefer's charge is a Phantasm only when Shadow Dance is active and `Dew() > 0`. The Phantasm spends 1 Dew at frame 48. Three Phantasms end the dance. The ordinary Charged Attack in the Deepwood runs is the empty-Dew branch, not a skipped action.

The 60s Moonridge expiry is a TODO in `internal/characters/columbina/asc.go`. These runs do not show that removal.

Opening Nefer's Skill does not itself start Lunar-Bloom. The first Verdant Dew in the samples is frame 201, after Hydro and Dendro have met.

## Swirl, Instructor, Deepwood, EM

- Deepwood application is logged as `dm 4pc proc`: Sucrose when she holds the set, Lauma when she holds it.
- Sucrose A4 is logged. The snapshot is 20% of her non-extra EM. Instructor's 2pc is visible as the 196.3024 versus 212.3024 gap. The 4pc is visible only on the frame-1490 snapshot.
- Hydro Swirl is not produced by the Deepwood order in these runs. The Instructor order produces one Hydro Swirl per 100s simulation, on the later cycle in the sampled seed. The first ED does not Swirl.
- Instructor has no dedicated "proc" message. The status key `instructor-4pc` is the evidence, and only the sustain sample was inspected for it. The 3-iteration aggregate proves the Swirl count, not the EM-mod expiry.
- A1 does not share EM with Nefer or Lauma on a Hydro Swirl.

## Constraints that are not ready for a profile

- Second-cycle Lauma Burst waits on energy in both sustain configs. Deepwood also waits on Columbina energy and skill CD. Sucrose's skill CD is 900 frames, and Sacrificial Fragments' reset is random, so the Deepwood skill-CD wait varies by iteration.
- The action lists finish inside the chosen durations only because gcsim waits and then performs the cast. That is not the same as a tight repeated rotation.
- Cycle 1 starts from full energy. It is the optimistic case.
- Shadow Dance does not end on swap. `OnCharacterSwap` in `internal/characters/nefer/cons.go` only calls `syncC4`. The KQM Nefer guide says the state ends on swap. The ordinary Charged Attack above is not this bug: Dew was empty while the dance was still the active branch.
- Dash cancel still emits `Phantasm Performance (Nefer 2)`. The KQM guide says a dash cancel skips that hit. The frame constants are explicitly unmeasured, so this is not a validated timing bug and the frames were not changed.
- Slither movement and A4's extra Dew from Slither are not implemented. See `docs/yuheng/nefer.md`.
- Field time after the last action is idle.

## Proposed fixes, not applied

1. End Shadow Dance on swap. Call the existing `exitDance` from Nefer's swap hook in `internal/characters/nefer/cons.go`, and add a unit test that the next charge after a swap is an ordinary Charged Attack until Skill is used again. Do not change coefficients.
2. Leave the dash-cancel frame alone until a frame sheet exists. A later change would put the dash queue point before `ppN2Frame` if the KQM cancel rule is adopted. Files: `internal/characters/nefer/nefer.go`, `internal/characters/nefer/charge.go`.

## Regression

`internal/services/assets/nefer_premium_stage1_test.go` parses the four configs and runs them with seeds 1, 2, and 3. It checks the scripted action counts and the Phantasm versus ordinary Charged Attack counts above. It does not assert DPS.

Existing Nefer, Vesna, Sandrone, character-ID, ICD, player, reaction, and pipeline tests were run with this change. No generated `*.dm.go` file, character ID, or ICD index was edited. `GCSIM_ALLOW_FULL_REGEN` was not set. `task pipeline` was not used to regenerate.
