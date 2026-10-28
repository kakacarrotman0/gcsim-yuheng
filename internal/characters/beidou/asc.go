package beidou

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

// A1, Retribution: countering on the frame a hit lands grants Tidecaller's maximum DMG bonus.
// That bonus is the skill's own 2-stack value, not a second multiplier. Incoming hits are not
// simulated, so skill[counter=2] is the perfect counter: base + 2*bonus, 4 particles, poise 300.
// KQM lists 2 stacks and a perfect parry on the same row (750.72% at talent 9).

// Gain the following effects for 10s after unleashing Tidecaller with its maximum DMG Bonus:
// - DMG dealt by Normal and Charged Attacks is increased by 15%. ATK SPD of Normal and Charged Attacks is increased by 15%.
// TODO: - Greatly reduced delay before unleashing Charged Attacks.
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}

	mDmg := make([]float64, attributes.EndStatType)
	mDmg[attributes.DmgP] = .15
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag("beidou-a4-dmg", 600),
		Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			if atk.Info.AttackTag != attacks.AttackTagNormal && atk.Info.AttackTag != attacks.AttackTagExtra {
				return nil
			}
			return mDmg
		},
	})

	mAtkSpd := make([]float64, attributes.EndStatType)
	mAtkSpd[attributes.AtkSpd] = .15
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBaseWithHitlag("beidou-a4-atkspd", 600),
		AffectedStat: attributes.AtkSpd,
		Amount: func() []float64 {
			return mAtkSpd
		},
	})
}
