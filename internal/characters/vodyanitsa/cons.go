package vodyanitsa

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c1ATKRatio    = 0.008
	c1Dur         = 5 * 60
	c1Key         = "vodyanitsa-c1"
	c2HydroCD     = 0.50
	c2StellarCD   = 0.60
	c2Dur         = 5 * 60
	c2Key         = "vodyanitsa-c2"
	c4HPThreshold = 0.40
	c4ExtraHeal   = 0.50
	c4HPRatio     = 0.20
	c4Dur         = 6 * 60
	c4MaxStacks   = 3
	c4Key         = "vodyanitsa-c4"
	c6DmgP        = 0.60
	c6Elevation   = 0.25
	c6Key         = "vodyanitsa-c6"
)

func (c *char) c1Init() {
	if c.Base.Cons < 1 {
		return
	}
	for _, char := range c.Core.Player.Chars() {
		this := char
		this.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(c1Key, -1),
			AffectedStat: attributes.ATK,
			Amount: func() []float64 {
				if !this.StatusIsActive(c1Key + "-active") {
					return nil
				}
				out := make([]float64, attributes.EndStatType)
				out[attributes.ATK] = c1ATKRatio * c.MaxHP()
				return out
			},
		})
	}
}

func (c *char) c1OnHeal() {
	if c.Base.Cons < 1 {
		return
	}
	for _, char := range c.Core.Player.Chars() {
		char.AddStatus(c1Key+"-active", c1Dur, false)
	}
}

func (c *char) c2Init() {
	if c.Base.Cons < 2 {
		return
	}
	for _, char := range c.Core.Player.Chars() {
		this := char
		this.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase(c2Key, -1),
			Amount: func(ae *info.AttackEvent, _ info.Target) []float64 {
				if !c.StatusIsActive(c2Key) || c.alterActive() {
					return nil
				}
				if ae.Info.ActorIndex != this.Index() {
					return nil
				}
				if c.Base.Cons < 6 && c.Core.Player.Active() != this.Index() {
					return nil
				}
				if ae.Info.Element != attributes.Hydro && ae.Info.Element != attributes.Cryo {
					return nil
				}
				if ae.Info.AttackTag.IsStellarReact() {
					return nil
				}
				out := make([]float64, attributes.EndStatType)
				out[attributes.CD] = c2HydroCD
				return out
			},
		})
	}

	c.Core.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		if !c.StatusIsActive(c2Key) || !c.alterActive() {
			return
		}
		atk := args[1].(*info.AttackEvent)
		if !atk.Info.AttackTag.IsStellarReact() {
			return
		}
		if c.Base.Cons < 6 && c.Core.Player.Active() != atk.Info.ActorIndex {
			return
		}
		atk.Snapshot.Stats[attributes.CD] += c2StellarCD
	}, "vodyanitsa-c2-stellar")
}

func (c *char) c2OnHorn(a info.AttackCB) {
	if c.Base.Cons < 2 {
		return
	}
	if a.Target.Type() != info.TargettableEnemy {
		return
	}
	c.AddStatus(c2Key, c2Dur, false)
}

func (c *char) c4Init() {
	if c.Base.Cons < 4 {
		return
	}
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(c4Key, -1),
		AffectedStat: attributes.HPP,
		Amount: func() []float64 {
			n := c.c4Count()
			if n == 0 {
				return nil
			}
			out := make([]float64, attributes.EndStatType)
			out[attributes.HPP] = c4HPRatio * float64(n)
			return out
		},
	})
}

func (c *char) addC4Stack() {
	if c.c4Count() >= c4MaxStacks {
		return
	}
	c.c4Expiry = append(c.c4Expiry, c.Core.F+c4Dur)
}

func (c *char) c4Count() int {
	kept := make([]int, 0, len(c.c4Expiry))
	for _, exp := range c.c4Expiry {
		if exp > c.Core.F {
			kept = append(kept, exp)
		}
	}
	c.c4Expiry = kept
	return len(kept)
}

func (c *char) c6Init() {
	if c.Base.Cons < 6 {
		return
	}
	for _, char := range c.Core.Player.Chars() {
		this := char
		this.AddAttackMod(character.AttackMod{
			Base: modifier.NewBase(c6Key, -1),
			Amount: func(ae *info.AttackEvent, _ info.Target) []float64 {
				if !c.songActive() || ae.Info.ActorIndex != this.Index() {
					return nil
				}
				if ae.Info.AttackTag.IsStellarReact() {
					return nil
				}
				if ae.Info.Element != attributes.Hydro && ae.Info.Element != attributes.Cryo {
					return nil
				}
				out := make([]float64, attributes.EndStatType)
				out[attributes.DmgP] = c6DmgP
				return out
			},
		})
	}

	c.Core.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		if !c.songActive() {
			return
		}
		atk := args[1].(*info.AttackEvent)
		if !atk.Info.AttackTag.IsStellarReact() {
			return
		}
		atk.Info.Elevation += c6Elevation
		if atk.Info.Element == attributes.Hydro || atk.Info.Element == attributes.Cryo {
			atk.Snapshot.Stats[attributes.DmgP] += c6DmgP
		}
	}, "vodyanitsa-c6-stellar")
}
