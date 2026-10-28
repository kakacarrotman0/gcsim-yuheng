package xingqiu

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	a1HealRatio     = 0.06
	a1ShatterICDKey = "xingqiu-a1-shatter-icd"
	a1ShatterICD    = 90 // 1.5s between shattered swords
)

// C1's only effect is one extra Rain Sword. The orbital aura stays a single application.
func (c *char) maxRainSwords() int {
	if c.Base.Cons >= 1 {
		return 4
	}
	return 3
}

func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}
	// A sword shatters when the active character is hit for damage or poise.
	// Raincutter keeps the swords at maximum, so they do not shatter during the burst.
	c.Core.Events.Subscribe(event.OnPlayerHit, func(args ...any) {
		if c.rainSwords <= 0 || !c.StatusIsActive(orbitalKey) || c.StatusIsActive(burstKey) {
			return
		}
		if c.StatusIsActive(a1ShatterICDKey) {
			return
		}
		atk := args[1].(*info.AttackEvent)
		if atk.Info.PoiseDMG <= 0 && atk.Info.Mult <= 0 && atk.Info.FlatDmg <= 0 {
			return
		}
		c.AddStatus(a1ShatterICDKey, a1ShatterICD, false)
		c.rainSwords--
		c.a1Heal(1)
	}, "xingqiu-a1")
}

// Hydropathic heals the active character for 6% of Xingqiu's Max HP per sword.
func (c *char) a1Heal(swords int) {
	if c.Base.Ascension < 1 || swords <= 0 {
		return
	}
	active := c.Core.Player.Active()
	for range swords {
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index(),
			Target:  active,
			Message: "Hydropathic",
			Src:     a1HealRatio * c.MaxHP(),
			Bonus:   c.Stat(attributes.Heal),
		})
	}
}

// Xingqiu gains a 20% Hydro DMG Bonus.
func (c *char) a4() {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.HydroP] = 0.2
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("xingqiu-a4", -1),
		AffectedStat: attributes.HydroP,
		Amount: func() []float64 {
			return m
		},
	})
}
