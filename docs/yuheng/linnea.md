# Linnea

Yuheng-maintained gcsim character. Upstream `genshinsim/gcsim` `main` does not contain her. Open upstream PR `#2596` (head `ef71cdb05e55c649d88ecf7a7eae21cf09ebb8c`) was used as a draft and then corrected. This is not that PR.

## Identity

| Field | Value |
| --- | --- |
| English name | Linnea |
| Key | `linnea` |
| Avatar ID | 10000130 |
| Skill depot | 13001 |
| Element | Geo |
| Weapon | Bow |
| Rarity | 5 |
| Region | Nod-Krai |
| Ascension stat | CRIT Rate |
| Tag | `AVATAR_TAG_MOONPHASE` |

## Data provenance

Numbers, identity, and talent text are from `https://gi.yatta.moe/api/v2/en/avatar/10000130`.

Mechanics are from `ConfigAbility_Avatar_Linnea.json` and `ConfigTalent_Linnea.json` on `DimbreathBot/AnimeGameData` commit `792978e5503ecfba73dcb3562ed44a0d35a2abe2`. Base and promotion stats are from `AvatarExcelConfigData` and `AvatarPromoteExcelConfigData` on that commit. Yatta's live talent rows match the skill tables checked against that commit.

`excel-hk4e` in `go.mod` predates this character, so `go run ./pipeline` was not executed. Talent tables in `zz_linnea.dm.go` are the upstream draft's Yatta rows.

## Validated scope

Combat kit at C0, plus the constellation code paths covered by unit tests. Lumi's Pound-Pound is Geo and can Crystallize. Heavy Overdrive and Million Ton Crush are direct Lunar-Crystallize. The party flag `lunarcrystallize-enabled` is set, and Lunar-Crystallize base DMG gains `min(DEF/100*0.007, 0.14)`.

`docs/yuheng/configs/linnea-lunar.txt` is a constructed baseline: Linnea / Xingqiu / Bennett / Zhongli, 20 iterations, 20 seconds. It completed 20 iterations with no failed actions. It is not a damage benchmark.

That run's damage log, from the gcsim result JSON:

- Lumi Pound-Pound Pummeler 1 and 2: 8 each
- Lumi Heavy Overdrive Hammer: 4
- Linnea Lunar-Crystallize reactions: mean 5
- Zhongli Lunar-Crystallize reactions: mean 7

The printed summary was average 301660.02 damage over 20.00 seconds (15083 dps, min 14291.30, max 15971.55, std 469.55). That number is the output of this one config. It is not a claimed team DPS.

Unit tests in `linnea_test.go` check identity and Moonsign, C3/C5 talent indexes, the burst heal formula and 12 ticks, 3 Geo particles, and C2 applying to both a Geo and a Hydro teammate when an Anemo character sits between them.

## Timing

| Timing | Class |
| --- | --- |
| Energy 60, skill CD 18s, burst CD 15s, Lumi duration 25s, particle ICD 9s, 3 Geo particles (`baseEnergy` 9, config 2023) | Ability graph / talent text |
| Burst initial heal on the heal modifier's `onAdded`. Continuous heal think interval 1.0s, duration 12s, so 12 ticks. The initial heal is queued at frame 1 because the cast-to-attach delay was not extracted | Ability graph. Tick count is duration/interval |
| Lumi cadence (`skillSuperInterval` 62/122/90, recast window 0.6s, second pound +21f, Million Ton at 50f, standard interval 120f, burst-summon first hit 214f) | Approximate. Taken from the upstream draft, which marks the skill frames as unknown. Not a measured sheet |
| Normal hitmarks 15/12/47 and aimed hitmarks 15/86 | Placeholders from that same draft (`TODO: implement frames`). Not a measured sheet |
| Plunge cancel windows | Approximate bow template. The draft had no plunge |

## Mechanics notes

- `Moonsign` is 1. Alyosha stays at 0.
- C3 is `SkillCon` and C5 is `BurstCon`. The upstream draft set `NormalCon` to 3, which would have raised normal attacks instead of the skill.
- Heal amounts are flat plus DEF ratio. The draft's formula multiplied the flat amount by DEF. Talent 10 initial heal is 1694.9546 + 2.88×DEF. Each tick is 338.9909 + 0.576×DEF. Healing uses current DEF when the heal happens.
- Refreshing an existing Lumi extends her status and does not replace `skillSrc`. Replacing it stops the live ticker.
- C2's party loop uses `continue` for non-Hydro/Geo characters. The draft returned, which skipped the rest of the party.
- C6 grants 18 Field Catalog stacks from the skill as well as from Moondrift Harmony, consumes twice as many stacks, and multiplies that flat damage by 1.5. The 25% elevation bonus is checked when the Lunar-Crystallize hit happens, and only if `GetMoonsignLevel()` is at least 2.
- A1 shreds 15% Geo RES near Lumi, or 30% while Ascendant Gleam is active. A4 gives 5% of Linnea's DEF as EM to the active character when that character is Moonsign, and to Linnea otherwise.
- Heavy Overdrive is one hit. The ability has two attack tags on that skill and zero `TriggerAttackEvent` entries, so the second tag is not modeled.
- Lumi is not a target. Her attacks use the primary enemy's position.
- Master Adventurer (wildlife and ore) is out of combat and is not implemented.
- Interruption resistance on the skill is not simulated.

## Not claimed

One 20-iteration config does not prove every team or constellation. Gleam was not part of the validation team. Normal, aimed, skill, and plunge frames are not measured.
