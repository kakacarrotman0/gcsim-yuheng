package vodyanitsa

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) onSkillCast() {
	if c.Base.Ascension >= 1 && c.vortexPresent() {
		c.wandering = true
		c.shredAnemo()
	}
	if c.Base.Ascension < 4 {
		return
	}
	c.lead = leadMax
	c.chorus = chorusMax
	c.AddStatus(concertoKey, concertoDur, false)
}

func (c *char) a1Init() {
	c.Core.Events.Subscribe(event.OnStellarSwirl, func(args ...any) {
		if c.Base.Ascension < 1 || !c.songActive() {
			return
		}
		c.wandering = true
		c.shredAnemo()
	}, "vodyanitsa-a1-swirl")

	c.Core.Events.Subscribe(event.OnStellarVortexDetonate, func(args ...any) {
		if !c.wandering {
			return
		}
		c.shredAnemo()
		c.alterUntil = c.Core.F + a1AlterDur
		c.wandering = false
	}, "vodyanitsa-a1-detonate")
}

func (c *char) vortexPresent() bool {
	for _, g := range c.Core.Combat.Gadgets() {
		if g != nil && g.GadgetTyp() == info.GadgetTypStellarVortex {
			return true
		}
	}
	return false
}

func (c *char) shredAnemo() {
	for _, t := range c.Core.Combat.Enemies() {
		e, ok := t.(*enemy.Enemy)
		if !ok {
			continue
		}
		e.AddResistMod(info.ResistMod{
			Base:  modifier.NewBase(anemoShredKey, a1AnemoDur),
			Ele:   attributes.Anemo,
			Value: -a1AnemoShred,
		})
	}
}

func (c *char) a4Init() {
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		if c.Base.Ascension < 4 || !c.StatusIsActive(concertoKey) {
			return
		}
		atk := args[1].(*info.AttackEvent)
		active := c.Core.Player.Active() == atk.Info.ActorIndex
		var stacks *int
		if active {
			stacks = &c.lead
		} else {
			stacks = &c.chorus
		}
		if *stacks <= 0 {
			return
		}
		if !c.concertoHit(atk) {
			return
		}
		atk.Info.FlatDmg += c.concertoFlat(c.alterActive())
		*stacks--
	}, "vodyanitsa-a4")
}

func (c *char) concertoHit(atk *info.AttackEvent) bool {
	if c.alterActive() {
		switch atk.Info.AttackTag {
		case attacks.AttackTagReactionStellarSwirl, attacks.AttackTagDirectStellarSwirl:
			return true
		default:
			return false
		}
	}
	if atk.Info.Element != attributes.Hydro && atk.Info.Element != attributes.Cryo {
		return false
	}
	switch atk.Info.AttackTag {
	case attacks.AttackTagNormal,
		attacks.AttackTagExtra,
		attacks.AttackTagPlunge,
		attacks.AttackTagElementalArt,
		attacks.AttackTagElementalArtHold,
		attacks.AttackTagElementalBurst:
		return true
	default:
		return false
	}
}

func (c *char) concertoFlat(stellar bool) float64 {
	hp := c.MaxHP()
	if hp <= a4MinHP {
		return 0
	}
	perKilo := a4HydroPerKilo
	cap := a4HydroCap
	if stellar {
		perKilo = a4StellarPerKilo
		cap = a4StellarCap
	}
	bonus := (hp - a4MinHP) * perKilo * 0.001
	if bonus > cap {
		return cap
	}
	return bonus
}
