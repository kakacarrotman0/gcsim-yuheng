package kachina

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	playercharacter "github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/core/player/shield"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c1EnergyKey = "kachina-c1-energy"
	c4DefKey    = "kachina-c4-def"
	c6ICDKey    = "kachina-c6-icd"
	c2Points    = 20.0
	a1Ratio     = 0.2
	a1Duration  = 12 * 60
	c6Radius    = 6.0
)

func (c *char) initC1() {
	if c.Base.Cons < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnShielded, func(args ...any) {
		sh, ok := args[0].(shield.Shield)
		if !ok || sh.Type() != shield.Crystallize {
			return
		}
		c.c1Energy()
	}, "kachina-c1-shard")
	c.Core.Events.Subscribe(event.OnLunarCrystallize, func(_ ...any) {
		c.c1Energy()
	}, "kachina-c1-lunar")
}

func (c *char) c1Energy() {
	if c.StatusIsActive(c1EnergyKey) {
		return
	}
	c.AddStatus(c1EnergyKey, 5*60, true)
	c.AddEnergy(c1EnergyKey, c1Energy)
}

func (c *char) initC4() {
	if c.Base.Cons < 4 {
		return
	}
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		if len(args) < 2 {
			return
		}
		if prev, ok := args[0].(int); ok {
			c.Core.Player.ByIndex(prev).DeleteStatMod(c4DefKey)
		}
		if next, ok := args[1].(int); ok {
			c.applyC4(next)
		}
	}, "kachina-c4-swap")
}

func (c *char) applyC4(index int) {
	if c.Base.Cons < 4 || !c.StatusIsActive(fieldKey) {
		return
	}
	if c.Core.Combat.Player().Pos().Distance(c.fieldCenter) > fieldSlamR {
		return
	}
	c.Core.Player.ByIndex(index).AddStatMod(playercharacter.StatMod{
		Base:         modifier.NewBaseWithHitlag(c4DefKey, -1),
		AffectedStat: attributes.DEFP,
		Amount: func() []float64 {
			if !c.StatusIsActive(fieldKey) {
				return nil
			}
			if c.Core.Combat.Player().Pos().Distance(c.fieldCenter) > fieldSlamR {
				return nil
			}
			enemies := len(c.Core.Combat.EnemiesWithinArea(combat.NewCircleHitOnTarget(c.fieldCenter, nil, fieldSlamR), nil))
			if enemies <= 0 {
				return nil
			}
			if enemies > 4 {
				enemies = 4
			}
			mod := make([]float64, attributes.EndStatType)
			mod[attributes.DEFP] = c4Ratio[enemies-1]
			return mod
		},
	})
}

func (c *char) initC6() {
	if c.Base.Cons < 6 {
		return
	}
	c.Core.Events.Subscribe(event.OnShielded, func(args ...any) {
		sh, ok := args[0].(shield.Shield)
		if !ok {
			return
		}
		key := c6ShieldKey(sh)
		if c.c6Seen[key] {
			c.triggerC6(sh)
		}
		c.c6Seen[key] = true
	}, "kachina-c6-shielded")
	c.Core.Events.Subscribe(event.OnShieldBreak, func(args ...any) {
		sh, ok := args[0].(shield.Shield)
		if !ok {
			return
		}
		delete(c.c6Seen, c6ShieldKey(sh))
		c.triggerC6(sh)
	}, "kachina-c6-break")
}

func c6ShieldKey(sh shield.Shield) string {
	return fmt.Sprintf("%d:%d", sh.Type(), sh.ShieldTarget())
}

func (c *char) triggerC6(sh shield.Shield) {
	if c.StatusIsActive(c6ICDKey) {
		return
	}
	active := c.Core.Player.Active()
	target := sh.ShieldTarget()
	if target != -1 && target != active {
		return
	}
	c.AddStatus(c6ICDKey, int(c6ICD*60), true)
	ai := info.AttackInfo{
		ActorIndex:     c.Index(),
		Abil:           "This Time, I've Gotta Win",
		AttackTag:      attacks.AttackTagExtra,
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeBlunt,
		PoiseDMG:       100,
		Element:        attributes.Geo,
		Durability:     25,
		Mult:           c6Ratio,
		UseDef:         true,
		IgnoreInfusion: true,
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, c6Radius), 0, 0)
}
