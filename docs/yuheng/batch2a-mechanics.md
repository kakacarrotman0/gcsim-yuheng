# Foundation batch 2A

Mechanics correction and deterministic coverage on `cursor/batch4-engine-integration`. No character ID, ICD index, engine pin, or pipeline gate changed. Production simulation traffic was not run.

## Zibai Stride

`docs/yuheng/configs/zibai-c0.txt` is unchanged and still does not cast Stride.

`docs/yuheng/configs/zibai-stride.txt` is solo Zibai, talent 1, Favonius Sword. The script is Skill, `wait(500)`, Skill. Nothing writes Radiance. The ticker reaches 70 at frame 432. The second Skill is after that. Seed 1 records two skill actions, one `Spirit Steed's Stride` hit, one `Spirit Steed's Stride (Lunar-Crystallize)` hit, and a debug line `Consumed 70 radiance`. That is one spend of 70 and one Direct Lunar-Crystallize hit. It is not a team DPS result.

The A1 4s window still expires before a ticker-only 70. This config does not claim the A1 flat damage is active on that hit.

## Nefer Shadow Dance

KQM's Nefer quick guide says Shadow Dance ends when she swaps out, and that teammates should act before she takes the field. The talent card lists the three-Phantasm and duration exits and does not mention the swap. This batch follows the quick guide.

`c4Init` already subscribed to `OnCharacterSwap`. The handler now calls `exitDance` when the previous index is Nefer, then `syncC4`. `exitDance` clears the status, the source counter, and the Phantasm count, and refreshes C4. A swap between other characters does not clear the state. Coming back on field does not restore it. The next Charged Attack with Dew still in hand is an ordinary Charged Attack until Skill is used again. Natural expiry is the existing 9s task. Three Phantasms still end it, and the charge after that is ordinary.

C4 haste still ends on that swap. The Dendro shred still lingers 4.5s. The Dew counter after the swap is still 147 frames in the existing test.

A Phantasm's hits and Dew spend are queued when the charge starts. Swapping afterward does not cancel those tasks. The shade still lands.

Dash-cancel frames were not changed. Slither and the 60s Moonridge expiry were not implemented.

`nefer-natural.txt` and `nefer-lunar.txt` were not edited. Natural still has one of each Phantasm hit and no ordinary Charged Attack, with no failed actions, on seeds 1 through 20. Lunar still has three of each Phantasm hit and no ordinary Charged Attack. Lauma's swap-CD mean stays 0.4 with or without the dance fix. The Stage 1 premium configs were re-run on seeds 1, 2, and 3. Their pinned action, hit, and reaction means did not move.

## Lunar-Crystallize

No file under `pkg/reactable` was edited.

Without the enable flag, Hydro plus Geo is `OnCrystallizeHydro` and spawns no Moondrift construct. After Zibai's init, the same pair is `OnLunarCrystallize`. Three such reactions spawn Moondrift constructs, emit one Moondrift Harmony, and the delayed damage instance is `lunarcrystallize` with `AttackTagReactionLunarCrystallize`, Geo, and full DEF ignore. A 0-gauge hit tagged `AttackTagDirectLunarCrystallize` stays on that direct tag.

Zibai and Linnea each add `min(DEF/100*0.007, 0.14)` from their own DEF. With both DEF values past the cap, the direct hit and the reaction contribution each carry 0.28. The caps are independent. The bonuses add. Columbina is not in that party, so her HP bonus is not part of the 0.28.

`CalcSpecialReactionDmg` still multiplies Lunar-Crystallize by 1.6. That number was already in the file. It was not re-derived.

## Lunar-Charged

No file under `pkg/reactable` was edited.

Xingqiu plus Kaeya, with no Ineffa, Flins, or Columbina, does not set the enable flag. Hydro plus Electro emits `OnElectroCharged` and does not open the cloud. Ineffa plus Xingqiu sets the flag. The same pair emits `OnLunarCharged`, opens `lunarcharged-cloud`, and the tick's damage is `lunarcharged` with `AttackTagReactionLunarCharge`, Electro, and full DEF ignore. The tick reduces Hydro and Electro. Ineffa's existing ATK bonus, `min(ATK/100*0.007, 0.14)`, is present on that reaction contribution. Frozen still rejects both Electro-Charged and Lunar-Charged.

The checked-in Lunar-Charged multiplier is 3. It was not re-derived.

At the end of batch 2A, `LcExtraHitOverride` and `LcrExtraHitOverride` were both the string `lunarcrystallize-bonus-hit-chance`. Batch 2B gives them distinct keys. See `docs/yuheng/batch2b-mechanics.md`.

Other checked-in assumptions left as written:

- `TryAddLC` still attaches Hydro in the reaction function. The comment says that should not happen there.
- The cloud tick still uses the player's current position and an 8-unit radius. The comment says the area should be saved when the cloud is created.
- Ineffa's A4 status duration is marked TODO. Flins uses the same enable flag and the same ATK bonus expression. Those two bonuses would add if both characters were present. That pair was not the representative team.

## Not in this batch

Character IDs, ICD indexes, the pipeline regeneration gate, engine pins, backend, frontend, and production traffic were left alone. No new reaction formula was added.
