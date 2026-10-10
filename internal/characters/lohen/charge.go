package lohen

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// Two ExtraAttack hits. Spacing is unverified. The two Joy modifiers are
// different names, so both hits add Joy even inside 0.1s.
// The ICD tag is normal attack, not extra attack: both hits share 洛恩战技
// with the normal string.
var (
	chargeFrames   []int
	chargeHitmarks = []int{18, 28}
)

func init() {
	chargeFrames = frames.InitAbilSlice(46)
	chargeFrames[action.ActionAttack] = 36
	chargeFrames[action.ActionSkill] = 36
	chargeFrames[action.ActionBurst] = 36
	chargeFrames[action.ActionDash] = 28
	chargeFrames[action.ActionSwap] = 34
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	ap := combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 0.8, 4)
	for _, delay := range chargeHitmarks {
		c.QueueCharTask(func() {
			lvl := c.TalentLvlAttack()
			mult := attack7[lvl]
			abil := "Charged Attack"
			ele := attributes.Physical
			ignore := false
			joy := 0.0
			if c.masterActive() {
				lvl = c.skillLvl()
				mult = enhanced7[lvl]
				abil = "Masterstroke Charged Attack"
				ele = attributes.Cryo
				ignore = true
				joy = joyOnCharge[lvl]
			}
			ai := info.AttackInfo{
				ActorIndex:     c.Index(),
				Abil:           abil,
				AttackTag:      attacks.AttackTagExtra,
				ICDTag:         attacks.ICDTagNormalAttack,
				ICDGroup:       attacks.ICDGroupLohenSkill,
				StrikeType:     attacks.StrikeTypeSpear,
				Element:        ele,
				Durability:     25,
				Mult:           mult,
				IgnoreInfusion: ignore,
			}
			c.Core.QueueAttack(ai, ap, 0, 0, c.hitCB(joy, true))
		}, delay)
	}
	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmarks[0],
		State:           action.ChargeAttackState,
	}, nil
}
