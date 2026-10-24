package sandrone

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const a4Key = "sandrone-a4"

// A Lady's Code of Conduct. 8 EM per 100 ATK, capped at 160.
// TotalAtk does not read EM, so the modifier does not recurse.
func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(a4Key, -1),
		AffectedStat: attributes.EM,
		Amount: func() []float64 {
			em := c.TotalAtk() / 100 * a4Per100
			if em > a4Cap {
				em = a4Cap
			}
			m[attributes.EM] = em
			return m
		},
	})
}
