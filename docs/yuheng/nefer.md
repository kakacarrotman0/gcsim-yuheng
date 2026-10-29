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

Veil of Falsehood is 8% Phantasm Performance base DMG per stack, 3 stacks, 9s each. KQM calls that 8% a Base DMG Multiplier. Yatta's A1 body does not expand the percent; the percent is the wiki and KQM text the veil entry points at. C2's +5s, 5-stack cap, 2 stacks on skill, and 200 EM are in the Yatta constellation text. Five stacks are 40% base DMG, which is the talent's "up to 140% of its original DMG" clause. There is no second 1.4 multiplier.

`excel-hk4e` in `go.mod` predates this character, so `go run ./pipeline` was not executed. The ratios in `zz_nefer.dm.go` are the Yatta rows.

## Validated scope

Moonsign is 1. Nefer sets `lunarbloom-enabled`. Her EM adds `min(EM*0.000175, 0.14)` to every Direct Lunar-Bloom hit's base-damage bonus, the same shape Lauma already uses. A party with both adds both bonuses. Transformative Bloom damage is still the engine's regular Bloom number. The lunar-bloom path here starts Verdant Dew and retags the event. It does not replace that formula.

Shadow Dance lasts 9s and ends after 3 Phantasm Performances or when Nefer leaves the field. The KQM quick guide is the source for the swap rule. Hits and the Dew spend already queued before that swap still resolve. Returning does not restore the state; Skill opens it again. With at least 1 Verdant Dew, the charged attack is Phantasm Performance and costs no stamina. Otherwise a Shadow Dance charged attack costs 25 stamina, and a normal one costs 50. One Dew is consumed when the first shade is queued. Nefer's two hits are ATK plus EM. The three shade hits are EM, ignore defense, and are Direct Lunar-Bloom. C1 adds 0.6 to that EM ratio. Each Veil stack then adds 0.08 to `BaseDmgBonus` on every Phantasm hit, including the C1 amount and the C6 hits, so it multiplies that base once. On the two ordinary hits the EM flat is scaled by the same factor, because the engine applies `BaseDmgBonus` to the ATK product and adds `FlatDmg` afterwards. The 8% is not also added as DMG Bonus. Direct Lunar-Bloom still ignores DMG Bonus, which is why a Dendro DMG Bonus does not move the shades. C6 elevation stays a separate 1.15 under Gleam.

Ascendant Gleam (party Moonsign at least 2) and ascension 1: the skill marks Dendro Cores, including ones created for the next 15s, as Seeds of Deceit. Those cores do not burst and do not Hyperbloom or Burgeon. Nilou's Bountiful Cores get the same mark on the following frame. A charged attack or Phantasm Performance absorbs every marked core on the field. Distance is not checked, because the sim has one player position. Three stacks, or a refresh of the third, grant 100 EM for 8s. At C2 the cap is 5, the duration is 14s, and the EM grant is 200 at 5 stacks.

C3 raises the skill. C5 raises the burst. C4, while she is on field in Shadow Dance, shortens the existing Verdant Dew counter from 146 frames to 117 and shreds 20% Dendro RES. The shred lingers 4.5s after she leaves the state or the field. The distance check is not simulated. C6 replaces her second Phantasm hit with 85% EM Lunar-Bloom and adds a 120% EM hit after the shades. Under Gleam those C6 hits, and the shades, have elevation 0.15.

The burst consumes every veil and adds the talent's per-stack amount as DMG Bonus on its two hits. That row is named "DMG Bonus" and is not the Phantasm base multiplier. A Dendro DMG Bonus adds to it.

## Timing

| Timing | Class |
| --- | --- |
| Energy 60, skill CD 9s, two charges, burst CD 15s, Shadow Dance 9s, veil 9s (14s at C2), EM buff 8s, deceit window 15s, C4 linger 4.5s | Talent text |
| Normal, charged, Phantasm, skill, burst, and plunge hitmarks and cancels | Placeholders. Not a measured sheet. The N3 pair is split by 2 frames. Phantasm is 30/45 for Nefer and 48/60/72 for the shades |
| Hitbox radius 4 | Placeholder circle on the primary target |
| Slither hold, the 2.5s move, and the 18.15/s stamina drain | Not simulated. The charged attack is the exit hit only |
| A4 extra Dew from Slither | Not implemented. The text gives the 10% per 100 EM over 500, capped at 50%, and does not give the base amount |
| C2 "up to 140%" | Implemented as 5 stacks × 8% base DMG. No extra multiplier |
| Particles | Skill DMG, on hit. 2 Dendro particles plus a 2/3 chance of a third. Expectation 8/3, the wiki's 2.6667. One roll per cast |
| Interruption resistance | Not simulated |

## Validation

`docs/yuheng/configs/nefer-lunar.txt` forces 3 starting Verdant Dew. It is a mechanics check for Phantasm Performance, not evidence that a team can sustain the resource. After the swap-exit fix it still records 3 of each Phantasm Performance hit and no ordinary Charged Attack. Lauma's swap-CD mean is 0.4. That count is unchanged if the swap hook only refreshes C4, so it is not a Shadow Dance result. Xingqiu's skill and burst do not always cover the swap cooldown before Lauma acts.

`docs/yuheng/configs/nefer-natural.txt` is the same team with no starting Dew. Lauma's skill and Xingqiu's skill apply one Bloom, the counter has to mint the Dew, and only then does Nefer use Phantasm Performance. Across 20 iterations of 30 seconds there were no failed actions, 1 Bloom, and 1 of each Phantasm Performance hit. That is one Dew, not a full field time. Neither file is a claimed team DPS.

Unit tests cover identity, the skill ratio, Phantasm's five hits and the Dew spend, Shadow Dance ending on the third Phantasm, the ordinary Charged Attack after that third Phantasm, swapping away, another character's swap, re-entering, natural expiry, a Phantasm already queued before the swap, Seeds of Deceit ignoring Electro and being absorbed, the 100 EM grant at 3 stacks, C6's converted hit, the C4 shred, the C4 Dew counter, and the C3/C5 talent indexes. They also check the Veil base-multiplier bucket against a Dendro DMG Bonus, C1 inside that multiplier, C2's five stacks without a second 1.4, C6 elevation staying separate, the burst's DMG Bonus bucket, Skill particles on a hit and not on a miss, ordinary Bloom, Hyperbloom, and Burgeon, Nilou's Bountiful Cores with and without Seeds of Deceit, a Veil refresh, and the C4 Dew counter returning to 146 frames after a swap or the end of Shadow Dance.
