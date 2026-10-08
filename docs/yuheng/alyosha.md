# Alyosha

Yuheng-maintained gcsim character. Upstream `genshinsim/gcsim` `main` does not contain him. This is not an upstream implementation.

## Identity

| Field | Value |
| --- | --- |
| English name | Alyosha |
| Key | `alyosha` |
| Avatar ID | 10000148 |
| Element | Electro |
| Weapon | Polearm |
| Rarity | 4 |
| Region | Snezhnaya |
| Ascension stat | Energy Recharge |
| Release | unix 1786395600 |

## Data provenance

Numbers, identity, and talent text are from `https://gi.yatta.moe/api/v2/en/avatar/10000148`.

Mechanics are from `ConfigAbility_Avatar_Alyosha.json`, `ConfigTalent_Alyosha.json`, and `AttackAttenuationExcel` on `DimbreathBot/AnimeGameData` commit `792978e5503ecfba73dcb3562ed44a0d35a2abe2`.

`excel-hk4e` in `go.mod` predates this character, so `go run ./pipeline` was not executed. Talent tables in `zz_alyosha.dm.go` are hand-maintained from the Yatta proud-skill rows.

Playstyle cross-check: off-field support, Skill then Burst. A text fetch of the KQM quick guide returned HTTP 403, so the rotation shape is the one specified for this batch and was not re-scraped.

## Validated scope

General support only. Skill applies Hunter's Mark. A second marked hit activates Hunter's Precision, which is an ATK% buff on the active character. Burst creates the Fulgurite Hunting Field and Tugarin. A1 heals the active character when Tugarin hits. A4 scales his skill and burst damage from his current Energy Recharge.

Stellar-Conduct is not a validated team-damage path. There is no Stellar-Conduct team config. The bonus itself is snapshotted when Precision is activated: 20% at one stack, and 40% at C6's second stack, only if Alyosha already has the engine's `polestar-field` status at that moment. It applies only to the active character's `AttackTagDirectStellarConduct`. Without that status the snapshotted bonus is zero.

Alyosha is not a Moonsign character. `Moonsign` stays 0, so he adds nothing to `GetMoonsignLevel()` and cannot arm Ascendant Gleam.

## Timing

Frame counts below are medians from caramielle's sheet `https://docs.google.com/spreadsheets/d/1UWavwdGOIy1jmfrIHgKLekIe23OKbxY18Pt2KewA1_Q` (video `https://youtu.be/VUiIhcBM9v4`), except where a cell is called out.

| Timing | Class |
| --- | --- |
| Mark 15s, Precision 15s, skill CD 15s, burst CD 18s, energy 70, field 14s (20s at C2+), particle ICD 0.5s, C1 18s | Ability graph / talent text |
| Field pulse every 2.0s (120f) | Ability `Alyosha_ElementalBurst_Material.thinkInterval`. The sheet's tick-to-tick counter clusters at 118 (range 117–119). That is kept at 120. It is not replaced with upstream's 118. |
| Tugarin damage 0.6s (36f) after each pulse | Ability graph. The field think applies `Avatar_Alyosha_Delay` (0.3s). When that modifier is removed it applies `Avatar_Alyosha_ElementalBurst_AttackLogic` (0.3s). The dog's `TriggerAttackEvent` is that second modifier's `onRemoved`. The sheet's bite hitmarks average about 39f. That measured hitmark is not used in place of the 0.6s chain, and upstream's 39 is not copied. |
| First field tick 79, burst energy drain 7, burst CD frame 1 | Sheet. First tick trials 79/79/78. Energy trials 5/7/8. |
| Burst cancels: N1 54, skill 53, dash 55, jump 54, walk 54, swap 52, animation 55 | Sheet medians. Dash 54/55/55, jump 54/54/57, walk 52/54/54, swap 49/52/52. |
| Normals: hitmarks N1 13, N2 14, N3 17 then +15, N4 17 | Sheet medians of 14/13/13, 15/14/14, 18/16/18, and 16/18/17. N3-2 is 15 on every trial, counted from the N3-1 hit. |
| Normal cancels: N1 walk 34 / N2 18 / CA 24; N2 walk 38 / N3 23 / CA 26; N3 walk 65 / N4 53 / CA 54; N4 walk 62 / N1 56 | Sheet medians. N2 into N3 is the integer midpoint of 24 and 21; the sheet drops trial 3 as an outlier. Skill, burst, dash, jump, and swap cells are `x`, so those cancels stay at the hitmark. |
| Normal hitlag and hitboxes | Sheet companion spec on the same workbook: N1 circle r2 offset (0, 0.6), halt 0.06 scale 0.01 defense halt; N2 box 2×4 offset (0.2, 0), same hitlag; N3-1 circle r1.8 offset (0, 0.8); N3-2 box 2×4 offset (0, -0.5); N4 box 2.4×6 offset (0, 0.5). N3 and N4 have no hitlag row. |
| Charged attack hitmark 22, N1 63, skill 62, burst 63, walk 65, swap 61, dash/jump at the hit | Sheet. Hitmark trials 23/22/22. Dash and jump are "anytime". Hitbox is the bullet circle r0.8, halt 0, scale 0.01, defense halt, deployable. |
| Tap skill hitmark 18, CD 16, cancels N1 30, burst 30, dash 31, jump 31, walk 31, swap 30 | Sheet. Hitmark 18/18/18. CD trials 6/16/16. |
| Hold skill hitmark 113, CD 111, cancels N1 138, burst 138, dash 138, jump 138, walk 142, swap 137 | Sheet. Hitmark 113/114/113. |
| Plunge cancel windows | Approximate. The exported sheet has no plunge tab. |

## Mechanics notes

- Tap and hold are both 1U Electro hits with empty attenuation, so no ICD. Both carry the mark's apply and activate tags. Tap is `AttackTagElementalArt` and the ability box, x 2.5 by z 10. Hold is `AttackTagElementalArtHold` and the ability fan, radius 15 and 150 degrees. The ability file tags the hold attack `Elemental_Art`; the separate hold tag is the gcsim split, and A4 lists both.
- The skill generates 5 Electro particles on hit, one batch, unique ICD 0.5s. `GenerateElemBall` is `baseEnergy` 15 and config id 2020. That `baseEnergy` is same-element energy (Fischl's single particle is `baseEnergy` 3 on the same config id).
- Hunter's Mark lasts 15s on that enemy. Hitting a marked enemy with an activating attack removes the mark and grants Precision. Precision is `AttackUp` from the skill talent (index 4) as ATK%, duration 15s, active character only. It is not teamwide, not permanent, and not a snapshot buff.
- The normal-attack talent text says the final strike applies Hunter's Mark. `ConfigAbility_Avatar_Alyosha.json` puts `Alyosha_AddLockTag` / `Alyosha_TriggerLockTag` only on the two Elemental Art attacks. N4 still applies and can activate the mark, following that talent text. The ability attack configs do not show the tag.
- Burst field pulses are 1U Electro and share `ICDGroupAlyoshaElementalBurst` with Tugarin: element sequence `[1, 0]`, reset 1.6s, damage sequence all ones. The first pulse is the sheet hitmark at 79f, then every 2.0s while the field still has duration left. A 14s field therefore pulses 7 times. Tugarin follows each pulse by 0.6s and does not apply gauge on that shared group. Tugarin prefers a marked enemy inside 15, then any nearby enemy. That preference is kept. At C0 he activates a mark if one is present and does not apply a new one. At C2 he does that activation first, then applies or refreshes a mark. A marked target therefore ends the bite still marked, with Precision granted. An unmarked target gains a mark and does not grant Precision. A second simultaneous mark target is not selected.
- A1 heals the active character for 120% of Alyosha's current ATK when Tugarin hits. C4 additionally heals the lowest HP% party member for 60% ATK. Neither heal is distance-gated.
- A4 is dynamic DMG% on `AttackTagElementalArt`, `AttackTagElementalArtHold`, and `AttackTagElementalBurst`: `min(0.70, ER * 0.35)`, where ER 1.0 means 100% Energy Recharge. Naked level 90 ER is about 126.67% from the 26.67% ascension. `PermanentSkill_2` lists `Elemental_Art` and `Elemental_Burst`. Hold is included because that attack is `Elemental_Art` in the ability file.
- C1 restores 15 energy every 18s. `Alyosha_Constellation_1_Energy_Recover` is a `DoActionByElementReactionMixin` whose reaction list is exactly `Explode`, `Shock`, `MoonShock`, `Superconductor`, `SwirlElectric`, `CrystallizeElectric`, `Overdose`, `OverdoseElectric`, `OvergrowMushroomElectric`, and `StarSuperconductor`. Those are wired as Overload, Electro-Charged, Lunar-Charged, Superconduct, Electro Swirl, Electro Crystallize, and Stellar-Conduct. The list does not contain Quicken, Aggravate, or Hyperbloom, so those three stay unwired. `Overdose`, `OverdoseElectric`, and `OvergrowMushroomElectric` still have no confident gcsim event. Spread, Bloom, and Burgeon are also absent from that list.
- C6 lets Precision stack to 2. The second stack is another `AttackUp`, the active character gains 100 EM, and the snapshotted Stellar-Conduct bonus becomes 40% when the Polestar Field requirement was met at that activation. One stack stays at 20%. The snapshot is the ability's own order: `Avatar_Alyosha_LockTag_TriggerCD.onAdded` writes `StarConduct_DamageUp_Calc` to `StarConduct_DamageUp` only when `_IS_IN_STARMODE` is set, and writes 0 otherwise. The team-buff property then reads that written value. It is not rechecked when the Stellar-Conduct damage happens.
- The burst taunt is not simulated. Hold interruption resistance is not simulated.

## Validation

`docs/yuheng/configs/alyosha-support.txt` is a Yuheng validation config: Alyosha / Chevreuse / Bennett / Xiangling, 20 iterations, 25s, Skill then Burst. Pyro and Electro only, so Chevreuse's team restriction is legal, and no Polestar Field is present.

There is no `alyosha-stellar-conduct.txt`. Stellar-Conduct DPS is not reported.
