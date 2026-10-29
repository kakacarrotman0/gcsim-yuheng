package beidou

import (
	"github.com/genshinsim/gcsim/internal/frames"
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

// A4's 15% damage and 15% attack speed already apply to this attack.
// The "greatly reduced delay before unleashing Charged Attacks" has no published
// frame value, so the windup stays the shared experimental schedule.
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
			c.ClaymoreSpinAttack(final, charge[lvl], chargeFinal[lvl], nil)
		},
	}, p)
}
