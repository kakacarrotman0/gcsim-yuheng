package candace

import (
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

// A1, Aegis of Crossed Arrows, is skill[perfect=1] after ascension 1.
// It releases the charged hold 55 frames earlier. The sim does not detect an incoming hit.
// Before ascension 1 the parameter is ignored.

const a4Key = "candace-a4"

// Characters affected by the Prayer of the Crimson Crown caused by Sacred Rite: Wagtail's Tide will deal 0.5% increased DMG
// to opponents for every 1,000 points of Candace's Max HP when they deal Elemental DMG with their Normal Attacks.
func (c *char) a4(char *character.CharWrapper) {
	if c.Base.Ascension < 4 {
		return
	}
	m := make([]float64, attributes.EndStatType)
	char.AddAttackMod(character.AttackMod{
		Base: modifier.NewBase(a4Key, -1),
		Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			if !c.StatusIsActive(burstKey) {
				return nil
			}
			if atk.Info.AttackTag != attacks.AttackTagNormal {
				return nil
			}
			if atk.Info.Element == attributes.Physical || atk.Info.Element == attributes.NoElement {
				return nil
			}
			m[attributes.DmgP] = 0.005 * c.MaxHP() / 1000
			return m
		},
	})
}
