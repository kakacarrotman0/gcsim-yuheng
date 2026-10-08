package kachina

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	playercharacter "github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const a1Key = "kachina-a1"

func (c *char) initA1() {
	if c.Base.Ascension < 1 {
		return
	}
	c.Core.Events.Subscribe(event.OnNightsoulBurst, func(_ ...any) {
		bonus := make([]float64, attributes.EndStatType)
		bonus[attributes.GeoP] = a1Ratio
		c.AddStatMod(playercharacter.StatMod{
			Base:         modifier.NewBaseWithHitlag(a1Key, a1Duration),
			AffectedStat: attributes.GeoP,
			Amount: func() []float64 {
				return bonus
			},
		})
	}, "kachina-a1")
}
