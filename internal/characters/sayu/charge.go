package sayu

import (
	"github.com/genshinsim/gcsim/internal/frames"
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

// Fuuin Dash is the skill. The attacks table prints the spin percent "x 9";
// that product is not a separate hit count. One slash lands per interval.
func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	return c.SpinChargeAttack(&c.caState, tmpl.SpinChargeSpec{
		Windup:     tmpl.ClaymoreSpinWindup,
		SpinHit:    tmpl.ClaymoreSpinHit,
		FinalHit:   tmpl.ClaymoreSpinFinalHit,
		Transition: tmpl.ClaymoreSpinTransition,
		Duration:   tmpl.ClaymoreSpinDuration,
		Adjust:     frames.AtkSpdAdjust,
		Hit: func(final bool) {
			lvl := c.TalentLvlAttack()
			c.ClaymoreSpinAttackRadius(final, charge[lvl], chargeFinal[lvl], nil, 2.8, 3.3)
		},
	}, p)
}
