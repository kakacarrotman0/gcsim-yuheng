# Foundation batch 3B

Yun Jin and Candace perfect counters on `cursor/batch4-engine-integration`. No character ID, ICD index, engine pin, or pipeline gate changed.

## Yun Jin, True to Oneself

Source: https://library.keqingmains.com/characters/geo/yun-jin

`skill[perfect=1]` was already the counter. It uses the press animation and Charge Level 2's damage, 4U gauge, 3 particles, poise, radius, and hitlag. The sim still does not detect an incoming hit.

The shortcut used to work before ascension 1. It is now ignored until ascension 1. A normal `hold` value still selects Tap, Level 1, or Level 2 at every ascension.

Talent 9, checked in the talent rows:

| Form | Multiplier | Gauge | Poise | Particles | Hitmark |
| --- | --- | --- | --- | --- | --- |
| Press | 253.504% DEF | 2U | 100 | 2 | 13 |
| Hold Level 1 | 443.632% DEF | 2U | 200 | 2 or 3 | 50 |
| Hold Level 2 | 633.76% DEF | 4U | 300 | 3 | 92 |
| Perfect, ascension 1+ | Level 2 numbers | 4U | 300 | 3 | 13 |

Hold Level 2's hitmark was 93. KQM lists 92, and the press and Level 1 hitmarks already matched that page, so the Level 2 hitmark is now 92. The cooldown still starts at frame 90.

Level 1 particles were a coin flip. KQM's `2~3 (3:2)` is a 60% chance of 2 and a 40% chance of 3, average 2.4. The roll is now `< 0.6`.

Press and hold cancel windows were left on the existing sheet. The current KQM cancel columns sit about 10 to 12 frames later. That offset also shows up on her charged attack, so those columns were not copied over.

C1 still shortens the 9 second cooldown to 443 frames. A perfect counter uses the press cooldown start, frame 11.

## Candace, Aegis of Crossed Arrows

Source: https://library.keqingmains.com/characters/hydro/candace

KQM's frame note is "A1: subtract 55f from Hold E hitmark." Hold hitmark 91 minus 55 is 36. Tap is 16. Both hitmarks already matched. The charged hit is 32.368% Max HP at talent 9, the tap is 20.4% Max HP, both are 1U Hydro, and particles are 2 on tap and 3 on the charged hit. Poise is 150 and 300.

`perfect` used to skip the 55 frames before ascension 1. It is now ignored until ascension 1. `hold=1` is still the full hold at every ascension.

The 55-frame skip was applied to the hit and to `Frames`, and not to the animation length, the queue point, or the cooldown start. The script waits for `CanQueueAfter` before it reads the next action, so the shorter `Frames` never let a rotation act early. Those three values now move with the hit. After the skip, swap is frame 55, the animation length is 58, and the cooldown starts at frame 34. The hold cooldown is still 9 seconds. C4 still makes a hold, including the counter, use the 6 second tap cooldown. C2's 20% Max HP still starts on the charged hit, at frame 91 or frame 36.

The KQM cancel columns are about 7 frames longer than this sheet. Those columns were not replaced.

## Batch 3A left as it was

Fischl's Oz shot, Beidou's two-stack counter, and Xingqiu's rain-sword heals were not edited. Their limits are unchanged: Oz is `aim[oz=1]` and hits the primary target only, Beidou does not detect an incoming hit, and Xingqiu does not simulate sword damage reduction or stop the orbital aura after the swords are gone.

## What defensive effects the sim can represent

Player damage uses the attacker's snapshot. An enemy has no ATK stat that a debuff can lower, and enemy attacks are not generated from such a stat. Diona's A4, 10% less ATK for opponents inside Signature Mix, would not change the damage these characters deal. It was not implemented.

The sim does represent shields, heals, and damage-reduction mods when a hit actually reaches the player. It does not represent interruption resistance. Diona's A1 stamina reduction is already in place. Beidou's charged-attack windup cut and Xingqiu's sword damage reduction remain outside this batch.
