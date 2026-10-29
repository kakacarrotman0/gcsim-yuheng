# Foundation batch 4A

Dehya and Xinyan charged attacks on `cursor/batch4-engine-integration`. No character ID, ICD index, engine pin, or pipeline gate changed. Full pipeline regeneration was not run.

## What was already there

`ChargeAttack` on the character template returns "action charge not implemented". Claymore `ActionStam` is 0 because the 40/s cost is drained inside the attack, not paid up front. Both characters had `// TODO: charged attack`. Upstream gcsim has the same TODO. The hitbox blocks in `config.yml` were the shared claymore placeholder: circle radius 3 at offset 0.3, circle radius 3.5, hitlag 0.03s and 0.15s, marked not implemented.

## Sourced combat

Talent rows are the game params behind KQM's rounded table. Talent 9:

| Character | Cyclic | Final | Poise | Gauge | Stamina | Cap |
| --- | --- | --- | --- | --- | --- | --- |
| Dehya | 103.49% | 187.072% | 60 / 120 | 1U Physical | 40/s | 5s |
| Xinyan | 114.909% | 207.77% | 60 / 120 | 1U Physical | 40/s | 5s |

Both hits are blunt, Charged Attack (`AttackTagExtra`), and share normal-attack ICD. Hitlag is the claymore class value from the KQM frames page: 0.03s on the slash, 0.15s on the finisher, factor 0.01, defense halt. The placeholder `deployable` flag was not kept. That flag would stop the hitlag from freezing the attacker. Shipped Noelle does not use it.

`charge` is one cyclic slash. `charge[final=1]` is the finisher. Releasing the button, running out of stamina, or reaching 5s ends the sequence on the finisher. `charge[no_limit=1]` skips only the 5s cap. Dehya's burst still replaces the charged attack with Roaring Barrage. Her normal attack is unchanged.

Xinyan C1 was already the crit callback on normals. Charged hits use the same callback. The 12% attack speed applies while she is in a normal or charged state, and the playback delay uses `frames.AtkSpdAdjust`. C6 cuts charged-attack stamina by 30%. The 50% DEF bonus is flat ATK on her stat sheet, which is why KQM measured it on her dynamic skill and burst as well as the charged attack.

## Timing that is not a measurement

KQM prints `?` for both frame columns. Dehya's frame sheet has no charged-attack tab. Xinyan's Attacks tab records the charged-attack hitmark and every cancel as `?,?,?`.

The playback clock is Noelle's neutral spin: 56 frames of windup, a slash every 23 frames, 13 frames into the finisher, and the finisher hit 26 frames later. Chained slashes skip the windup. Attack speed scales that clock. The 5s cap and the 40/s rate do not scale. Stamina steps every 0.3s, which is the existing claymore drain step: 12 stamina, or 8.4 at C6.

Dehya is a tall continuous slash and Xinyan is a medium spin. The wiki calls those families functionally the same. They are not the same animation. Do not quote DPS from this clock.

## Checks

`dehya-charge.txt` and `xinyan-charge.txt` run `attack, charge, charge[final=1]` through the simulator. Both record a cyclic hit and a finisher. They are execution checks, not benchmarks.

## Next validation batch

The first characters with no local mechanics tests, chosen because they are on almost every team and the checks are energy, gauge, and snapshot rather than a new frame sheet:

1. Bennett — burst ATK share, heal, particle, C1.
2. Xiangling — Guoba particles and gauge, Pyronado snapshot.
3. Sucrose — skill charges, swirl, EM share.

Do not start that batch here. Razor, Beidou, Diluc, Chongyun, Eula, Dori, Kaveh, and Sayu still have the charged-attack TODO. Diona's A4 is still an enemy ATK debuff, and enemy hits are not built from an enemy ATK stat.
