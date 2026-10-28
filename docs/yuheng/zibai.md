# Zibai engine foundation

Upstream source: `genshinsim/gcsim` `2350be9e835a18436a16de60fc8eecb2da1443be`.

The nine package files under `internal/characters/zibai/` are that commit's character package, copied without formula, gauge, frame, or particle changes. Registration is adapted to this fork.

## Keys and ICD

`keys.Zibai` is appended after Sandrone. Zhongli stays 130. Nefer stays 131. Vesna stays 132. Sandrone stays 133. Zibai is 134. `InvalidChar` moves from 134 to 135 because it is the sentinel after the new key. No earlier playable ID moves.

Shortcut, import, and avatar-map entries are alphabetical. That does not change iota values.

Every Zibai attack file uses `attacks.ICDGroupDefault`. No ICD const, timer, or sequence was added. `ICDGroupLohenSkill` stays 65. `ICDGroupSandroneMachineGun` stays the last group.

The catalog avatar id `10000126` and depot `12601` come from the upstream catalog entry. They are not the engine key. The icon is `UI_AvatarIcon_Zibai`. Talent ids are attack `11261`, skill `11262`, burst `11265`. Burst energy in that entry is 60. The ascension stat is CRIT DMG, 38.4% at the level-90 promotion.

## Pipeline

`pinnedCharacterSlugs` ends in `Sandrone`, `Zibai`. `validateCharacterKeyNames` rejects a list that inserts Zibai ahead of Zhongli. Full regeneration was not run. `task pipeline` and `scripts/pipeline-clean.sh` remain closed unless `GCSIM_ALLOW_FULL_REGEN=1`, and an open gate still refuses a character or ICD list that does not match the pin. Zibai's `config.yml` is hand-maintained. A generator that sorted that file alphabetically would miss the tail pin and fail closed.

## Mechanics that the tests execute

Radiance is never assigned by a test. It comes from the 6-frame ticker, Normal Attack hits, Lunar-Crystallize, or C1's skill grant.

- C0 Normal Attacks outside the mode are Physical and use the default ICD group. Inside the mode they are Geo, DEF-scaled, infusion-locked, and blunt. A connecting Normal adds 5 radiance, then the 0.5s ICD holds.
- Charged Attacks outside the mode are Physical. Inside the mode they are Geo and use the skill talent row.
- The opening skill starts Lunar Phase Shift at frame 18, starts the 18s skill cooldown at that frame, and adds 1 radiance. Spirit Steed's Stride is not ready until radiance is at least 70. The failure reason is `InsufficientStamina`, which is the upstream `ActionReady` result.
- Ticker-only generation reaches 70 at frame 432. The A1 status, applied at frame 18 for 4s, has already expired by then. Weaving Normals and Hydro so the Geo hits become Lunar-Crystallize reaches 70 while A1 is still active. The stride spends 70. Hit 1 is an Elemental Art with 25 gauge at apply time. Hit 2 is Direct Lunar-Crystallize, 0 gauge, full DEF ignore. Four strides end the mode. The next skill is then on skill cooldown.
- Burst has two DEF-scaled Geo hits. The second is Direct Lunar-Crystallize. While the mode is active, burst extends it by 102 frames. Burst cooldown is 15s. Starting energy is the simulator's full bar, and the burst drains it.
- Leaving the field deletes the mode.
- A Hydro aura plus a Geo hit is ordinary Hydro Crystallize when Zibai is absent and Lunar-Crystallize when she has been initialized. The shared flag is `LunarCrystallizeEnableKey`. This package does not change the reactable implementation. The flag is set even at ascension 0.
- A4 adds 15% DEF per other Geo character and 60 EM per other Hydro character.
- A1's stride-hit flat damage is 60% DEF while the 4s status is active. C2 under Moonsign level 2 raises that to 550% DEF. `a1StrideBonusDmg` returns 0 when ascension is below 1, so the C2 term cannot apply without A1.
- C2 also gives the party 30% reaction bonus on Lunar-Crystallize and Direct Lunar-Crystallize while the mode is active.
- C1 grants 100 radiance when the mode starts and allows 5 strides. Its 220% reaction bonus applies to the first stride's second hit, then the status ends one frame later.
- C4 keeps the Normal counter during the mode. After a stride hits an enemy, the Gleam N4 extra hit uses 2.5 times the base extra multiplier and consumes the status.
- C6 multiplies radiance gain by 1.5, so C1's 100 becomes 150 and is capped at 100. Stride spends the whole bar. Elevation for the next 3s is 1.6% per point above 70. At 100 radiance that is 0.48.
- Particles are 1 Geo, a 67% roll, and a 2s ICD. The seed-1 test observes one particle and no second particle inside that ICD. Energy restoration uses the shared particle system.

`docs/yuheng/configs/zibai-c0.txt` parses and runs skill, Normal, Charged, and Burst. It does not cast Stride, because that cast is illegal before 70 radiance.

## Not re-measured

Frame data in `config.yml` is credited to mikeskies (sheet `1aHOBC33Q8ML4GTMPlgcmqc5TdkbZJk8pBtK3kgr_ALk`, videos `29WWu2pYiRs`, `SyLeTDRbR0w`, `0wGPOVGF2xw`). This assignment did not re-time those animations.

The inactive `radiance` condition returns the integer 0. That is the upstream `return 0, nil`. While the mode is active the same condition returns the float bar.

Zibai and Linnea both set the same Lunar-Crystallize flag and each adds a DEF bonus. Those bonuses stack if both characters are present. The shared reaction was not rewritten.

No damage number from these tests is a statement about team DPS.
