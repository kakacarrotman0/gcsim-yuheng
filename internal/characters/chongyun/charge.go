package chongyun

import (
	"github.com/genshinsim/gcsim/internal/frames"
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

// Chonghua's field infuses claymore Charged Attacks through the existing weapon infusion.
// C1 stays on the fourth Normal Attack. A1's 8% is a Normal ATK SPD stat on
// "chongyun-field". Normals keep that stat. This clock subtracts it.
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
			c.ClaymoreSpinAttack(final, charge[lvl], chargeFinal[lvl], nil)
		},
	}, p)
}

func (c *char) chargeAdjust(f int, spd float64) int {
	if c.StatModIsActive("chongyun-field") {
		spd -= 0.08
		if spd < 0 {
			spd = 0
		}
	}
	return frames.AtkSpdAdjust(f, spd)
}
