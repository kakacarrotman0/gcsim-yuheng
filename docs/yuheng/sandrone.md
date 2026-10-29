# Sandrone

Yuheng-maintained gcsim character. Upstream `genshinsim/gcsim` and the production engine pin do not contain her. `excel-hk4e` on this branch predates her, so `go run ./pipeline` was not run and she has no character `config.yml`. Her machine-gun ICD row lives in `pipeline/config.yml` with the other hand-maintained groups. `pipeline run` refuses to write when that regeneration would drop her or renumber production IDs.

## Identity

| Field | Value |
| --- | --- |
| English name | Sandrone |
| Internal name | MarionetteNew |
| Key | `sandrone` |
| Avatar ID | 10000133 |
| Skill depot | Not in the Yatta release dump or the gachabase release page. Catalog `SubId` 13301 is the conventional `(id % 1000) * 100 + 1` guess. Combat does not read `SubId`; only a future pipeline depot lookup would. It is not a runtime integration risk |
| Element | Cryo |
| Weapon | Claymore |
| Rarity | 5 |
| Region | Snezhnaya (`ASSOC_TYPE_SNEZHNAYA_STAR`) |
| Body | Girl |
| Ascension stat | CRIT Rate, 19.2% at ascension 6 |
| Constellation | Narcissorolegium |
| Energy / cooldowns | 60 energy, Skill 4s, Burst 15s |
| Icon | `UI_AvatarIcon_MarionetteNew` |
| Namecard | `UI_NameCardIcon_MarionetteNew` (not used by the sim) |
| Yatta release | 2026-06-29 4:00 PM CT (unix 1782766800) |

## Source identifiers

Talent rows from the Yatta release dump of avatar 10000133. Percent specials in `ConfigAbility_Avatar_MarionetteNew.json` are 0 because the proud-skill rows are not in `excel-hk4e`. Literal graph intervals and the talent percentages below are what the sim uses.

| Row | ID |
| --- | --- |
| Normal Attack, Formule Phenomenale: Self-Evident Proposition | 11331 |
| Skill, Formule Phenomenale: Differential Analysis | 11332 |
| Burst, Formule Phenomenale: Q.E.D. | 11335 |
| A1, Eternal Speculation Engine | 13321 |
| A4, A Lady's Code of Conduct | 13322 |
| Stellar Jubilee, Light of Rationalisme | 13323 |
| Utility, A Caucus Prelude and a Long Tale | 13325 (not combat) |
| C1, Morrow After the Golden Dusk | 1331 |
| C2, An Heiress Gazed Into the Looking-Glass | 1332 |
| C3, Refuse the Wake of Dusk, the Moonlit Yoke | 1333 |
| C4, In Knowledge Lies the World's True Ground | 1334 |
| C5, Of All Beside, She Takes No Part | 1335 |
| C6, Narcissus Wakes, Her Eyes Upon the Dawn | 1336 |

Ability graph: `ConfigAbility_Avatar_MarionetteNew.json` on `DimbreathBot/AnimeGameData` commit `792978e5503ecfba73dcb3562ed44a0d35a2abe2` (2026-10-01). That file is the Version 7.1-era dump also used for Vodyanitsa. Bullets referenced from it:

| Bullet / gadget | Role |
| --- | --- |
| 41133002 | Sweeping-fire machine gun. Spawned once on attach. Its internal hit loop is not in the avatar file |
| 41133004 | Power Overdrive shot, think 0.4s |
| 41133008 | Named by the C6 cluster path. The avatar file's T6 bullet ability is what supplies the four extra hits; the bullet asset itself is not in this dump |

Gameplay cross-check: KQM Sandrone Quick Guide updated for Version 7.1, `https://keqingmains.com/q/sandrone-quickguide/`. KQM documents the C0 combo `3[C E] (Q)` and the C1+ combo `2[C E] (Q)`. It does not publish frame data. Those combos are not frame-validated here.

## Mechanics matrix

| Mechanic | Status |
| --- | --- |
| 3-hit Physical normals, Physical plunge, Cryo charged / skill / burst ratios from the talent table | Implemented. NA and plunge hitmarks are placeholders |
| Fagio Decoding, condensed beams, Overdrive, decay, skill repair | Implemented from the ability intervals |
| Two Prism Shots. Second shot is direct Stellar-Conduct or direct Stellar Swirl when that radiance is active | Implemented |
| A1 400% on the stellar second shot only, Refined Tactics, burst ray multiplier | Implemented. See the A1 note |
| Shared 2.5s particle ICD, 1 Cryo particle | Implemented for the elemball attack tags |
| Burst: 3 bombardments and 1 ray. Ray class is snapshotted at 0.28s | Implemented. Burst cancel frames are placeholders |
| Stellar Jubilee enables both reactions and adds the ATK Base DMG bonus | Implemented on the existing reaction engine |
| A4, 8 EM per 100 ATK, cap 160 | Implemented |
| C1–C6 combat effects | Implemented. C6 cluster spacing is approximate |
| Hover steering, sprint stamina, 6s hover as movement | Unsupported. The skill state can last up to 360 frames if it is not cancelled |
| Claymore charged-attack stamina | Left at the template value of 0. The ability special is also 0 |
| Continuous sweeping-fire ticks inside bullet 41133002 | Unsupported. The avatar file spawns the bullet once. The bullet asset was not in the public dumps checked for commit `792978e`. One attach hit stands in for it |
| Hitlag on her own hits | Omitted, so the Core task clock stays aligned with the ability thinks |
| Commission snack utility | Unsupported |

## Fagio

Decoding Power is 0–100.

Charged Attack from idle or Decoding enters Decoding. Charged Attack during Overdrive returns an error. The charge stays locked until continuous gain would reach 100. From 0 that is frame 192 at C0 and frame 372 at C1. That horizon is the 0.2s / 1.0s ability clock. It is not a measured cancel window. The sim ticks that frame before it accepts the next action, so the capping tick enters Overdrive, including the Overdrive attach shot, and then Skill can repair.

While Decoding, every 12 frames adds `3.2 * ratio`. Every 60 frames fires a condensed beam and then adds `17 * ratio`. C1 sets `ratio` to 0.5. From empty, C0 ends frame 180 at 99 with three beams and frame 192 at 100. C1 ends frame 360 at 99 with six beams and frame 372 at 100. `UNIQUE_MarionetteForceSet100` is the 0.25s backup that forces 100 when the beam count is exactly 3, or exactly 6 at C1. Natural gain reaches 100 first when the bar starts empty, so the backup does not change that path. A bar left at 30 by an instant 70-point repair reaches the cap on the second beam.

Sweeping Fire is one Cryo hit on the machine-gun attach (frame 12), talent param 4, `ICDGroupSandroneMachineGun`. That ICD group is a provisional copy of the default 2.5s / 1-0-0 pattern. On the integrated branch it follows `ICDGroupVesnaSkill`, which follows `ICDGroupVesnaSkillStart`, which follows `ICDGroupLohenSkill`. Do not regenerate `icd_groups.dm.go`. Overdrive uses the same curve (param 7) and the same ICD group: one hit on attach, then every 24 frames, and it does not drop particles.

Condensed beams use param 5 when ordinary, param 6 for direct Stellar-Conduct, and param 11 (the same curve as param 5) for direct Stellar Swirl. Ordinary condensed beams are 1U Cryo with `ICDGroupDefault`. Direct stellar hits are 0 gauge, no ICD, and ignore DEF. C2's CRIT DMG snapshot is taken on the beam, and C6 cluster subhits share it.

Decay runs every 30 frames whenever she is not Decoding and not in her on-field Skill state: −2.5 on-field, −7.5 off-field. The graph's driver also skips the animator states ExtraAttack, ExtraAttack1, and ExtraAttack1Weak. The talent and KQM say power decreases whenever she is not in Decoding, including Overdrive, at 300% off-field. The overload modifier removes itself when power changes to 50 or below, which only happens if decay runs during Overdrive. The sim follows that pair of facts. A value that is written onto 50 without going through the change does not exit until the next decrease. Landing on 50 exits. The next 0.4s Overdrive shot after that tick does not fire.

The same decay think records Refined Tactics. Both frost-reduce predicates run, so a loss that exactly equals the remaining power is recorded twice and then clamped at 100. Each such think refreshes the 60s timer. The talent sentence describes the 10-point rule without saying that passive decay counts. The graph's decay think does add `_ABILITY_Marionette_FrostReduce`.

Skill immediately removes 70, then 10 every 6 frames while she stays in the Skill state on-field. `ReduceFrost` is attached to the hover animator states, so a cancel stops later repair ticks. The 0.1s ticks do not refresh the 60s timer. The skill cast does, because it triggers `FrostValue_Driver` with opt type 1, including when the bar is already empty. Both the "at least the chunk" and "at most the chunk" frost-reduce actions run, matching the graph, so an exact boundary counts twice and the recorded loss is clamped at 100. Refined Tactics is `floor(loss / 10)`, maximum 10. A dash at the placeholder frame 12 leaves power at 10 from a full bar. An attack cancel at the placeholder frame 36 drains a full bar to 0 and records 10 tactics. Those cancel frames are labels for the ability repair clock, not a measured EN0 or ED window.

The Prism cannon is gadget `41133006`, created with `lifeByOwnerV2`. `FunnelEff` lasts 1.45s on that gadget. `ElementalArt_Cancel` does not kill it, and the bullet actions are `doOffStage`. The first shot is `FunnelShooterReal` at 0.25s (frame 15), which then removes itself. `FunnelEff`'s 0.3s think (frame 18) snapshots `FunnelStar` and attaches the enhanced shooter. That shooter fires at 0.55s (frame 33). A dash at frame 12 does not suppress either shot. Radiance gained after frame 18 does not change the second shot.

## Stellar reactions and Radiance

`stellarInit` sets `StellarConductEnableKey` and `StellarSwirlEnableKey`. Sandrone does not reimplement the reaction formulas. Polestar Field wins over the Radiance: Stellar Swirl status, using the same key string Odette uses. A new Stellar Swirl applies `RadianceSwirlDuration` (8s, or 12s when the Vodyanitsa song flag is active and the status is not already up).

Her Jubilee bonus is `min(TotalATK / 100 * 0.007, 0.14)` Base DMG, added on her direct Stellar-Conduct hits, her direct Stellar Swirl hits, and reaction Stellar Swirl. Odette's bonus is a separate hook and stacks with hers. Conduct as a reaction still does not deal its own damage number. Ordinary Superconduct and Cryo Swirl still occur when those enable flags are absent. With Sandrone in the party they become the stellar reactions.

Direct stellar hits do not carry elemental gauge. Cryo DMG% is not part of the stellar formula. DEF is ignored on those hits. EM, reaction bonus, resistance, elevation, and CRIT still apply through the existing calculator.

## A1

The skill records the A1 mark when ascension is at least 1 and Decoding Power is at least 50 before the 70-point drain. The ability comparison is `MoreThanAndEqual` 50. The talent sentence says "greater than 50". The sim follows the ability.

The 400% multiplier is applied only when the second Prism Shot is a direct stellar hit. `Funnel_Bullet_Enhanced` multiplies `P1_UpRatio` on the stellar factors. The ordinary Elemental Art shot uses `DamageRatio` alone, and the talent sentence is under the Radiance prefix. `P1AttackUpRatio` is 0 in the dump, so 400% is the talent wording (`1 + 3` in the graph's `1 + P1AttackUpRatio` shape, implemented as a ×4 on the chosen stellar ratio).

Refined Tactics are cleared on every Burst once A1 is unlocked, including a Burst outside Radiance. `ElementalBurstEnhance` is triggered with only the A1 talent predicate, and its modifier zeros `FrostReduce` and removes the 60s timer on cast. The talent sentence says the clear happens while in Radiance: Stellar Glimmer. The clear is at cast, not at the 0.28s ray snapshot. `BurstP1RatioUp` is `1 + 0.10 * stacks` and is multiplied into both the Stellar-Conduct ray and the Stellar Swirl ray when that snapshot is stellar. A Polestar Field that appears after cast and before frame 17 still boosts the ray. One that disappears before frame 17 leaves an ordinary ray, and the stacks are already gone. The talent sentence names only Stellar-Conduct for the bonus. The burst attacker multiplies the same global into both moon factors. The ordinary ray does not receive it. The `MarionetteBurstSC` damage-extra mixin is not paired with a burst attack tag in the attacker, so it is not a second hit.

The ray's element is the `MarionetteBurstStarmode` value written by `SetStarModeDelay` at 0.28s (frame 17), not radiance at the 0.85s ray.

## A4

At ascension 4 or higher, a stat modifier adds 8 EM per 100 total ATK, capped at 160. `TotalAtk` does not read EM, so the modifier does not recurse.

## Constellations

C1 halves Decoding gain. While Decoding, a 2s party status is refreshed every 1.5s and grants 30% reaction bonus on direct Stellar-Conduct, direct Stellar Swirl, and reaction Stellar Swirl. The 40m range is the whole party. The bonus is not permanent.

C2, only while radiance is active, adds CRIT DMG on condensed beams and their cluster subhits. `MarionetteCylinderAttackFather` triggers `Avatar_MarionetteNew_T2` and then attaches the beam. T2's `onAbilityStart` adds one layer, max 3, before that attach. With the talent's 40% plus 20% per layer, the beams snapshot +60%, +80%, and +100%. Further beams stay at +100%. Leaving radiance or ending Decoding clears the layers. The graph expresses this as `ModifyDamageMixin` specials that are 0 in the dump. The percentages and the CRIT DMG stat are the talent text.

C3 adds 3 to the Normal Attack talent. C5 adds 3 to the Burst talent. The Skill talent is not constellation-upgraded.

C4, on her direct Stellar-Conduct or direct Stellar Swirl hit, queues one cannon after 30 frames for 125% ATK or 187.5% ATK. The ICD status is set for 4s before the cannon is queued. The cannon uses the same direct-stellar tags and a 0.5s delay, and its name is ignored by the hook, so it does not fire again. Reaction Stellar Swirl is not in the ability's tag list. The swirl predicate also excludes `StarSwirl_Ice_Burst_01` and `StarSwirl_Ice_Burst_02`. Ratios are the talent text. The specials are 0 in the dump.

C6, from the third condensed beam onward, keeps the original beam and adds four hits at +10, +20, +30, and +40 frames. That spacing is not in the avatar file. Ordinary cluster hits are 100% ATK Cryo, 1U, and do not roll the particle ICD (`MarionetteExtrdaAttackLaserNormalT6` is absent from the elemball tag list). In radiance the four hits replace that ordinary damage with 80% direct Stellar-Conduct or 120% direct Stellar Swirl, and those hits use the `MarionetteLaser` tag, which does roll the shared particle ICD. Elevation +20% is applied once on her direct stellar hits and on reaction Stellar Swirl where she is the actor. It is not also stamped as a second bonus. The graph's fail path can replace the original beam and skip the count increment. That reading would freeze the beam count at 2 and never reach the force-100 threshold. The sim follows the talent: the original beam still fires, and the four hits are additional.

## Timing

| Timing | Class |
| --- | --- |
| Gain 0.2s, condensed beam 1.0s, Overdrive 0.4s, decay 0.5s, force-100 0.25s, skill repair 0.1s and the instant 70, prism shots at 0.25s and 0.55s with the radiance snapshot at 0.3s, particle ICD 2.5s, bombardment 0 / 0.2 / 0.4s, starmode snapshot 0.28s, ray 0.85s, C1 2s refreshed every 1.5s, C4 ICD 4s and delay 0.5s, tactics 60s | Ability-graph intervals |
| Charge lock until power would hit 100 (192 frames from 0 at C0, 372 at C1) | Derived from those intervals. Not a measured animation |
| Skill dash/jump at frame 12, attack/charge/burst/swap at frame 36, hover length 360 | Placeholder windows on the repair clock. Not an EN0 or ED sheet |
| Normal hitmarks 16 / 14 / 28, plunge collision 36, low 42, high 44, burst cancel slice | Placeholder. Not a frame sheet |
| C6 cluster spacing of 10 frames | Approximation |
| Burst energy drain | On cast. The drain frame is not measured |

No Sandrone DPS figure in this document is frame-validated.

## Pipeline

`ICDGroupSandroneMachineGun` follows `ICDGroupVesnaSkill` and `ICDGroupVesnaSkillStart`, which follow `ICDGroupLohenSkill`. Every production ICD index through Lohen is unchanged. `keys.Nefer`, `keys.Vesna`, and `keys.Sandrone` are appended after `Zhongli`, so existing character numbers stay put. Talent curves live in `internal/characters/sandrone/zz_sandrone.dm.go` and are hand-maintained. `go run ./pipeline run` refuses to overwrite those tables while the three kits are outside excel. Catalog `SubId` 13301 remains unverified. The simulator's combat path does not read it.
