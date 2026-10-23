# Nefer

Yuheng-maintained gcsim character. Upstream `genshinsim/gcsim` has no Nefer package. This is not an upstream implementation.

## Identity

| Field | Value |
| --- | --- |
| English name | Nefer |
| Key | `nefer` |
| Avatar ID | 10000122 |
| Skill depot | 12201 |
| Element | Dendro |
| Weapon | Catalyst |
| Rarity | 5 |
| Region | Nod-Krai |
| Ascension stat | CRIT DMG |
| Body | Tall female |

## Data provenance

Numbers, identity, and talent ratios are from `https://gi.yatta.moe/api/v2/en/avatar/10000122`.

Gauge rows are that payload's `advancedProps`: normals 1U on the normal-attack ICD, charged attack 1U with no ICD, skill 1U on the skill ICD, Phantasm Performance's own two hits 1U on the charged-attack ICD, shades 0U, burst 1U on the burst ICD, plunge collision 0U, low and high plunge 1U with no ICD.

Veil of Falsehood is 8% Phantasm Performance DMG per stack, 3 stacks, 9s each. That text is on the wiki talent page and is what the Yatta veil entry points at. Yatta's A1 body does not expand the percent parameter. C2's +5s, 5-stack cap, 2 stacks on skill, and 200 EM are in the Yatta constellation text. The "up to 140% of its original DMG" clause has no curve in that payload, so it is not implemented.

`excel-hk4e` in `go.mod` predates this character, so `go run ./pipeline` was not executed. The ratios in `zz_nefer.dm.go` are the Yatta rows.

## Validated scope

Moonsign is 1. Nefer sets `lunarbloom-enabled`. Her EM adds `min(EM*0.000175, 0.14)` to every Direct Lunar-Bloom hit's base-damage bonus, the same shape Lauma already uses. A party with both adds both bonuses. Transformative Bloom damage is still the engine's regular Bloom number. The lunar-bloom path here starts Verdant Dew and retags the event. It does not replace that formula.

Shadow Dance lasts 9s and ends after 3 Phantasm Performances. With at least 1 Verdant Dew, the charged attack is Phantasm Performance and costs no stamina. Otherwise a Shadow Dance charged attack costs 25 stamina, and a normal one costs 50. One Dew is consumed when the first shade is queued. Nefer's two hits are ATK plus EM. The three shade hits are EM, ignore defense, and are Direct Lunar-Bloom. C1 adds 0.6 to that EM ratio. The 8% veil bonus is a damage-percent mod on the hit, which is what also covers the C1 amount. There is no second veil multiplier on the 0.6.

Ascendant Gleam (party Moonsign at least 2) and ascension 1: the skill marks Dendro Cores, including ones created for the next 15s, as Seeds of Deceit. Those cores do not burst and do not Hyperbloom or Burgeon. Nilou's Bountiful Cores get the same mark on the following frame. A charged attack or Phantasm Performance absorbs every marked core on the field. Distance is not checked, because the sim has one player position. Three stacks, or a refresh of the third, grant 100 EM for 8s. At C2 the cap is 5, the duration is 14s, and the EM grant is 200 at 5 stacks.

C3 raises the skill. C5 raises the burst. C4, while she is on field in Shadow Dance, shortens the existing Verdant Dew counter from 146 frames to 117 and shreds 20% Dendro RES. The shred lingers 4.5s after she leaves the state or the field. The distance check is not simulated. C6 replaces her second Phantasm hit with 85% EM Lunar-Bloom and adds a 120% EM hit after the shades. Under Gleam those C6 hits, and the shades, have elevation 0.15.

The burst consumes every veil and adds the talent's per-stack damage bonus to its two hits.

## Timing

| Timing | Class |
| --- | --- |
| Energy 60, skill CD 9s, two charges, burst CD 15s, Shadow Dance 9s, veil 9s (14s at C2), EM buff 8s, deceit window 15s, C4 linger 4.5s | Talent text |
| Normal, charged, Phantasm, skill, burst, and plunge hitmarks and cancels | Placeholders. Not a measured sheet. The N3 pair is split by 2 frames. Phantasm is 30/45 for Nefer and 48/60/72 for the shades |
| Hitbox radius 4 | Placeholder circle on the primary target |
| Slither hold, the 2.5s move, and the 18.15/s stamina drain | Not simulated. The charged attack is the exit hit only |
| A4 extra Dew from Slither | Not implemented. The text gives the 10% per 100 EM over 500, capped at 50%, and does not give the base amount |
| C2 "up to 140%" | Not implemented. No curve in the Yatta constellation payload |
| Particles | Not implemented. The talent payload has no particle count |
| Interruption resistance | Not simulated |

## Validation

`docs/yuheng/configs/nefer-lunar.txt` is a constructed baseline: Nefer / Lauma / Xingqiu / Collei, Gleam on, starting Verdant Dew 3, 20 iterations, 20 seconds. It completed with no failed actions. Each iteration recorded 2 skill hits and 3 of each Phantasm Performance hit. It is not a claimed team DPS.

Unit tests cover identity, the skill ratio, Phantasm's five hits and the Dew spend, Shadow Dance ending on the third Phantasm, Seeds of Deceit ignoring Electro and being absorbed, the 100 EM grant at 3 stacks, C6's converted hit, the C4 shred, the C4 Dew counter, and the C3/C5 talent indexes.
