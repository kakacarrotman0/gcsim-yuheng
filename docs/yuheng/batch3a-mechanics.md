# Foundation batch 3A

Character passive corrections on `cursor/batch4-engine-integration`. No character ID, ICD index, engine pin, or pipeline gate changed.

## Fischl A1, Stellar Predator

Released text: when Fischl hits Oz with a fully charged Aimed Shot, Oz fires Thundering Retribution, AoE Electro DMG equal to 152.7% of the arrow's DMG.

KQM treats that 152.7% as a raw multiplier on the aimed-shot base DMG, not as a damage-bonus stat. The bolt is a Charged Attack, Electro, 1U, no ICD, poise 40, and no particles. Bow travel bonuses use the flight from Oz to the enemy. Sources: [KQM Fischl](https://library.keqingmains.com/characters/electro/fischl) and the [A1 base-DMG evidence](https://library.keqingmains.com/evidence/characters/electro/fischl).

The old TODO was accurate. Aimed shots only hit the primary enemy. Oz is not a target, so a geometric collision is not available. `aim[oz=1]` is the trigger, the same kind of stand-in as Beidou's `counter`.

Conditions, all required:

- `oz` is not 0.
- The shot is fully charged (`hold` defaults to that).
- Ascension is at least 1.
- Oz's status is already active when the aimed shot starts, and still active when the arrow's `travel` ends.

The arrow does not hit the enemy. The bolt is queued at arrow arrival and then flies `oz_travel` (default 10). Its multiplier is the fully charged aimed-shot ratio times 1.527. It snapshots when the bolt is launched. `weakspot=1` applies to the bolt. A normal aimed shot that does not pass `oz` is unchanged.

The AoE radius is not published as a number used here, so the bolt hits the primary target only. If Oz's status ends during the arrow flight, the shot lands as an ordinary fully charged arrow with no extra travel. There is no separate Oz reaction delay between impact and the bolt.

## Beidou A1, Retribution

Released text: countering with Tidecaller at the moment of a hit grants the maximum DMG bonus.

That maximum is the skill's own 2-stack bonus. KQM puts "2 stacks / perfect parry" on one row: 750.72% at talent 9, 4 particles, poise 300. Sources: [KQM Beidou](https://library.keqingmains.com/characters/electro/beidou) and the [particle evidence](https://library.keqingmains.com/evidence/characters/electro/beidou).

`skill[counter=2]` already uses `skillbase + skillbonus*counter`, clamps at 2, emits `2+counter` particles, and sets poise to `100*(counter+1)`. At talent 9 that is 2.0672, 4.7872, and 7.5072. No second A1 multiplier was added. Incoming hits are not simulated, so the parameter remains the perfect counter.

A4's 15% Normal/Charged damage and attack speed still starts only at counter 2 and ascension 4 or higher. The A4 charged-attack windup reduction is still not implemented. It is not A1.

## Xingqiu A1, Hydropathic

Released text: when a Rain Sword shatters or expires, it heals the current character for 6% of Xingqiu's Max HP.

KQM: each sword is its own 6% heal. C0 is 3 swords. C1 adds a fourth. Shatters are one sword at a time, with a 1.5s cooldown. During Raincutter the swords stay at maximum and do not shatter. Sources: [KQM Xingqiu](https://library.keqingmains.com/characters/hydro/xingqiu), the [shatter-cooldown evidence](https://library.keqingmains.com/evidence/characters/hydro/xingqiu), and the KQM Xingqiu guide.

The orbital was one Hydro aura. It did not heal, and C1 did not change a sword count. C1's only effect is that extra sword, so the heal now counts 4 at C1. The aura is still one application.

Expiry is the orbital status ending. The next orbital tick, which for a skill cast at frame 0 is frame 988, heals every sword that was not shattered. Healing uses the existing heal formula, including Xingqiu's healing bonus. The active character at that moment is the target.

A hit on the player shatters one sword when the orbit is up, the burst is not up, and the hit has poise, a talent multiplier, or flat damage. The check is `OnPlayerHit`, which also fires while a shield holds, matching KQM. The 1.5s lock is 90 frames and is not extended by hitlag. Zero-damage aura pulses do not shatter a sword. `OnPlayerHit` does not carry the rolled damage, so this poise/multiplier/flat-damage check is the stand-in for "more than 0 DMG and/or Poise DMG".

Raincutter suppresses shattering and refills the swords to the maximum. It does not delete the expiry heal. The burst status lasts 33 frames longer than the orbit; the heal is tied to the orbit, not to that extra window.

Damage reduction, interruption resistance, and per-sword HP are not simulated. Hydro orbital ticks continue after every sword has shattered. Standard rotations that never hit the player still heal the full set when the orbit ends.

## Left for batch 3B

Yun Jin A1 and Candace A1 are the same kind of gap as Beidou, not as Fischl or Xingqiu. Both already have `skill[perfect=1]`. Yun Jin's parameter forces the level-2 charged hit on the tap animation. Candace's parameter forces the charged hit and removes 55 frames of windup. Neither has a test file. Diona A4 is different: opponents inside Signature Mix lose 10% ATK for 15s, and that debuff is absent. Enemy ATK does not change the damage these characters deal.

Batch 3B should pin Yun Jin `perfect=1` and Candace `perfect=1`, and should cite a frame source before treating Candace's 55-frame cut as the instant charge. It should not add a hit detector, and it should not implement Diona A4 as a damage change.
