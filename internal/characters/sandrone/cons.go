package sandrone

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

// C1. Decoding gain is halved in gainRatio. The reaction bonus lasts 2s and is refreshed
// every 1.5s while Decoding is active. The 40m range check is treated as the whole party.
func (c *char) c1Init() {
	if c.Base.Cons < 1 {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		ch.AddReactBonusMod(character.ReactBonusMod{
			Base: modifier.NewBase(c1Key, -1),
			Amount: func(ai info.AttackInfo) float64 {
				if !c.StatusIsActive(c1Key) {
					return 0
				}
				switch ai.AttackTag {
				case attacks.AttackTagDirectStellarConduct,
					attacks.AttackTagDirectStellarSwirl,
					attacks.AttackTagReactionStellarSwirl:
					return c1Bonus
				default:
					return 0
				}
			},
		})
	}
}

func (c *char) c2Init() {
	if c.Base.Cons < 2 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(c2Key, -1),
		AffectedStat: attributes.CD,
		Amount: func() []float64 {
			if !c.c2Snap {
				return nil
			}
			m[attributes.CD] = c2Base + c2Per*float64(c.c2Stacks)
			return m
		},
	})
}

// C4 fires on Sandrone's direct Stellar-Conduct and Stellar Swirl hits.
// The ability tag list is those two attack tags. Reaction Stellar Swirl is not in that list.
// The ICD is set before the cannon is queued, and the cannon's name is ignored, so it cannot retrigger.
func (c *char) c4Init() {
	if c.Base.Cons < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		e, ok := args[0].(*enemy.Enemy)
		if !ok {
			return
		}
		atk, ok := args[1].(*info.AttackEvent)
		if !ok || atk.Info.ActorIndex != c.Index() {
			return
		}
		if len(atk.Info.Abil) >= len(c4Abil) && atk.Info.Abil[:len(c4Abil)] == c4Abil {
			return
		}
		var mult float64
		var tag attacks.AttackTag
		var suffix string
		switch atk.Info.AttackTag {
		case attacks.AttackTagDirectStellarConduct:
			mult = c4Conduct
			tag = attacks.AttackTagDirectStellarConduct
			suffix = stellarConductText
		case attacks.AttackTagDirectStellarSwirl:
			mult = c4Swirl
			tag = attacks.AttackTagDirectStellarSwirl
			suffix = stellarSwirlText
		default:
			return
		}
		if c.StatusIsActive(c4ICDKey) {
			return
		}
		c.AddStatus(c4ICDKey, c4CD, true)
		ai := c.stellarInfo(c4Abil+suffix, tag, mult)
		ai.StrikeType = attacks.StrikeTypeDefault
		ap := combat.NewCircleHitOnTarget(e, nil, 3)
		c.Core.QueueAttack(ai, ap, c4Delay, c4Delay)
	}, c4Key)
}

// C6 elevation is one property on her Stellar hits. It is not also stamped onto the attack info.
func (c *char) c6Init() {
	if c.Base.Cons < 6 {
		return
	}
	c.Core.Events.Subscribe(event.OnApplyAttack, func(args ...any) {
		atk := args[0].(*info.AttackEvent)
		if atk.Info.ActorIndex != c.Index() {
			return
		}
		switch atk.Info.AttackTag {
		case attacks.AttackTagDirectStellarConduct, attacks.AttackTagDirectStellarSwirl:
			atk.Info.Elevation += c6Elevate
		default:
		}
	}, c6Key)

	c.Core.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != c.Index() {
			return
		}
		if atk.Info.AttackTag != attacks.AttackTagReactionStellarSwirl {
			return
		}
		atk.Info.Elevation += c6Elevate
	}, c6Key+"-reaction")
}
