# Kachina

Yuheng-maintained gcsim character. Upstream `genshinsim/gcsim` `main` does not contain her. This is not an upstream implementation.

## Identity

| Field | Value |
| --- | --- |
| English name | Kachina |
| Key | `kachina` |
| Avatar ID | 10000100 |
| Element | Geo |
| Weapon | Polearm |
| Rarity | 4 |
| Region | Natlan |
| Ascension stat | Geo DMG Bonus |
| Release | unix 1724706000 |

## Data provenance

Numbers, identity, and talent text are from `https://gi.yatta.moe/api/v2/en/avatar/10000100`.

Mechanics are from `ConfigAbility_Avatar_Kachina.json` on `DimbreathBot/AnimeGameData` commit `792978e5503ecfba73dcb3562ed44a0d35a2abe2`.

`excel-hk4e` in `go.mod` predates this character, so `go run ./pipeline` was not executed. Talent tables in `zz_kachina.dm.go` are hand-maintained from the Yatta proud-skill rows.

Playstyle cross-check: off-field Scroll support. Skill, optional Burst, swap, Turbo Twirly keeps attacking. A text fetch of the KQM quick guide returned HTTP 403, so the rotation shape is the one specified for this batch and was not re-scraped.

## Validated scope

Off-field support. Tap skill enters Nightsoul's Blessing at 60 points, summons an independent Turbo Twirly, and that summon keeps slamming after Kachina swaps out. Burst, C1–C6 combat effects, and a limited mount/dismount path are implemented and unit-tested. Mounted movement is not a damage rotation.

## Timing

| Timing | Class |
| --- | --- |
| First independent slam at 132 frames (2.0s think + 0.2s attack delay), then every 120 frames | Ability graph |
| Skill CD 20s, burst CD 18s, energy 70, field 12s, C1/C6 ICD 5s, A1 12s, particle ICD 0.2s | Ability graph / talent text |
| Normal, charged, skill-cancel, burst hitmark 35, and mounted hitmarks | Community sheet (caramielle), cited by open gcsim PR #2697. Not an independent capture |
| Plunge cancel windows | Standard polearm template. Approximate |

## Mechanics notes

- Independent and mounted slams are Nightsoul-aligned Geo, 1U, `ICDTagElementalArt` / `ICDGroupDefault`, DEF ratio, blunt. A4 adds `0.2 * current DEF` as flat damage on those Elemental Art hits, read when the slam is queued.
- Each slam generates one Geo particle. The graph's `GenerateElemBall` uses `baseEnergy` 3 and config id 2023. That `baseEnergy` is same-element energy, matching Fischl Oz (config 2020, `baseEnergy` 3, one particle). Config 2023 is not a gadget-excel row. One slam can produce only one batch; the unique ICD is 0.2s.
- Points: skill grants 60. Each slam spends 10. Dismount spends 2. The summon ends when the next slam cannot be paid. Phlogiston continuation after points hit 0 is not simulated, so she does not keep attacking on phlogiston.
- Hold starts mounted. A later skill press toggles mount. Dismount restarts the 132-frame independent timer. Mounted movement drain and the mounted normal-attack hitmarks exist, but the validated path is the unmounted summon.
- Turbo Twirly is registered as a limited Geo construct (`GeoConstructKachinaSkill`). Recreating it replaces her previous one. It does not expire on a timer; it is destroyed when the gadget ends, including when the three-construct limit evicts it. Slam radius is 4, or 5.2 while her Turbo Drill Field is up. Field movement speed is not simulated.
- Burst is a 1U Geo DEF hit with no ICD, radius 6. It removes the current gadget, then at the hitmark creates the 12s field and, if she still has points and was not mounted, summons an independent Twirly on the 132-frame timer. C2 grants 20 Nightsoul points (entering the blessing if she was not in it) before that check, which is what lets a bare Burst summon Twirly.
- A1 is 20% Geo DMG Bonus on Kachina herself for 12s after a Nightsoul Burst. It is not a team DMG% buff.
- C1 restores 3 energy on a Crystallize shield gain and on Lunar-Crystallize, once per 5s. It is not gated on Twirly. The mount/dismount shard vacuum is not simulated, and shard pickup is not auto-granted; the energy follows the shield and Lunar-Crystallize events the sim already emits.
- C4 is 8/12/16/20% DEF for 1/2/3/4+ enemies while the active character is within 5.2 of the field center. The bonus is dynamic and cleared on swap. 5.2 is the in-field slam radius, not a separately extracted gadget radius.
- C6 deals 200% DEF Geo, radius 6, no ICD, once per 5s, when an active-character or party shield is replaced or broken. The first shield gain does not trigger it.
- Night Realm phlogiston, Natlan transmission, stamina-on-climb, and the minimap utility passive are not simulated.
- Slams are centered on the active character, which is the usual gcsim stand-in for a placed AoE.

## Validation

`docs/yuheng/configs/kachina-support.txt` is a Yuheng validation config: Kachina / Navia / Bennett / Xiangling, 20 iterations, 20s. Kachina taps skill and swaps. Burst is intentionally not cast.

Direct CLI checks also covered a solo skill-and-burst config and a C2 burst config. Those runs are described in the batch handoff, not as a DPS benchmark.
