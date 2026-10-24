package sandrone

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

// Light of Rationalisme. Sandrone enables both Stellar reactions and adds her own
// ATK-scaled Base DMG bonus. Odette's bonus, when she is also present, is a separate hook.
// Polestar Field wins over Radiance: Stellar Swirl, matching the existing Odette check.
func (c *char) stellarInit() {
	c.Core.Flags.Custom[reactable.StellarConductEnableKey] = 1
	c.Core.Flags.Custom[reactable.StellarSwirlEnableKey] = 1

	bonus := func() float64 {
		return min(c.TotalAtk()/100*jubileePer100, jubileeCap)
	}

	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		switch atk.Info.AttackTag {
		case attacks.AttackTagDirectStellarConduct, attacks.AttackTagDirectStellarSwirl:
		default:
			return
		}
		add := bonus()
		atk.Info.BaseDmgBonus += add
		if c.Core.Flags.LogDebug {
			c.Core.Log.NewEvent("sandrone stellar base damage", glog.LogCharacterEvent, c.Index()).Write("bonus", add)
		}
	}, stellarKey)

	c.Core.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.AttackTag != attacks.AttackTagReactionStellarSwirl {
			return
		}
		atk.Info.BaseDmgBonus += bonus()
	}, stellarKey+"-reaction")

	c.Core.Events.Subscribe(event.OnStellarSwirl, func(args ...any) {
		if _, ok := args[0].(*enemy.Enemy); !ok {
			return
		}
		c.AddStatus(radianceKey, reactable.RadianceSwirlDuration(c.Core, c.StatusIsActive(radianceKey)), false)
	}, stellarKey+"-radiance")
}
