package nefer

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) c4Init() {
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		c.syncC4()
	}, "nefer-c4-swap")
}

func (c *char) syncC4() {
	if c.Base.Cons < 4 {
		return
	}
	onField := c.StatusIsActive(shadowKey) && c.Core.Player.Active() == c.Index()
	c.Core.Player.SetVerdantDewHaste(onField)
	if onField {
		c.c4Shred = true
		c.applyC4(c.StatusDuration(shadowKey) + c4Linger)
		return
	}
	if c.c4Shred {
		c.c4Shred = false
		c.applyC4(c4Linger)
	}
}

func (c *char) applyC4(dur int) {
	for _, t := range c.Core.Combat.Enemies() {
		e, ok := t.(*enemy.Enemy)
		if !ok {
			continue
		}
		e.AddResistMod(info.ResistMod{
			Base:  modifier.NewBaseWithHitlag("nefer-c4", dur),
			Ele:   attributes.Dendro,
			Value: -0.2,
		})
	}
}
