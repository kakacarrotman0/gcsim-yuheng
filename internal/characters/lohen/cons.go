package lohen

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func attacksTagNormal() attacks.AttackTag { return attacks.AttackTagNormal }
func attacksTagExtra() attacks.AttackTag  { return attacks.AttackTagExtra }

func (c *char) c6Init() {
	if c.Base.Cons < 6 {
		return
	}
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("lohen-c6", -1),
		Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			if !c.masterActive() {
				return nil
			}
			switch {
			case len(atk.Info.Abil) >= 6 && atk.Info.Abil[:6] == "Etched":
			case len(atk.Info.Abil) >= 8 && atk.Info.Abil[:8] == "Manifest":
			default:
				return nil
			}
			m := make([]float64, attributes.EndStatType)
			m[attributes.CD] = c6Crit
			return m
		},
	})
}

func (c *char) c4OnEnter() {
	if c.Base.Cons < 4 {
		return
	}
	if c.Energy < c.EnergyMax {
		c.AddEnergy("lohen-c4", c4Energy)
		c.c4RefundUntil = 0
		return
	}
	c.c4RefundUntil = c.Core.F + c4Window
}

func (c *char) armEvilsbane() {
	if c.Base.Cons < 2 || c.Core.F < c.c2ICDUntil {
		return
	}
	c.c2Until = c.Core.F + c2Window
	c.c2ICDUntil = c.c2Until
}

func (c *char) tryEvilsbane() {
	if c.Base.Cons < 2 || !c.masterActive() || c.Core.F >= c.c2Until {
		return
	}
	c.c2Until = 0
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Evilsbane Blade",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Cryo,
		Durability: 25,
		Mult:       c2Ratio,
	}
	// Radius is not in the attack pattern that was extracted. 1U is from Yatta.
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3)
	c.Core.QueueAttack(ai, ap, 0, 0)
	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = c2EM
	for _, char := range c.Core.Player.Chars() {
		if char.Index() == c.Index() {
			continue
		}
		char.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(c2EMKey, c2Duration),
			AffectedStat: attributes.EM,
			Amount: func() []float64 {
				return m
			},
		})
	}
}
