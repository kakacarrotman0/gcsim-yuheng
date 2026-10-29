package kaveh

import (
	"github.com/genshinsim/gcsim/internal/frames"
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

// Painted Dome already infuses charged attacks and enlarges their hitbox.
// C2's 15% is Normal Attack SPD. The stat stays on normals. This clock subtracts it.
// Ordinary hits do not rupture Dendro Cores. Constellation 6 already listens for charged attacks.
func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	return c.SpinChargeAttack(&c.caState, tmpl.SpinChargeSpec{
		Windup:     tmpl.ClaymoreSpinWindup,
		SpinHit:    tmpl.ClaymoreSpinHit,
		FinalHit:   tmpl.ClaymoreSpinFinalHit,
		Transition: tmpl.ClaymoreSpinTransition,
		Duration:   tmpl.ClaymoreSpinDuration,
		Adjust:     c.chargeAdjust,
		Hit: func(final bool) {
			lvl := c.TalentLvlAttack()
			cyclicR, finalR := 2.6, 3.3
			if c.StatModIsActive(burstKey) {
				cyclicR, finalR = 3.1, 3.6
			}
			c.ClaymoreSpinAttackRadius(final, charge[lvl], chargeFinal[lvl], nil, cyclicR, finalR)
		},
	}, p)
}

func (c *char) chargeAdjust(f int, spd float64) int {
	if c.StatModIsActive(c2Key) {
		spd -= 0.15
		if spd < 0 {
			spd = 0
		}
	}
	return frames.AtkSpdAdjust(f, spd)
}
