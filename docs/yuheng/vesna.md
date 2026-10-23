# Vesna

Yuheng-maintained gcsim character. Upstream `genshinsim/gcsim` `main` has no Vesna implementation.

## Identity

| Field | Value |
| --- | --- |
| English name | Vesna |
| Key | `vesna` |
| Avatar ID | 10000143 |
| Skill depot | 14301 |
| Element | Anemo |
| Weapon | Sword |
| Rarity | 5 |
| Region | Snezhnaya (`ASSOC_TYPE_SNEZHNAYA_STAR`) |
| Ascension stat | CRIT Rate |
| Constellation | Amentum Vernum |

In-game talent IDs: attack 11431, skill 11432, burst 11435. Windborne occupies the skill button as 11433, 11434, and 11436. C6 Transpose switches the normal to 11430. Energy 60. Skill cooldown 18s. Burst cooldown 15s. Armed for Action lasts 15s.

## Data provenance

Numbers, identity, and talent text are from `https://gi.yatta.moe/api/v2/en/avatar/10000143`.

Mechanics are from `ConfigAbility_Avatar_Vesna.json` and `ConfigGadget_SkillObj_Vesna.json` on `DimbreathBot/AnimeGameData` commit `792978e5503ecfba73dcb3562ed44a0d35a2abe2`. Base and promotion stats are from the Yatta promote block, which matches that avatar's excel rows. Proud groups are 14331 (normal), 14332 (skill), and 14339 (burst). Groups 11431, 11432, and 11439 in older dumps are Skirk (`SkirkNew_SkillUpgrade_*`), not Vesna.

`excel-hk4e` in `go.mod` predates this character, so `go run ./pipeline` was not executed. Talent tables in `zz_vesna.dm.go` are hand-maintained from the Yatta proud-skill rows.

The KQM library page was used for the released talent wording. Its findings are not added, so it is not a frame sheet.

C1's 20% reaction bonus, C2's 40% ATK, C4's 3× passive, C6's 20% elevation, and Transpose's 150% / 200% hits are released talent text. Those ability specials are 0 in the graph because the proud constellation rows are not in the excel. Sword Essence amounts (60, 30, 90, 5, and a cost of 30) are ability literals. The pinion's +5 also matches skill parameter 13.

## Validated scope

Armed for Action, Sword Essence, Windborne Sword levels 1–3, spirit blades, the burst's one spirit blade, wind pinions, and C1–C6 where the released text and the graph agree. While Armed is active, Normal, Charged, and Plunging attacks are Anemo and ignore weapon infusion. Outside Armed those attacks stay Physical and can be infused.

Direct Stellar Swirl is used for spirit blades only during Radiance. The same talent percentage is kept. The hit ignores DEF, has no gauge, takes the EM reaction bonus, can crit, and does not take a DMG% bonus. It is not an ordinary Swirl and it is not the transformative Stellar Swirl reaction, which stays on the level-base formula in `pkg/reactable`.

`docs/yuheng/configs/vesna-natural.txt` is a mechanics smoke test: Vesna / Kaeya / Bennett / Xiangling, 20 iterations, 20 seconds. Essence is earned from the skill and from pinions. Burst is not cast. Kaeya applies Cryo before Vesna enters Armed, because a swap during Armed would end it. Bennett and Xiangling do not take the field. It is not a damage benchmark.

## Timing

| Timing | Class |
| --- | --- |
| Energy 60, skill CD 18s, burst CD 15s, Armed 15s, essence 60/30/90/5 and a cost of 30 | Ability literals and talent parameters |
| Disciplinary stacks: 20s, max 6, +10% spirit-blade ratio per stack, independent durations | Proud skill 14321 and the damage handler |
| Lv2 spirit blade 30 frames after the slash | Gadget think interval 0.5s. The slash's own frame is a placeholder |
| Burst spirit blade at frame 33 | Gadget think interval 0.55s, one hit, then the modifier removes itself |
| C6 window 5s | Literal `SwitchToD_Duration` on `ElementalArt_D` |
| Radiance 8s, extendable to 12s only by Vodyanitsa's existing enter rule | `RadianceSwirlDuration` |
| Normal, charged, plunge, skill, Windborne, Lv3, and Transpose hitmarks | Unverified. See the frame table below |
| Particles | One Anemo particle (`baseEnergy` 3, config 2021) on the graph's art, pinion, spirit-blade, Transpose, and burst tags. Unique modifier 2.5s. Normals, charged attacks, and plunges are not in that list |

Placeholder frames, all unverified:

| Window | Frames |
| --- | --- |
| N1–N6 hitmarks | 8, 12, 10/20, 16, 22, 28 |
| Charged hit / animation | 18 / 40 |
| Plunge collision / low / high | 30 / 36 / 38 |
| Skill and Windborne animation; cancel into attack, charge, skill, or burst; dash | 60; 20; 18 |
| Skill entry hit, Lv1, Lv2 slash | 14, 12, 12 |
| Lv2 blade | 42 |
| Lv3 blades | 14, 24, 34, 44, final 58 |
| Pinion | 12 |
| Transpose slash / blade | 16 / 46 |
| Armed exit after the capping Lv3 | 60 |

The Lv3 checker attached to animator state `Vesna_ElementalArt_C` lasts 1.5s. That state length is not a measured cancel. This implementation ends Armed at frame 60 after the capping Lv3 and blocks another Windborne as soon as that use is cast. A cancel before frame 60 can still be Transpose, and Transpose still summons a pinion if Armed has not ended.

## Mechanics notes

- Entering Armed clears Disciplinary stacks, then C2 grants six stacks if Rite of Spring's Procession is unlocked. Essence already held is kept and the skill adds 60, capped at 90. The mode starts at Windborne level 1.
- Each Windborne costs 30 and raises the level by one, capped at 3. Only level 3 spends one of the three uses (four at C1). The capping use blocks further Windborne immediately. Leaving the field, the 15s timer, or that delayed exit clears essence. Disciplinary stacks are not cleared by the exit. They are cleared by the next entry and by swapping.
- C1's first level-3 cast in that entry does not spend essence. A and B still spend 30.
- Windborne does not restart the 18s skill cooldown.
- Spirit blades outside Radiance are ordinary Anemo talent hits with 1U, DEF, and DMG%. During Radiance they are Direct Stellar Swirl: no gauge, DEF ignored, EM reaction bonus, no DMG%. C1's 20% is a reaction bonus on Vesna's stellar tags and only while Armed. C6's 20% is elevation on her stellar reaction damage at all times. Elevation is applied once. The queued Stellar Swirl attack does not also carry it.
- Jubilee is available at ascension 0. It sets `stellar-swirl-enabled` and adds `min(ATK/100×0.007, 0.14)` base damage to party Direct Stellar Swirl and to reaction Stellar Swirl. It does not set `stellar-conduct-enabled`. Pyro, Hydro, and Electro Swirl stay ordinary. Superconduct stays Superconduct.
- Truth Prevails is active only during Radiance. Each Cryo or Anemo member, including Vesna, gives 6% ATK. Each other element gives 25 EM. C4 makes both 3× and requires the passive, so it does nothing below ascension 4. `OnStellarSwirl` runs before that reaction is snapshotted, so the buff is included in the swirl that enters Radiance.
- The burst adds 30 essence even while Armed is down, grants one Disciplinary stack, and fires one spirit blade. Declared specials of 6s and 0.5s are not referenced by the ability or gadget file and are not implemented as extra hits.
- A pinion is one ordinary 1U Anemo hit per Normal, Charged, Plunging, or armed Transpose action, plus 5 essence. The graph swaps the bullet id during Radiance and the particle list names `SmallSword_Star`, but the extracted small-sword hit is the 1U art hit. That star bullet is not given a Direct Stellar Swirl formula.
- Transpose is one 150% Anemo art hit plus one spirit blade at 200%, with the Disciplinary factor on the blade only. It does not cost essence, raise Windborne's level, spend a level-3 use, or add a stack. It adds a pinion only while Armed. The extra `D_Sword_2` tag is not a second blade. The talent text says one additional instance.
- At six Disciplinary stacks, C2 gives 40% ATK. A new stack at the cap of six is rejected and does not refresh the older timers.
- ICD group `Vesna_MagicSword` has no extracted reset or sequence. Hits use `ICDGroupDefault`. That can mis-count gauge.
- Coursing, the quest-modified skill, and the 0.5 interruption resistance are not implemented. The 40m "nearby" check cannot fail: gcsim has one player position.

## Completion

| Item | Status | Source and check |
| --- | --- | --- |
| Identity, sword, Anemo, 60 energy, CRIT Rate ascension, level-90 HP and DEF | IMPLEMENTED AND TESTED | Yatta avatar 10000143. `TestIdentity` |
| Armed Anemo infusion that a weapon infusion cannot replace, Physical outside | IMPLEMENTED AND TESTED | `TestArmedInfusionAndParticles` |
| Essence 60/30/90, Windborne cost 30, pinion +5, burst +30 outside Armed | IMPLEMENTED AND TESTED | `TestNaturalWindborne`, `TestEssenceCapAndBurst` |
| Windborne levels, Lv3 ×4 plus final, one Lv2 blade, prerequisites | IMPLEMENTED AND TESTED | `TestNaturalWindborne` |
| C1 four uses and a free first Lv3, 15s exit, swap clearing Armed, essence, stacks, and Transpose | IMPLEMENTED AND TESTED | `TestC1UsesAndDuration`, `TestSwapResets` |
| C2 six stacks and 40% ATK, absent below ascension 1, stacks surviving the Armed exit | IMPLEMENTED AND TESTED | `TestC2AndTruth` |
| Truth Prevails during Radiance, C4 ×3, absent below ascension 4 | IMPLEMENTED AND TESTED | `TestC2AndTruth` |
| C3 and C5 talent indexes | IMPLEMENTED AND TESTED | `TestTalentCons` |
| C6 Transpose, one blade, pinion only while Armed, window consumed | IMPLEMENTED AND TESTED | `TestC6Transpose` |
| Ordinary Anemo hit uses DMG% and DEF. Direct Stellar Swirl does not, ignores DEF, uses the EM bonus, Jubilee, and crit | IMPLEMENTED AND TESTED | `TestStellarTagsAndFormula`, `TestC6DirectAndReaction` |
| C6 elevation once on the reaction contribution, not on the queued swirl attack | IMPLEMENTED AND TESTED | `TestC6DirectAndReaction` |
| Cryo Swirl becomes Stellar Swirl. Pyro Swirl and Superconduct do not | IMPLEMENTED AND TESTED | `TestSwirlSplit` |
| One Anemo particle and a 2.5s ICD | IMPLEMENTED AND TESTED | `TestArmedInfusionAndParticles` |
| Plunge and charged Anemo while Armed, one pinion each | IMPLEMENTED AND TESTED | `TestPlungeAndCharge` |
| Hitmarks and cancel windows | IMPLEMENTED BUT APPROXIMATE | No frame sheet |
| `Vesna_MagicSword` ICD sequence | NOT IMPLEMENTED | No coefficient row was extracted. Default 2.5s / 3-hit is used |
| Star-mode pinion as Direct Stellar Swirl | NOT IMPLEMENTED | Bullet id changes. The extracted hit is 1U Anemo |
| Burst 6s / 0.5s specials as extra blades | NOT IMPLEMENTED | Specials are unreferenced. The gadget fires once |
| Coursing, quest skill, interruption resistance | NOT IMPLEMENTED | Not combat, or not a damage event |
| 40m party distance | NOT IMPLEMENTED | One player position |

## Validation

`vesna-natural.txt` was run with `go run ./cmd/gcsim`, 20 iterations, 20 seconds. Failed actions were 0 and the warning list was empty.

Vesna's action counts were 4 skills and 6 normals. Damage-source means:

- The Art of Victory 1, Windborne Sword Lv1 1, Windborne Sword Lv2 1
- Spirit Blade Lv2 1, Spirit Blade Lv3 4, Spirit Blade Lv3 Final 1
- Wind Pinion 6
- Armed Normal 0 through 5 (the third normal is two hits)
- Stellar Swirl 4, Stellar Swirl Detonation 2
- Kaeya Frostgnaw 1
- No Physical normal, no ordinary Swirl, no Burst

The printed summary was average 156698.00 damage over 20.00 seconds (7834.90 dps, min 6970.73, max 8255.13, std 448.39). That number is the output of this one config. It is not a claimed team DPS.

Unit tests in `vesna_test.go` are the mechanics checks. `go test` on `internal/characters/lohen`, `internal/characters/linnea`, `internal/characters/vodyanitsa`, and `internal/tests/reactable` passed. Ordinary Pyro Swirl and Superconduct are the `TestSwirlSplit` cases. This config does not swap in Bennett, so it does not show that Pyro Swirl.
