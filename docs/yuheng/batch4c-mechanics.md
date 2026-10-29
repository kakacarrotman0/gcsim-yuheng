# Foundation batch 4C

Dori, Sayu, Eula, and Kaveh charged attacks on `cursor/batch4-engine-integration`, plus the Chongyun ascension 1 clock correction. The shared helper gained a radius argument. `ClaymoreSpinAttack` still plays the 3 / 3.5 placeholder used by Dehya, Xinyan, Razor, Beidou, Diluc, Chongyun, and Eula. No character ID, ICD index, engine pin, or pipeline gate changed. Full pipeline regeneration was not run.

## What was already there

`ChargeAttack` on the character template returns "action charge not implemented". None of these four packages defined the action. The hitbox blocks in `config.yml` were marked not implemented. Dori and Sayu use radius 2.8 then 3.3. Eula uses 3 then 3.5. Kaveh uses 2.6 then 3.3, and Painted Dome uses 3.1 then 3.6. The `deployable` flag was not kept.

The attack type matches the helper: continuous blunt slashes, then a stronger finisher, 40 stamina per second, 5 second cap, 1U, poise 60 then 120. KQM prints `?` for every charged-attack frame column. Playback stays on Noelle's neutral schedule from batch 4A. That clock is not a measurement of these characters, and it is not frame-validated DPS.

## Sourced combat

Talent rows are the game params behind KQM's rounded table. Talent 9:

| Character | Cyclic | Final | Poise | Gauge | Stamina | Cap | Radius |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Dori | 114.9091% | 207.77% | 60 / 120 | 1U Physical | 40/s | 5s | 2.8 / 3.3 |
| Sayu | 114.9091% | 207.77% | 60 / 120 | 1U Physical | 40/s | 5s | 2.8 / 3.3 |
| Eula | 126.4% | 228.547% | 60 / 120 | 1U Physical | 40/s | 5s | 3 / 3.5 |
| Kaveh | 97.644% | 176.644% | 60 / 120 | 1U Physical | 40/s | 5s | 2.6 / 3.3 |

Hits are blunt, Charged Attack (`AttackTagExtra`), and share normal-attack ICD. Hitlag is the claymore class value: 0.03s on the slash, 0.15s on the finisher, factor 0.01, defense halt.

### Dori

Constellation 6 already infuses Normal, Charged, and Plunge attacks with Electro for the existing window. A charged hit inside that window is Electro. The heal on that constellation is attached to Normal Attacks only. A charged hit does not heal. The code's infusion length stays 228 frames.

### Sayu

Fuuin Dash stays the skill. The attacks table prints the spin percent as "x 9". That product is not a fixed volley. One slash lands per experimental interval, then the finisher.

### Eula

Lightfall Sword charges when Eula's own Normal Attack, Elemental Skill, or Elemental Burst deals damage, once every 0.1s. Charged Attacks are outside that list. Constellation 6's extra stack lives in the same callback, so a charged hit does not roll it either. A Normal Attack during the sword still adds one stack. Grimheart and the shattered hold-skill stay off this action.

### Kaveh

Painted Dome already turns Normal, Charged, and Plunge hits into Dendro, and normals already switch to the enlarged hitboxes. A charged hit during the dome is Dendro and uses the configured 3.1 / 3.6 radii. Ascension 4 already counts charged hits toward the 25 Elemental Mastery stacks. Constellation 6 already listens for charged hits and fires Pairidaeza's Dreams from one. Ordinary hits still do not rupture Dendro Cores. That gap is the same one normals have.

Constellation 2's text is Normal Attack SPD, implemented as a general 15% attack-speed stat during the dome. Normals keep it. The charged-attack clock subtracts that 15% while `kaveh-c2` is active.

### Chongyun

Batch 4B left ascension 1's 8% on the charged-attack clock because the existing mod is a general attack-speed stat. The talent names Normal Attacks. The clock now subtracts 0.08 while `chongyun-field` is active. The stat itself is unchanged, so a normal inside the field is still faster at ascension 1 than at ascension 0. A charged attack inside the field takes the same experimental delay at both ascensions. Cryo infusion, melt, and constellation 1 are unchanged.

## Timing that is not a measurement

KQM prints `?` for Dori, Sayu, Eula, and Kaveh charged-attack frames. The playback clock is the batch 4A schedule: 56 frames of windup, a slash every 23 frames, 13 frames into the finisher, and the finisher hit 26 frames later. Chained slashes skip the windup. Attack speed scales that clock, except the two normal-only bonuses above. The 5 second cap and the 40/s rate do not scale. Stamina steps every 0.3 seconds, which is 12 stamina.

Do not quote DPS from this clock.

## Checks

`dori-charge.txt`, `sayu-charge.txt`, `eula-charge.txt`, and `kaveh-charge.txt` run a charged attack through the simulator and record a cyclic hit and a finisher. They are execution checks, not benchmarks. Kaveh's file stays outside Painted Dome.

## Still open

Beidou's shortened windup stays open until a frame count exists. The next major validation batch is Bennett, Xiangling, and Sucrose. This batch does not start it.
