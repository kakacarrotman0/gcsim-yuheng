# A2 claymore charged attacks

Aino, Freminet, Gaming, and Navia on `cursor/yuheng-claymore-coverage-a2`. No character ID, ICD index, engine pin, reaction formula, or pipeline gate changed. Full pipeline regeneration was not run.

## What was already there

`ChargeAttack` on the character template returns "action charge not implemented". Claymore `ActionStam` is 0 because the 40/s cost is drained inside the attack. Aino had no charge ability. Freminet and Gaming had an empty charge block. Navia's gunshot is the skill. Upstream gcsim has the same gap. The talent files had no charged-attack rows.

## Sourced combat

KQM prints the cyclic hit as 114.9%, and as 114.87% in the talent-9 frame column. The finisher is 208% there and 207.77% in the frame column. The finisher series is Xinyan's 2.0777 row. The cyclic series is not. ProudSkill groups 8531, 9131, 9231, and 12131 use 1.14866 at talent 9. Aino's normal has three hits, so her cyclic and finisher params are {3} and {4}. The other three use {4} and {5}. Stamina is 40/s. The cap is 5s. Both hits are blunt, Charged Attack (`AttackTagExtra`), and share normal-attack ICD. Poise is 60 and 120. Gauge is 1U. The base element is Physical. Hitlag is the claymore class value: 0.03s on the slash, 0.15s on the finisher, factor 0.01, defense halt. Radii are the shared 3 and 3.5 placeholders.

`charge` is one cyclic slash. `charge[final=1]` is the finisher. Releasing the button, running out of stamina, or reaching 5s ends the sequence on the finisher. `charge[no_limit=1]` skips only the 5s cap.

Aino's aimed hold skill stays unimplemented. That note remains in `config.yml`.

Freminet's Pers Timer frost and Shattering Pressure stay inside `Attack` and the skill detonation. A charged attack does not add Pers stacks, release frost, or replace itself with Shattering Pressure. Cryo and Physical pressure still come from a Normal Attack at 4 stacks.

Gaming's Charmed Cloudstrider stays the skill plunge. A charged attack does not enter that path.

Navia's Ceremonial Crystalshot stays the skill and is still the spender of Crystal Shrapnel. Ascension 1 already converts `AttackTagExtra` to Geo that cannot be overridden and adds 40% DMG for 4s. The charged attack uses that existing snapshot and attack mod.

## Timing that is not a measurement

KQM prints `?` or leaves the charged-attack frame columns blank. The playback clock is Noelle's neutral spin: 56 frames of windup, a slash every 23 frames, 13 frames into the finisher, and the finisher hit 26 frames later. Chained slashes skip the windup. Attack speed scales that clock. The 5s cap and the 40/s rate do not scale. Stamina steps every 0.3s: 12 stamina.

Aino is a short spin, Freminet and Gaming are medium spins, and Navia is a tall character who still uses the spinning iteration. The wiki calls the spin and the continuous slash functionally the same. They are not the same animation. Do not quote DPS from this clock.

Holding through Freminet's N1, N2, or N3 to keep the spin past the 5s cap is described by KQM and is not implemented. The frames for that cancel are unpublished.

## Checks

`aino-charge.txt`, `freminet-charge.txt`, `gaming-charge.txt`, and `navia-charge.txt` run `attack, charge, charge[final=1]` through the simulator. Each records a cyclic hit and a finisher. They are execution checks, not benchmarks. Navia's file does not cast the skill, so that run stays physical.
