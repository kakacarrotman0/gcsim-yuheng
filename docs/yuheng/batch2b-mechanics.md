# Foundation batch 2B

One shared-mechanics correction on `cursor/batch4-engine-integration`. No character ID, ICD index, engine pin, or pipeline gate changed.

## Root cause

`LcExtraHitOverride` and `LcrExtraHitOverride` were both the string `lunarcrystallize-bonus-hit-chance`. Columbina's A4 writes `0.33` through the Lunar-Charged constant on `OnLunarCharged` and through the Lunar-Crystallize constant on `OnLunarCrystallize`. Both writes hit the same map entry. Deleting either name deleted both.

KQM's Law of the New Moon lists the two 33% chances separately. Lunar-Charged is an extra thundercloud strike. Lunar-Crystallize is an extra hit on each Moondrift Harmony wave. The source is the KQM Columbina entry, Law of the New Moon: https://library.keqingmains.com/characters/hydro/columbina

The readers are `DoLCAttack` and `DoLCrAttackWithContrib`. Each rolls `core.Rand.Float64() < chance` and queues one more copy of that reaction's own damage function. The damage formulas were not changed.

## Activation and expiry

A4 runs only at ascension 4 or higher. It arms a chance only when Columbina's `columbina-q-buff` react-bonus mod is active. That mod is the Lunar Domain check: `applyBurstBuff` adds it for 1.2s while the active character is inside the burst area, and the burst refreshes it every 0.8s. A reaction of one type, while the buff is absent, deletes only that type's chance.

Before this batch the armed value stayed in the map until a later reaction of either type. A Lunar-Crystallize reaction outside the domain cleared the Lunar-Charged chance, and thundercloud strikes after the domain ended still used 33% until some later reaction happened to delete the shared key.

Each domain-buff application now schedules a cleanup for one frame after that application would expire. The task deletes both chances only when the buff is already inactive, so a refresh does not clear them. Hitlag extends that char task together with the buff. The 33% values and the reaction damage functions are unchanged.

The chance is still one global value per reaction, not a snapshot stored on one thundercloud or one harmony. A strike that happens while the 1.2s linger is still up uses the armed chance. A strike after that linger does not, including a cloud that was created earlier. Per-cloud ownership was not added.

## Seed 1 strike counts

Forty `DoLCAttack` calls and forty `DoLCrAttackWithContrib` calls, seed 1. A Lunar-Charged strike with no extra roll is 1 hit. A Lunar-Crystallize call with no extra roll is 3 hits, so 40 calls are 120 hits.

| Armed chance | Lunar-Charged hits | Lunar-Crystallize hits |
| --- | --- | --- |
| Lunar-Charged only | 54 | 120 |
| Lunar-Crystallize only | 40 | 161 |
| Both | 54 | 160 |
| Neither | 40 | 120 |

Clearing Lunar-Crystallize leaves the 54 Lunar-Charged hits and returns Lunar-Crystallize to 120. Clearing Lunar-Charged leaves the 161 Lunar-Crystallize hits and returns Lunar-Charged to 40. After the domain buff expires, the same 40 calls produce 40 and 120.

Lunar-Charged is rolled first. Its extra rolls are why "both" has 160 Lunar-Crystallize hits instead of 161. Each reader checks only its own flag.

A Lunar-Bloom while the domain is up still grants 1 Moonridge Dew and no Lunar-Charged or Lunar-Crystallize strike. After the domain expires, that bloom grants no Moonridge Dew.

## Left unchanged

`TryAddLC` still attaches Hydro inside the reaction function. The cloud tick still uses the player's current position and an 8-unit radius. Ineffa's A4 timing TODO remains. The checked-in Lunar-Charged multiplier is still 3 and the Lunar-Crystallize multiplier is still 1.6. Neither was re-derived.
