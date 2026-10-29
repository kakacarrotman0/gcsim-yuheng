package razor

import (
	"errors"

	"github.com/genshinsim/gcsim/internal/frames"
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/keys"
)

// Lightning Fang disables Charged Attacks. The Wolf Within strikes on Normal Attacks only.
var errLightningFangCharge = errors.New("Lightning Fang disables Charged Attacks")

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.StatusIsActive(burstBuffKey) {
		return action.Info{}, errLightningFangCharge
	}
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

func (c *char) NextQueueItemIsValid(k keys.Char, a action.Action, p map[string]int) error {
	if a == action.ActionCharge && c.StatusIsActive(burstBuffKey) {
		return errLightningFangCharge
	}
	return c.Character.NextQueueItemIsValid(k, a, p)
}
