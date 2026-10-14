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

`docs/yuheng/configs/lohen-masterstroke.txt` is a constructed baseline: Lohen / Xiangling / Bennett / Kaeya, 20 iterations, 20 seconds, `hexerei` left off. Re-run after the shared 5s Cryo ICD, it completed 20 iterations with no failed actions. It is not a damage benchmark.

That run's Lohen damage log, from the gcsim result JSON:

- Masterstroke normals from the four-attack string (N3 is three hits). No Physical normal source
- Lohen's own damage in that config is Cryo only
- Etched Into Bone and Soul 1–4: one each
- Manifest Judgment 1–6: one each
- Bennett and Kaeya each recorded a Melt. Lohen's converted hits land before Pyro is applied in this rotation, so Lohen records no Melt here. The unit test covers his own Melt

The printed summary was average 265553.09 damage over 20.00 seconds (13278 dps, min 11816.84, max 14058.35, std 592.64). That number is the output of this one config. It is not a claimed team DPS.

Unit tests in `lohen_test.go` check identity, Joy on a normal hit, the 0.1s Joy gates (normal, charged A, and charged B are separate), Will from a teammate hit including the A1 extra, the per-character 0.1s Will gates, Will at ascension 0 without that extra, High Spirits at ascension 0, Will consumption on Etched, the `1 + Will×0.004` ratio at 50 Will, C1's 5× gain and 300 cap, C2 Evilsbane damage and the 200 EM buff and its 8s expiry, C4's immediate 15 energy and the post-burst 15, C5's burst index, C6 keeping Will and refilling Joy, C3's skill index, the A4 ATK buff on Melt, Stellar Swirl, and Stellar Conduct plus the 8s expiry, six burst hits, Physical normals and charged attacks outside Masterstroke (including a Pyro infusion), Cryo normals, charged attacks, and plunges inside Masterstroke with infusion blocked, Cryo application, a Melt, the one-particle / 2s particle ICD, Masterstroke ending on swap, the 5s/2-hit application sequence and its reset, a normal and a charged attack sharing that sequence, the placeholder N1C count of 13, a 58-frame baseline spacing with hitlag on, the optional `n1c_cadence=1` profile, Hexerei's 40% normal bonus, the excluded fifth-normal second hit, the 6s expiry, and a one-Hexerei party that does not gain the buff. `TestCadenceOnlyInMasterstroke` checks that the optional profile changes N1→CA and CA→N1 only while Masterstroke is active. `TestA4StellarFromReactions` runs those two Stellar reactions through gauge application.

## Timing

| Timing | Class |
| --- | --- |
| Energy 60, skill CD 18s, burst CD 15s, Masterstroke 13s at every talent level, Joy cap 100, Will cap 100 (300 at C1) | Ability graph / talent text |
| Joy +17 per normal hit and per charged hit. Charged stamina 25, or 10 during Masterstroke | Skill talent parameters, constant across levels |
| Will on a teammate hit: +20 if the hit is at least 10× Lohen's base ATK, otherwise +1. This is the skill, including at ascension 0. A1 (promote level 1) adds +60 when the hit is at least 30× base ATK. C1 multiplies those gains by 5 and raises the cap to 300. Lohen's own hits do not add Will. Success, fail, and the A1 extra each have their own 0.1s unique modifier on that teammate | Ability graph. The 30× and +60 figures are the talent text |
| Joy's normal, charged-A, and charged-B modifiers are each unique for 0.1s. A second hit of the same kind inside that window adds nothing. The three kinds do not block each other | `AddShowLevel` modifiers on `GrandHandler` |
| Etched damage is `Raid × (1 + Will × 0.004)`. Burst damage is `Burst × (1 + Will × 0.004)`. Both snapshot Will, then Will is cleared unless C6's ICD flag is clear and Masterstroke is active | Ability graph |
| Etched hit order 0.086, 0.2095, 0.4074, 0.5316 of the raid state, played as frames 5, 13, 24, 32 | Normalized times are from the graph. The 60-frame clip length is not in the JSON. Absolute frames are unverified |
| Burst hits at frames 12, 36, 42, 48, 54, 78 | Gadget think of 0.1s, assuming the first increment is immediate. Six hits match the talent's ×6. The blank step is not a hit |
| Burst extends Masterstroke by 1.65s. The next Etched after a C6 mark extends it by 1.25s, and remaining duration is capped at 20s | Ability graph |
| Normal, charged, skill-cancel, burst-cancel, and plunge frames | Unverified polearm spacing. The default N1C loop stays 58. An optional profile sets the combined loop to the recorded 46-frame input cadence. See Frame review |
| Particles | One Cryo particle (`baseEnergy` 3, config 2022) when an Ice hit with a normal, charged, plunge, or elemental-art tag lands. Unique modifier 2.0s |

## Frame review

No measured Lohen frame sheet is public. The KQM library page for Lohen still says its findings have not been added. `gensri.wiki` has no Lohen frame page. The ability JSON has Joy modifiers of 0.1s and raid-bullet durations, and it does not give attack-state lengths that convert into hitmarks. Combo counts were not turned into hitmarks.

BowTae's 60 FPS recording (`Day 1 Lohen Combos 60fps`) has a gamepad overlay. On the final `E 17N1C N1 E` segment the attack-trigger onsets are 46–47 frames apart after the first transition (median 46, mean 46.47, range 46–48). The same cadence shows up on the earlier 14, 15, and 13 N1C runs. The split clip reproduces that interval sequence. The on-screen ping in the clip is 24 ms. The overlay can shift the represented input by a few frames. The recorded frames include hitlag. The onset is the attack-button press, not the damage hitmark.

KQM's quickguide (Version 7.1) lists `E 15–17[N1C] (N1 E)`. Average play completes 15–16 N1C loops during Masterstroke. 17 needs near-frame-perfect input and a favorable enemy. The count also depends on ping and frame rate. Charged-attack input cannot be buffered. That row is a combo count, not a frame sheet. The footage is one optimized 17-loop example of that count.

The default placeholders:

| Window | Frames |
| --- | --- |
| N1 hitmark / N1→CA cancel | 12 / 22 |
| CA hitmarks | 18, 28 |
| CA animation; cancel into attack, skill, or burst; dash; swap | 46; 36; 28; 34 |
| N1C loop (22 + 36) | 58 |
| Skill animation; attack, charge, or burst; dash; swap | 32; 24; 22; 26 |
| Burst animation; attack or skill; dash; swap | 90; 80; 78; 82 |
| Etched delays | 5, 13, 24, 32 |
| Plunge | Standard polearm template. Low and high impact hitmarks 45 and 46 |

On the 17 N1C segment the overlay shows one attack-trigger rising edge per cycle. The repeated holds last 19–24 frames (the opening hold is 28). The extra N1 after the 17th string is the same kind of rising edge, 46 frames later, held for 6 frames. So each onset is the start of the attack, and the longer hold is the N1C gesture. The presses are not a mix of N1 inputs and charged-attack inputs. The interval between those starts is the combined cycle.

`n1c_cadence=1` on the character (`+params=[n1c_cadence=1]`) is an optional profile. It sets that combined cycle to 46 only when the normal or the charged attack starts during Masterstroke. The footage is Masterstroke gameplay. Outside the mode the placeholder cancels stay in place. Omitting the param keeps the 58-frame baseline in both states. The 46 still includes hitlag and any delay shared by both presses. The profile does not move a hitmark. The 12 frames between 58 and 46 are taken from the recovery after the existing last hitmarks (N1 at 12, second charged hit at 28), and the 6 frames that remain are split evenly, leaving a 3-frame tail on each. That is N1→CA 15 and CA→N1 31. Those two numbers are a partition of 46, not a counted cancel pair. Every hitmark and cancel window in both profiles is approximate.

A second recording, 1080p at 60 FPS with an OBS mouse overlay (`2026-10-14 09-38-51`, enemy present), agrees on the combined interval. No-enemy recordings from the same session show the same input cadence, about 46–48 frames. The CSV for the enemy clip is an early-click trial. Each long press is the N1C gesture. The gaps are from that press to a later tap. They are not the start of the charged-attack animation, so they are not a CA→N1 cancel and they are not an N1→CA cancel.

| Hold start (frame) | Hold length | Later taps, frames after the hold |
| --- | --- | --- |
| 48 | 28 | 44, 69 |
| 358 | 30 | 45, 68 |
| 657 | 31 | 52 |
| 1118 | 32 | 52 |
| 1399 | 31 | 51 |
| 1668 | 29 | 56 |

Taps at 44 and 45 may have been early. That is not established. An earlier clip the same morning (`09-26-18`) had taps at 49–51 that appeared to connect. Taps at 51–56 in this clip have a clearer following attack. The first two trials also have a second tap, so the accepted press cannot be picked out of the pair. None of these gaps is a proven earliest input. The optional loop stays at 46. It does not move to 44, 45, or 51.

Still unmeasured after both recordings: the N1→CA cancel, the CA→N1 cancel, skill, burst, Etched, plunge, every hitmark, and hitlag. The 15/31 pair stays an even partition around the old hitmarks.

`TestN1CProfilesAgainstVideoCadence` runs both profiles with hitlag and defense halt enabled, queueing the next attack at the first legal frame. Lohen still sets no halt, so the clock does not grow.

| Profile | Attack-to-attack gap | N1C strings started in 13s | Charged hits that land inside Masterstroke |
| --- | --- | --- | --- |
| Baseline, no param | 58 | 13 | 13 strings, both hits |
| `n1c_cadence=1` | 46 | 17 | 16 strings, both hits. The 17th normal hit lands inside the mode. Its two charged hits land after the unextended 13s |

The baseline 13 is below KQM's 15–17 band. The cadence profile's 17 started strings match the frame-perfect end of that band and the 17 in the video. Sixteen of them finish both charged hits before frame 780. Average play at 15–16 is this cadence plus a frame or two of lateness per loop, which is what the 47- and 48-frame gaps in the video are. The skill cancel into the first normal stays the unverified 24. That 24 is why the 17th charged hits miss the 13s. The footage did not isolate the skill press, so the skill window was not changed.

Hitmarks on the first cadence cycle are still 12, 18, and 28. The damage config is unchanged and stays on the baseline. The cadence profile is not a damage benchmark.

## Elemental application

KQM and Prydwen describe 6 Cryo applications every 5 seconds from empowered normals and charged attacks during Masterstroke. With N1C, KQM says that cap is usually reached in 3–4 seconds, leaving a 1–2 second gap. The Bilibili wiki element-application table (`元素附着论/角色数据`) names the group `洛恩战技`: a 5 second reset, sequence `1,0,1,0,1,0,1,0,1,0,1,0`, and at most 6 applications before the reset. Normal attacks, charged attacks, and both of those during Masterstroke (`奇谋状态`) share tag `普通攻击` and that group. Physical hits and Masterstroke hits advance one counter.

That is `ICDGroupLohenSkill`: reset timer 300, element sequence six `1,0` pairs, damage sequence all `1`. Both the normal and the charged attack use `ICDTagNormalAttack` with this group. The charged attack's damage tag stays extra attack. gcsim evaluates ICD for any hit with durability, including Physical, so the shared counter matches the table. The engine reset fires at timer − 1, the same convention as every other group.

The same table gives plunge landing no tag and no ICD. Plunge stays `ICDTagNone`. GameVika writes "5s/2 Hits" on the first normal-attack row and a dash on charged attacks and plunge; the plunge dash disagrees with the wiki's explicit `无冷却`, so the wiki was followed. Etched (`镂骨彻心`, group `默认`) and Manifest Judgment (elemental burst, group `默认`) stay on the standard 2.5s/3-hit group. GameVika lists those the same way. Icy Veins' "5 applications every 6 seconds" was not used. The ability JSON has no attenuation row for this group; the published table is the source.

N1C is three hits. The sequence applies on hits 1 and 3 of the first string, then one, then two, then one: six applications by the eleventh hit, the end of the fourth N1C. Later hits in the window stay dry until the 5 second reset. `TestLohenSkillICD` and `TestNormalAndChargedShareSkillICD` check that sequence on engine results.

## Mechanics notes

- Entering Masterstroke clears Joy, Will, the etched counter, and the C6 mark. The counter reset is `Lohen_CountExtraE_Report` when the mode flag rises.
- Leaving the field ends Masterstroke before the 13s timer. `GrandHandler` `onAvatarOut` removes the modifier, and `onRemoved` zeroes Joy, Will, and the C6 mark. Etched is then unavailable. Swapping back does not resume the mode. The etched counter itself is not cleared until the next Masterstroke, matching the graph.
- A normal or charged hit chooses Physical or Cryo when it lands. A plunge chooses when the action starts; its cancel window is after the impact, so a later swap cannot change that hit.
- Etched is `skill[etched=1]`. It requires Masterstroke, Joy at 100, and remaining uses (3, or 5 at C6). It does not restart the 18s skill CD.
- A plain skill while Masterstroke is already active restarts the mode. That matches `RemoveUniqueModifier` plus `AttachModifier` on the skill start.
- Etched and Burst consume Will after the snapshot. C6 skips that clear only while its 7s flag is down. The talent sentence can be read as "never consume." The graph ties the skip to the same flag that gates the Joy refill. This implementation follows the graph.
- C4, on entering Masterstroke, refunds 15 energy when energy is not full. When it is full, the next Burst within 15s refunds 15. A Burst during Masterstroke sets Will to its cap before the snapshot.
- C2 arms Evilsbane for 4s after the fourth Etched hit or the last burst hit, ICD 4s. The next normal or charged hit during Masterstroke deals 500% ATK Cryo. The pattern is a 5.5 by 5.5 box starting 1.3 in front of Lohen. The pattern's vertical extent of 2.5 is not a ground radius. Gauge is 1U. Other members within 40m gain 200 EM for 8s. The caster is excluded (`isTarget: false`).
- C4's burst revive is queued one frame after the energy drain. Character tasks run before that drain, so a same-frame revive would be added to a full bar and lost. The net result is 15 energy once the bar is empty.
- C6 adds 175% CRIT DMG to hits whose names start with "Etched" or "Manifest" while Masterstroke is active. Evilsbane does not get that bonus.
- A4 listens for Melt, Freeze, Superconduct, Cryo Swirl, Stellar Swirl, Cryo Crystallize, and Stellar Conduct from another member during Masterstroke, then gives that member and Lohen 15% ATK for 8s. The graph names those `Melt`, `Freeze`, `Superconductor`, `SwirlIce`, `StarSwirlIce`, `CrystallizeIce`, and `StarSuperconductor`. `OnStellarSwirl` is emitted only for Cryo or Frozen. `OnStellarConduct` is the Cryo and Electro Stellar reaction.
- High Spirits adds 1 skill level for 9s, plus 6s when another member's normal, skill, or burst level is at least Lohen's skill level, ICD 18s. Proud skill 12923 has no promote level on skill depot 12901. Proud skills 12921 and 12922 are promote levels 1 and 4. An omitted level is 0, so High Spirits is available at ascension 0.
- Hexerei is off unless the character param `hexerei=1` is set. With at least two Hexerei characters and Will at least half of its cap when Etched's last hit lands or Burst is cast, normals and charged attacks gain 40% DMG for 6s. The mixin lists `NormalAttack_01` through `NormalAttack_05` and `ExtraAttack`. The second hit of the fifth normal is `NormalAttack_06` and is excluded. Its ability name ends in ` B`.
- `n1c_cadence=1` selects the optional 46-frame N1C loop only while Masterstroke is active at the start of that normal or charged attack. The default, and any attack that starts outside the mode, keeps the 58-frame placeholder. Hitmarks do not change with the param. Both sets of hitmarks and cancels are approximate.
- The quest sneak bullet is not implemented.

## Completion

| Item | Status | Source and check |
| --- | --- | --- |
| Masterstroke Cryo conversion, Joy cap 100, Etched at 100 Joy, Will ratio 0.004, particle 1 Cryo / 2s, swap ends the mode | IMPLEMENTED AND TESTED | Ability graph. `TestMasterstrokeCryoConversion`, `TestMasterstrokeEndsOnSwap`, `TestEtchedConsumesWill` |
| Joy 0.1s unique gates, separate for normal, charged A, and charged B | IMPLEMENTED AND TESTED | `AddShowLevel` modifiers, duration 0.1, stacking Unique. `TestJoyAndWillGates` |
| Will 0.1s gates, separate for success, fail, and the A1 extra, and separate per teammate | IMPLEMENTED AND TESTED | `AddAnswerLevel_*_Pre` modifiers on the teammate monitor. `TestJoyAndWillGates` |
| 40m check before a teammate receives the Will monitor, the A4 monitor, or the C2 EM buff | NOT IMPLEMENTED | Graph predicate `LesserOrEqual` 40 on `AllPlayerAvatars`. Blocked: gcsim has one player position, so two party members are never 40m apart |
| A4 Melt, Freeze, Superconduct, Cryo Swirl, Cryo Crystallize | IMPLEMENTED AND TESTED | `TestA4CryoReaction`. Melt also appears in the team configs below |
| A4 Stellar Swirl and Stellar Conduct | IMPLEMENTED AND TESTED | `TestA4StellarFromReactions` applies Cryo, then Anemo, then Electro, with Cryo Traveler in the party. The engine emits `OnStellarSwirl` and `OnStellarConduct` and not the plain Cryo Swirl or Superconduct events. Lohen and the triggering character gain 15% ATK. `TestStellarA4AndHexerei` still checks the 8s expiry. Without a character that sets `stellar-swirl-enabled` and `stellar-conduct-enabled`, the same gauges are ordinary Swirl and Superconduct. Lohen does not set those flags |
| C1 5× Will and cap 300, C3 skill +3, C5 burst +3 | IMPLEMENTED AND TESTED | Talent text for the magnitudes. `TestC1C2C4C5`, `TestC6KeepsWill` |
| C2 Evilsbane 500% / 200 EM / 8s / 4s, box 5.5 starting 1.3 forward | IMPLEMENTED AND TESTED | Graph pattern and talent text. `TestC1C2C4C5`. `lohen-c2-melt.txt` records Evilsbane hits |
| C4 15 energy now, or 15 on the next burst within 15s, and Burst sets Will to the cap | IMPLEMENTED AND TESTED | `ReviveElemEnergy` on the C4 handler and on the burst gadget. The burst revive is one frame after the drain because of queue order. `TestC1C2C4C5` |
| C6 keep Will, refill Joy, 1.25s extension, +2 Etched, 175% CRIT DMG, then consume Will while the 7s flag is up | IMPLEMENTED AND TESTED | `TestC6KeepsWill` |
| Hexerei 40% for 6s, two Hexerei members, Will at half the cap, fifth-normal second hit excluded | IMPLEMENTED AND TESTED | `UNIQUE_Lohen_Hexenzirkel_Buff`. `TestStellarA4AndHexerei`. Config pair below |
| Energy, particles, swap | IMPLEMENTED AND TESTED | `TestPhysicalOutsideMasterstroke`, `TestMasterstrokeCryoConversion`, `TestMasterstrokeEndsOnSwap` |
| `n1c_cadence=1` changes only N1→CA and CA→N1, and only when that attack starts during Masterstroke | IMPLEMENTED AND TESTED | `TestCadenceOnlyInMasterstroke`. Outside the mode, and for N2, dash, and skill cancels, the placeholder slices stay. The 15 and 31 frame numbers are an approximate partition |
| N1C hitmarks, N1→CA, CA→N1, skill, burst, plunge, and Etched windows | IMPLEMENTED BUT APPROXIMATE | No frame sheet. The Oct 14 mouse gaps are hold-start to a later tap, not those cancels. 15/31 remains a partition of the 46-frame combined interval |
| C2 vertical hitbox extent 2.5 | NOT IMPLEMENTED | The combat space is flat. The ground box is 5.5 by 5.5 |
| ProudSkill numeric rows for C1, C2, C4, C6, A1, A4, and Hexerei | BLOCKED BY MISSING EVIDENCE | `ConfigTalent` binds `%n` and does not carry the numbers. Magnitudes follow the released talent text |

## Validation runs

Each file was re-run after the Masterstroke-only cadence gate, 20 iterations, failed actions 0. These are engine results for these files. They are not a leaderboard and not a claim that the support casts are a KQM sheet. The on-field string `E 2[N1C] E 4[N1C] E 5[N1C] E` is the average combo from KQM's Lohen quick guide. `E` after Masterstroke is `skill[etched=1]`. The melt files put Burst in the first special-cast slot, which is the guide's `E/Q`. Team configs place the target at `(0, 3)` so it sits inside the forward Evilsbane box. Means move between seeds. The action and reaction counts are what these runs are for.

| Config | Profile | Team DPS mean (min–max, sd) | What the run showed |
| --- | --- | --- | --- |
| `lohen-n1c-baseline.txt` | 58f | 4978.7 (4439.4–5951.5, sd 340.3) | 13 N1 and 13 CA, all Masterstroke. 26 charged hits |
| `lohen-n1c-cadence.txt` | experimental 46f, Masterstroke only | 6556.1 (6042.7–7155.0, sd 336.4) | 17 N1 and 17 CA started. 32 Masterstroke charged hits, 2 charged hits after the mode ended |
| `lohen-standard-rotation.txt` | 58f | 12235.7 | 11 N1C. Lohen Melt mean 10 |
| `lohen-standard-cadence.txt` | experimental 46f, Masterstroke only | 13654.1 | Same action list. Lohen Melt mean 6 |
| `lohen-c0-melt.txt` | 58f | 11961.5 (Lohen 6651.2) | Lohen Melt mean 14 |
| `lohen-c1-melt.txt` | 58f | 12795.9 (Lohen 7557.4) | Same actions as C0 |
| `lohen-c2-melt.txt` | 58f | 14207.2 (Lohen 8835.2) | Evilsbane Blade mean 1 and Evilsbane Blade (melt) mean 1. Lohen Melt mean 15 |
| `lohen-c6-melt.txt` | 58f | 18805.1 (Lohen 13504.7) | Same Evilsbane counts |
| `lohen-freeze.txt` | 58f | 5065.3 | Lohen Freeze mean 8. Kokomi Freeze mean 1, Kaeya Freeze mean 1 |
| `lohen-hexerei.txt` | 58f | 14990.3 (Lohen 5628.9) | `hexerei=1` plus Fischl. Lohen Melt mean 9 |
| `lohen-hexerei-off.txt` | 58f | 13567.2 (Lohen 4593.1) | Same actions, param omitted. Lohen Melt mean 9. Lohen's own damage is lower |

Stellar Swirl and Stellar Conduct replace Cryo Swirl and Superconduct only while `stellar-swirl-enabled` and `stellar-conduct-enabled` are set. Cryo Traveler and Odette set both flags. Lohen does not. A team that lacks one of those characters still procs A4 through the plain reactions. The validation teams above do not include a Radiance character, so their logs show Melt and Freeze rather than the Stellar names. `TestA4StellarFromReactions` is the gauge-level check.

The 46-frame cadence matches the repeated Masterstroke input interval in BowTae's gamepad clip and in the 2026-10-14 mouse-overlay recordings. It is not a measured cancel pair and not a hitmark sheet. The Oct 14 taps at 44–45 and 51–56 are hold-start gaps, not N1→CA or CA→N1. It is not applied to attacks that start outside Masterstroke. Every hitmark and cancel window remains approximate. The 15 and 31 frame cancels are only how that sum was partitioned around the old hitmarks. Hitlag halt was not visible as a stable freeze in the clip, so none is applied. A real halt would make a 46-frame animation loop slower than the video, because the 46 already includes the freeze. A constant input delay drops out of the interval between two presses. The overlay can still shift that interval by a few frames, and the opening gap in the video is 51 rather than 46.
