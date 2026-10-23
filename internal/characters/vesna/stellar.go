package vesna

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

// Stellar Jubilee: Splendid Prelude has no promote level on the skill depot,
// the same encoding as an always-unlocked utility passive.
// It enables Stellar Swirl, adds min(ATK/100*0.007, 0.14) base damage to
// direct and reaction Stellar Swirl from any party member, and enters
// Radiance: Stellar Swirl for 8s after a nearby Stellar Swirl.
// Vesna does not enable Stellar Conduct. Superconduct stays Superconduct.
func (c *char) stellarInit() {
	c.Core.Flags.Custom[reactable.StellarSwirlEnableKey] = 1

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.AttackTag != attacks.AttackTagDirectStellarSwirl {
			return
		}
		atk.Info.BaseDmgBonus += c.jubileeBonus()
	}, "vesna-jubilee")

	c.Core.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.AttackTag != attacks.AttackTagReactionStellarSwirl {
			return
		}
		atk.Info.BaseDmgBonus += c.jubileeBonus()
		// C6 elevation is written on this copy only. The queued reaction
		// attack does not copy Elevation back, so calc() does not apply it twice.
		if c.Base.Cons >= 6 && atk.Info.ActorIndex == c.Index() {
			atk.Info.Elevation += c6Elevation
		}
	}, "vesna-jubilee-reaction")

	c.Core.Events.Subscribe(event.OnStellarSwirl, func(args ...any) {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return
		}
		// OnStellarSwirl is emitted before the reaction damage is snapshotted,
		// so Truth Prevails is already active for that same reaction.
		c.AddStatus(radianceKey, reactable.RadianceSwirlDuration(c.Core, c.StatusIsActive(radianceKey)), false)
	}, "vesna-radiance")
}

func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}
	c.AddStatMod(character.StatMod{
		Base: modifier.NewBase("vesna-truth", -1),
		Amount: func() []float64 {
			m := make([]float64, attributes.EndStatType)
			if !c.inRadiance() {
				return m
			}
			atk, em := c.truthAmounts()
			m[attributes.ATKP] = atk
			m[attributes.EM] = em
			return m
		},
	})
}

func (c *char) truthAmounts() (float64, float64) {
	var same, other int
	for _, ch := range c.Core.Player.Chars() {
		switch ch.Base.Element {
		case attributes.Anemo, attributes.Cryo:
			same++
		default:
			other++
		}
	}
	// The graph caps Cryo+Anemo at 5 and then zeroes the other-element term.
	// A party of 4 cannot reach that cap.
	if same > 5 {
		same = 5
		other = 0
	}
	atk := a4ATK * float64(same)
	em := a4EM * float64(other)
	// C4's proud ratio is an unbound special. The talent text says the bonuses become 3×.
	// The additive graph form is base*count*(1+C4_PowerUpRatio), so the ratio is 2.
	if c.Base.Cons >= 4 {
		atk *= c4Mult
		em *= c4Mult
	}
	return atk, em
}
