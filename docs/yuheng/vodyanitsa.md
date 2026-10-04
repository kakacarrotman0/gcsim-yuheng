# Vodyanitsa

Yuheng-maintained gcsim character. Upstream `genshinsim/gcsim` v2.48.8 and `main` do not contain her.

## Identity

| Field | Value |
| --- | --- |
| English name | Vodyanitsa |
| Key | `vodyanitsa` |
| Avatar ID | 10000140 |
| Element | Hydro |
| Weapon | Catalyst |
| Rarity | 5 |
| Region | Snezhnaya |
| Ascension stat | HP% |
| Release | 2026-09-21 (unix 1790024400), live in the v7.1.0 datamine |

## Data provenance

Numbers are from the live proud-skill and avatar excel used by the gcsim pipeline source `github:iam-akuzihs/excel/live` at commit `a8763e1359d11b82ba3aac4902be544c7cf14c07` (`v7.1.0`), cross-checked against `https://gi.yatta.moe/api/v2/en/avatar/10000140`.

Mechanics are from `ConfigAbility_Avatar_Vodyanitsa.json` and `ConfigTalent_Vodyanitsa.json` on `DimbreathBot/AnimeGameData` commit `792978e5503ecfba73dcb3562ed44a0d35a2abe2`.

`excel-hk4e` in `go.mod` predates this character, so `go run ./pipeline` was not executed. Talent tables in `talents.go` are copied from those ProudSkill rows. A full pipeline run would also rewrite every other character.

## Implementation status

C0–C6 combat kit is implemented: normals, charged attack, plunge, skill, horns, song healing, Hydro/Cryo RES shred, particles, burst, A4 Lead Vocal / Chorus, A1 Wandering Vortex flag, and constellations.

Not simulated: out-of-combat swim, the utility “sings instead of playing an instrument” passive, and song interruption resistance.

## Timing

| Timing | Class |
| --- | --- |
| Horn interval 3s, horn hit +0.2s, heal pre-delay 0.2s then every 1.5s, particle ICD 0.2s, RES 6s, song 16s (25s at C2+), concerto 30s, C1 5s, C2 5s, C4 6s, Anemo shred 6s, alter window 5s | Verified from the ability graph |
| NA / CA / plunge / skill-cast / burst hitmarks | Approximate placeholders. No frame sheet was used. They are not copied from another character. |

## Mechanics notes

- Skill, horn, burst, heal, A4, and C1 all read current Max HP at the moment they resolve.
- Burst song bonus: the ability default `DamageRatio` is 1.0 and `ConfigTalent` applies the talent percent as `paramDelta`. While Song of Ages Past is active the hit is `ratio * MaxHP * (1 + bonus)`. Otherwise it is `ratio * MaxHP`.
- Horns are the only particle source. The bullet attack type is Range, and the particle action requires that type. Each horn's `GenerateElemBall` has `baseEnergy` 3 and `ratio` 1. The drop config id 2018 is not in the gadget excel, so the Hydro element is taken from the caster. The initial skill hit does not generate particles.
- C4's heal is calculated from Max HP before the new Max HP stack is applied. The stack is added after that heal.
- Skill and horns share `ICDTagElementalArt` / `ICDGroupDefault` (1U). Burst is 2U with no ICD. `elementRank` 3 is not represented.
- A4's flat bonus is zero at or below 40000 Max HP. A naked level 90 Vodyanitsa is below that line (about 19085 HP after ascension). The bonus is `(MaxHP - 40000) * perKilo / 1000`, capped at 3500 Hydro/Cryo (140 per 1000) and 6500 Stellar (260 per 1000), both capping at 65000 HP.
- A4 consumes one Lead Vocal stack for the active character and one Chorus stack for anyone else. Hydro and Cryo talent hits receive the flat on `OnEnemyHit`, so it is included in base damage before DMG%, defense, resistance, and crit. While a Wandering Vortex is present, or for 5s after it detonates, the flat is added on `OnSpecialReactionAttack` instead, which is the stellar formula's `FlatDmg` term (before elevation and the contributor crit). The queued swirl attack does not consume a second stack. The ability file also has a normal-damage mixin on the same stellar tags; gcsim has one stellar damage channel, so the flat is applied once.
- Wandering Vortex has no separate damage config in her ability file. Generation and detonation apply 35% Anemo RES shred for 6s. The vortex’s own damage and gauge stay the existing Stellar Vortex.
- Entering Radiance: Stellar Swirl during the song uses 12s instead of 8s. A swirl that only refreshes an existing Radiance status stays at 8s. This is wired through Odette, Qiqi, and Cryo Traveler.

## Validation

Configs under `docs/yuheng/configs/` are Yuheng validation configs, not public gcsim samples.

`vodyanitsa-skirk.txt` is Vodyanitsa / Skirk / Escoffier / Furina. That team is legal on this baseline: Hydro and Cryo only, her skill shreds both, and the song heals the active character. Stellar Swirl does not occur without Anemo, so A1’s vortex conversion is idle in that config.

Representative run of that config (C0, level 90, talents 10/10/10, 100 iterations, 90s, no failed actions):

- Team DPS: 50936
- Vodyanitsa personal DPS: 4360

`vodyanitsa-minimal.txt` (C0, talents 1/1/1, Emerald Orb, 20 iterations, 30s) completed at 1032 DPS. That run only checks that she can be initialized and can act.
