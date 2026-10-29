package diluc

import (
	"github.com/genshinsim/gcsim/internal/frames"
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

// Relentless: Charged Attack stamina cost decreased by 50%, duration increased by 3s.
func (c *char) a1() {
	c.Core.Player.AddStamPercentMod("diluc-a1", -1, func(a action.Action) (float64, bool) {
		if a == action.ActionCharge {
			return -0.5, false
		}
		return 0, false
	})
}

func (c *char) chargeDuration() int {
	if c.Base.Ascension >= 1 {
		return tmpl.ClaymoreSpinDuration + 3*60
	}
	return tmpl.ClaymoreSpinDuration
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	return c.SpinChargeAttack(&c.caState, tmpl.SpinChargeSpec{
		Windup:     tmpl.ClaymoreSpinWindup,
		SpinHit:    tmpl.ClaymoreSpinHit,
		FinalHit:   tmpl.ClaymoreSpinFinalHit,
		Transition: tmpl.ClaymoreSpinTransition,
		Duration:   c.chargeDuration(),
		Adjust:     frames.AtkSpdAdjust,
		Hit: func(final bool) {
			lvl := c.TalentLvlAttack()
			c.ClaymoreSpinAttack(final, charge[lvl], chargeFinal[lvl], nil)
		},
	}, p)
}
