package nefer

import (
	"strings"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

func (c *char) moonsignInit() {
	c.Core.Flags.Custom[reactable.LunarBloomEnableKey] = 1
	c.Core.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.AttackTag != attacks.AttackTagDirectLunarBloom {
			return
		}
		em := c.Stat(attributes.EM)
		bonus := min(em*0.000175, 0.14)
		atk.Info.BaseDmgBonus += bonus
		if c.Core.Flags.LogDebug {
			c.Core.Log.NewEvent("nefer lunarbloom base damage", glog.LogCharacterEvent, c.Index()).Write("bonus", bonus)
		}
	}, "nefer-lunarbloom-bonus")
}

func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnDendroCore, func(args ...any) {
		if !c.StatusIsActive(deceitKey) {
			return
		}
		if d, ok := args[0].(interface{ MakeDeceit() }); ok {
			d.MakeDeceit()
		}
		c.Core.Tasks.Add(func() {
			if c.StatusIsActive(deceitKey) {
				c.markCores()
			}
		}, 1)
	}, "nefer-deceit")
}

func (c *char) buffInit() {
	// Burst's consumed-veil row is named "DMG Bonus" on the talent table.
	// It stays in the ordinary DMG Bonus bucket. Phantasm's 8% per stack does not:
	// that one is a Base DMG Multiplier, applied on the attack itself.
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase("nefer-burst-veil", -1),
		Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			if atk.Info.ActorIndex != c.Index() || c.burstBonus == 0 {
				return nil
			}
			if !strings.HasPrefix(atk.Info.Abil, "Sacred Vow") {
				return nil
			}
			m := make([]float64, attributes.EndStatType)
			m[attributes.DmgP] = c.burstBonus
			return m
		},
	})
}

func (c *char) markCores() {
	for _, g := range c.Core.Combat.Gadgets() {
		if d, ok := g.(interface{ MakeDeceit() }); ok {
			d.MakeDeceit()
		}
	}
}

func (c *char) absorbSeeds() {
	if c.Base.Ascension < 1 || !c.gleam {
		return
	}
	n := 0
	for _, g := range c.Core.Combat.Gadgets() {
		d, ok := g.(interface {
			IsDeceit() bool
			Absorb()
		})
		if !ok || !d.IsDeceit() {
			continue
		}
		d.Absorb()
		n++
	}
	if n > 0 {
		c.addVeils(n)
	}
}

func (c *char) veilCount() int {
	c.pruneVeils()
	return len(c.veils)
}

func (c *char) pruneVeils() {
	if len(c.veils) == 0 {
		return
	}
	kept := c.veils[:0]
	for _, until := range c.veils {
		if until > c.Core.F {
			kept = append(kept, until)
		}
	}
	c.veils = kept
}

func (c *char) addVeils(n int) {
	maxN := 3
	em := 100.0
	dur := veilDur
	if c.Base.Cons >= 2 {
		maxN = 5
		em = 200
		dur = veilDurC2
	}
	for range n {
		c.pruneVeils()
		if len(c.veils) < maxN {
			c.veils = append(c.veils, c.Core.F+dur)
			if len(c.veils) == maxN {
				c.grantVeilEM(em)
			}
			continue
		}
		c.veils[len(c.veils)-1] = c.Core.F + dur
		c.grantVeilEM(em)
	}
}

func (c *char) grantVeilEM(em float64) {
	m := make([]float64, attributes.EndStatType)
	m[attributes.EM] = em
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("nefer-veil-em", veilEMDur),
		AffectedStat: attributes.EM,
		Amount: func() []float64 {
			return m
		},
	})
}
