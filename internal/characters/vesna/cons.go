package vesna

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) c1Init() {
	if c.Base.Cons < 1 {
		return
	}
	// While Armed, +20% Stellar Swirl reaction DMG. Direct reactions ignore DMG%,
	// so this is a reaction bonus on Vesna's stellar tags only.
	c.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase("vesna-c1", -1),
		Amount: func(ai info.AttackInfo) float64 {
			if !c.armed() {
				return 0
			}
			switch ai.AttackTag {
			case attacks.AttackTagDirectStellarSwirl, attacks.AttackTagReactionStellarSwirl:
				return c1React
			default:
				return 0
			}
		},
	})
}

func (c *char) c2Init() {
	if c.Base.Cons < 2 || c.Base.Ascension < 1 {
		return
	}
	c.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("vesna-c2", -1),
		AffectedStat: attributes.ATKP,
		Amount: func() []float64 {
			m := make([]float64, attributes.EndStatType)
			if c.discCount() == discMax {
				m[attributes.ATKP] = c2ATK
			}
			return m
		},
	})
}

const (
	transposeSlash      = 16
	transposeBladeDelay = 46
)

func (c *char) transpose() {
	// One use. The window also ends on swap or after 5s.
	armed := c.armed()
	c.endTranspose()
	if armed {
		c.summonPinion()
	}
	// The constellation block has no gauge row. This slash stays on the standard
	// group rather than being folded into Vesna_MagicSword without a tag.
	c.artHit("Windborne Sword: Transpose", transposeTalent, transposeSlash, attacks.StrikeTypeSlash, attacks.ICDGroupDefault)
	c.queueSpirit("Transpose Spirit Blade", transposeBlade, transposeBladeDelay, false, bladeRadius)
}
