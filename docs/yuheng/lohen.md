# Lohen

Yuheng-maintained gcsim character. Upstream `genshinsim/gcsim` `main` has no Lohen implementation and no open Lohen pull request.

## Identity

| Field | Value |
| --- | --- |
| English name | Lohen |
| Key | `lohen` |
| Avatar ID | 10000129 |
| Skill depot | 12901 |
| Element | Cryo |
| Weapon | Polearm |
| Rarity | 5 |
| Region | Mondstadt |
| Ascension stat | CRIT DMG |
| Tag | `AVATAR_TAG_HEXENZIRKEL` |

## Data provenance

Numbers, identity, and talent text are from `https://gi.yatta.moe/api/v2/en/avatar/10000129`.

Mechanics are from `ConfigAbility_Avatar_Lohen.json` and `ConfigTalent_Lohen.json` on `DimbreathBot/AnimeGameData` commit `792978e5503ecfba73dcb3562ed44a0d35a2abe2`. Base and promotion stats are from `AvatarExcelConfigData` and `AvatarPromoteExcelConfigData` on that commit.

`excel-hk4e` in `go.mod` predates this character, so `go run ./pipeline` was not executed. Talent tables in `zz_lohen.dm.go` are hand-maintained from the Yatta proud-skill rows.

C1, C2, C4, C6, Hexerei, and the two combat passives bind `%n` parameters in `ConfigTalent`, but the numeric ProudSkill rows for those talents are not in that file. The values used are the released talent text (300% Will cap, 500% teammate Will, 500% ATK / 200 EM / 8s / 4s, 15 energy / 15s, 175% CRIT DMG / 7s, 40% / 6s / 50% Will, 15% ATK / 8s, 3000% ATK / +60 Will). The literals 1.25s and 1.65s are in the ability graph.

## Validated scope

Masterstroke, Joy, Will, Etched Into Bone and Soul, and Manifest Judgment. While Masterstroke is active, Normal, Charged, and Plunging Attack damage is Cryo and ignores weapon infusion. Outside Masterstroke those attacks stay Physical and can be infused. The released talent text is the source for the conversion. The damage handler only replaces the ratio; the Ice Falling Anthem attacks and the Ice weapon enhancement carry the element. Plunge ratios match on the normal-attack and skill sheets at the same level. During Masterstroke the plunge uses the skill-talent row, so High Spirits and C3 can raise it above the normal-attack level.

`docs/yuheng/configs/lohen-masterstroke.txt` is a constructed baseline: Lohen / Xiangling / Bennett / Kaeya, 20 iterations, 20 seconds, `hexerei` left off. Re-run after the Cryo conversion, it completed 20 iterations with no failed actions. It is not a damage benchmark.

That run's Lohen damage log, from the gcsim result JSON:

- Masterstroke normals from the four-attack string (N3 is three hits). No Physical normal source
- Lohen's own damage in that config is Cryo only
- Etched Into Bone and Soul 1–4: one each
- Manifest Judgment 1–6: one each
- Bennett and Kaeya each recorded a Melt. Lohen's converted hits land before Pyro is applied in this rotation, so Lohen records no Melt here. The unit test covers his own Melt

The printed summary was average 262404.56 damage over 20.00 seconds (13120 dps, min 11693.65, max 14356.35, std 681.68). That number is the output of this one config. It is not a claimed team DPS.

Unit tests in `lohen_test.go` check identity, Joy on a normal hit, Will from a teammate hit including the A1 extra, Will at ascension 0 without that extra, High Spirits at ascension 0, Will consumption on Etched, the `1 + Will×0.004` ratio at 50 Will, C6 keeping Will and refilling Joy, C3's skill index, the A4 ATK buff on Melt, six burst hits, Physical normals and charged attacks outside Masterstroke (including a Pyro infusion), Cryo normals, charged attacks, and plunges inside Masterstroke with infusion blocked, Cryo application, a Melt, the one-particle / 2s particle ICD, and Masterstroke ending on swap.

## Timing

| Timing | Class |
| --- | --- |
| Energy 60, skill CD 18s, burst CD 15s, Masterstroke 13s at every talent level, Joy cap 100, Will cap 100 (300 at C1) | Ability graph / talent text |
| Joy +17 per normal hit and per charged hit. Charged stamina 25, or 10 during Masterstroke | Skill talent parameters, constant across levels |
| Will on a teammate hit: +20 if the hit is at least 10× Lohen's base ATK, otherwise +1. This is the skill, including at ascension 0. A1 (promote level 1) adds +60 when the hit is at least 30× base ATK. C1 multiplies those gains by 5 and raises the cap to 300. Lohen's own hits do not add Will | Ability graph. The 30× and +60 figures are the talent text |
| Etched damage is `Raid × (1 + Will × 0.004)`. Burst damage is `Burst × (1 + Will × 0.004)`. Both snapshot Will, then Will is cleared unless C6's ICD flag is clear and Masterstroke is active | Ability graph |
| Etched hit order 0.086, 0.2095, 0.4074, 0.5316 of the raid state, played as frames 5, 13, 24, 32 | Normalized times are from the graph. The 60-frame clip length is not in the JSON. Absolute frames are unverified |
| Burst hits at frames 12, 36, 42, 48, 54, 78 | Gadget think of 0.1s, assuming the first increment is immediate. Six hits match the talent's ×6. The blank step is not a hit |
| Burst extends Masterstroke by 1.65s. The next Etched after a C6 mark extends it by 1.25s, and remaining duration is capped at 20s | Ability graph |
| Normal, charged, skill-cancel, and plunge frames | Unverified polearm spacing. Not a measured sheet |
| Particles | One Cryo particle (`baseEnergy` 3, config 2022) when an Ice hit with a normal, charged, plunge, or elemental-art tag lands. Unique modifier 2.0s |

## Mechanics notes

- Entering Masterstroke clears Joy, Will, the etched counter, and the C6 mark. The counter reset is `Lohen_CountExtraE_Report` when the mode flag rises.
- Leaving the field ends Masterstroke before the 13s timer. `GrandHandler` `onAvatarOut` removes the modifier, and `onRemoved` zeroes Joy, Will, and the C6 mark. Etched is then unavailable. Swapping back does not resume the mode. The etched counter itself is not cleared until the next Masterstroke, matching the graph.
- A normal or charged hit chooses Physical or Cryo when it lands. A plunge chooses when the action starts; its cancel window is after the impact, so a later swap cannot change that hit.
- Etched is `skill[etched=1]`. It requires Masterstroke, Joy at 100, and remaining uses (3, or 5 at C6). It does not restart the 18s skill CD.
- A plain skill while Masterstroke is already active restarts the mode. That matches `RemoveUniqueModifier` plus `AttachModifier` on the skill start.
- Etched and Burst consume Will after the snapshot. C6 skips that clear only while its 7s flag is down. The talent sentence can be read as "never consume." The graph ties the skip to the same flag that gates the Joy refill. This implementation follows the graph.
- C4, on entering Masterstroke, refunds 15 energy when energy is not full. When it is full, the next Burst within 15s refunds 15. A Burst during Masterstroke sets Will to its cap before the snapshot.
- C2 arms Evilsbane for 4s after the fourth Etched hit or the last burst hit, ICD 4s. The next normal or charged hit during Masterstroke deals 500% ATK Cryo in a radius-3 circle (the radius was not in the extracted pattern; Yatta's gauge note is 1U) and gives other members 200 EM for 8s.
- C6 adds 175% CRIT DMG to hits whose names start with "Etched" or "Manifest" while Masterstroke is active. Evilsbane does not get that bonus.
- A4 listens for Melt, Freeze, Superconduct, Cryo Swirl, and Cryo Crystallize from another member during Masterstroke, then gives that member and Lohen 15% ATK for 8s. Stellar Swirl and Stellar Superconductor are in the ability list and are not wired.
- High Spirits adds 1 skill level for 9s, plus 6s when another member's normal, skill, or burst level is at least Lohen's skill level, ICD 18s. Proud skill 12923 has no promote level on skill depot 12901. Proud skills 12921 and 12922 are promote levels 1 and 4. An omitted level is 0, so High Spirits is available at ascension 0.
- Hexerei is off unless the character param `hexerei=1` is set. With at least two Hexerei characters and Will at least half of its cap when Etched's last hit lands or Burst is cast, normal and charged attacks gain 40% DMG for 6s.
- The quest sneak bullet is not implemented.

## Not claimed

One 20-iteration config does not prove every team or constellation. Raid and burst hit frames are not a measured sheet. The 0.1s Joy and Will gates are not simulated; the placeholder normal spacing is already outside 0.1s, and the two charged hits use separate Joy modifiers in the graph. The 40m Will distance check is not simulated. ProudSkill numeric rows for the constellations and passives were not in the readable talent config, so those magnitudes follow the talent text. Stellar Swirl and Stellar Superconductor stay unwired.
