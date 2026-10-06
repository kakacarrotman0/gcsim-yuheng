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

Stellar-Conduct is not a validated damage path. The 20% bonus is snapshotted only when Precision is activated while Alyosha already has the engine's `polestar-field` status, and it applies only to the active character's `AttackTagDirectStellarConduct`. Without that status the bonus is zero. There is no Stellar-Conduct team config.

## Timing

| Timing | Class |
| --- | --- |
| Mark 15s, Precision 15s, skill CD 15s, burst CD 18s, energy 70, field 14s (20s at C2+), field pulse 2.0s with the first pulse immediate, Tugarin 0.3s after each pulse, particle ICD 0.5s, C1 18s | Ability graph / talent text |
| Normal, charged, skill, burst, and plunge cancel windows | Approximate. No frame sheet was used |

## Mechanics notes

- Tap and hold are both 1U Electro Elemental Art hits with empty attenuation, so no ICD. Both carry the mark's apply and activate tags. Tap is a 2.5 by 10 box. Hold uses the hold ratio.
- The skill generates 5 Electro particles on hit, one batch, unique ICD 0.5s. `GenerateElemBall` is `baseEnergy` 15 and config id 2020. That `baseEnergy` is same-element energy (Fischl's single particle is `baseEnergy` 3 on the same config id).
- Hunter's Mark lasts 15s on that enemy. Hitting a marked enemy with an activating attack removes the mark and grants Precision. Precision is `AttackUp` from the skill talent (index 4) as ATK%, duration 15s, active character only. It is not teamwide, not permanent, and not a snapshot buff.
- The normal-attack talent text says the final strike applies Hunter's Mark. `ConfigAbility_Avatar_Alyosha.json` puts `Alyosha_AddLockTag` / `Alyosha_TriggerLockTag` only on the two Elemental Art attacks. N4 still applies and can activate the mark, following that talent text. The ability attack configs do not show the tag.
- Burst field pulses are 1U Electro and share `ICDGroupAlyoshaElementalBurst` with Tugarin: element sequence `[1, 0]`, reset 1.6s, damage sequence all ones. The first pulse is immediate, then every 2.0s while the field still has duration left. A 14s field therefore pulses 7 times. Tugarin follows each pulse by 0.3s and does not apply gauge on that shared group. Tugarin prefers a marked enemy inside 15. At C0 he activates a mark and does not apply one. At C2 he applies or refreshes a mark on the enemy he hit and does not activate it. A second simultaneous mark target is not selected.
- A1 heals the active character for 120% of Alyosha's current ATK when Tugarin hits. C4 additionally heals the lowest HP% party member for 60% ATK. Neither heal is distance-gated.
- A4 is dynamic DMG% on his Elemental Art and Elemental Burst: `min(0.70, ER * 0.35)`, where ER 1.0 means 100% Energy Recharge. Naked level 90 ER is about 126.67% from the 26.67% ascension.
- C1 restores 15 energy every 18s on Overload, Electro-Charged, Lunar-Charged, Superconduct, Electro Swirl, Electro Crystallize, and Stellar-Conduct. Quicken, Aggravate, Spread, Bloom, Burgeon, and Hyperbloom do not trigger it. The ability names Overdose, OverdoseElectric, and OvergrowMushroomElectric have no confident gcsim event, so they are not wired.
- C6 lets Precision stack to 2. The second stack is another `AttackUp`, and the active character also gains 100 EM. C6 does not double the 20% Stellar-Conduct bonus.
- The burst taunt is not simulated. Hold interruption resistance is not simulated.

## Validation

`docs/yuheng/configs/alyosha-support.txt` is a Yuheng validation config: Alyosha / Chevreuse / Bennett / Xiangling, 20 iterations, 25s, Skill then Burst. Pyro and Electro only, so Chevreuse's team restriction is legal, and no Polestar Field is present.

There is no `alyosha-stellar-conduct.txt`. Stellar-Conduct DPS is not reported.
