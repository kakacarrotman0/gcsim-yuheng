# Foundation batch 4B

Razor, Beidou, Diluc, and Chongyun charged attacks on `cursor/batch4-engine-integration`. The shared claymore helper from batch 4A was not edited. No character ID, ICD index, engine pin, or pipeline gate changed. Full pipeline regeneration was not run.

## What was already there

`ChargeAttack` on the character template returns "action charge not implemented". None of these four packages defined the action. Claymore `ActionStam` is 0 because the 40/s cost is drained inside the attack. The hitbox blocks in `config.yml` were the shared claymore placeholder, marked not implemented.

The attack type matches the helper: continuous blunt slashes, then a stronger finisher, 40 stamina per second, 5 second cap, 1U, poise 60 then 120. KQM prints `?` for every charged-attack frame column. Playback stays on Noelle's neutral schedule from batch 4A. That clock is not a measurement of these characters, and it is not frame-validated DPS.

## Sourced combat

Talent rows are the game params behind KQM's rounded table. Talent 9:

| Character | Cyclic | Final | Poise | Gauge | Stamina | Cap |
| --- | --- | --- | --- | --- | --- | --- |
| Razor | 114.8976% | 207.77% | 60 / 120 | 1U Physical | 40/s | 5s |
| Beidou | 103.332% | 187.072% | 60 / 120 | 1U Physical | 40/s | 5s |
| Diluc | 126.4% | 229.1% | 60 / 120 | 1U Physical | 40/s | 5s |
| Chongyun | 103.4078% | 186.993% | 60 / 120 | 1U Physical | 40/s | 5s |

Hits are blunt, Charged Attack (`AttackTagExtra`), and share normal-attack ICD. Hitlag is the claymore class value: 0.03s on the slash, 0.15s on the finisher, factor 0.01, defense halt. The placeholder `deployable` flag was not kept.

### Razor

Lightning Fang disables Charged Attacks. The script check returns that error while the burst status is up, and executing the action returns the same error, so the sim does not wait out the burst and then spin. The Wolf Within stays on Normal Attacks. The burst attack-speed bonus still applies to those normals (39% at talent 9). A charged attack outside the burst does not call the wolf.

### Beidou

Stormbreaker already listens for Charged Attacks, and a charged hit now triggers the bounce. Ascension 4's 15% damage bonus and 15% attack speed already target Normal and Charged Attacks. The attack speed shortens the experimental clock through `frames.AtkSpdAdjust`. It does not replace the windup.

The other half of that passive, "greatly reduced delay before unleashing Charged Attacks," has no published frame value. It stays unimplemented. Constellation 4 remains a Normal Attack instance. KQM only notes that its ICD is shared with Charged Attacks.

### Diluc

Ascension 1 cuts Charged Attack stamina by 50% and adds 3 seconds to the cap. The drain step is 6 stamina instead of 12, and the cap is 8 seconds instead of 5. Dawn's infusion already covered `AttackTagExtra` in `Snapshot`. A charged hit during Dawn is Pyro. Outside Dawn it stays Physical.

### Chongyun

Chonghua's field already infuses claymore Normal, Charged, and Plunge attacks with Cryo. A charged hit inside the field is Cryo and melts a Pyro aura at 1.5x. Constellation 1 stays on the fourth Normal Attack and still releases three Cryo blades there.

Ascension 1's text names Normal ATK SPD. The existing implementation is a general attack-speed stat of 8% inside the field. The experimental clock reads that stat, so a charged attack inside the field is shorter at ascension 1 than at ascension 0. No second attack-speed path was added.

## Timing that is not a measurement

KQM prints `?` for Razor, Beidou, Diluc, and Chongyun charged-attack frames. The playback clock is the batch 4A schedule: 56 frames of windup, a slash every 23 frames, 13 frames into the finisher, and the finisher hit 26 frames later. Chained slashes skip the windup. Attack speed scales that clock. The 5 second cap, Diluc's 8 second cap, and the 40/s rate do not scale. Stamina steps every 0.3 seconds: 12 stamina, or 6 with Diluc's ascension 1.

`charge[no_limit=1]` skips only the duration cap. From a full stamina bar the spin then ends on the finisher once stamina reaches 0, which is after the 5 second mark. Cancelling into a dash drops the trailing finisher and the next charged attack pays the windup again.

Do not quote DPS from this clock.

## Checks

`razor-charge.txt`, `beidou-charge.txt`, `diluc-charge.txt`, and `chongyun-charge.txt` run a charged attack through the simulator and record a cyclic hit and a finisher. Chongyun's file casts the skill first. They are execution checks, not benchmarks.

## Still open

Eula, Dori, Kaveh, and Sayu still have no charged attack. Beidou's shortened windup stays open until a frame count exists. After those four claymores, the next major validation batch is Bennett, Xiangling, and Sucrose.
