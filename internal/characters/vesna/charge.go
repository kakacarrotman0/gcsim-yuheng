package vesna

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// One charged hit. Spacing is unverified. Sword charged stamina is already 20
// on the weapon-class default, which matches the talent row.
var (
	chargeFrames   []int
	chargeHitmarks = []int{18}
)

func init() {
	chargeFrames = frames.InitAbilSlice(40)
	chargeFrames[action.ActionAttack] = 28
	chargeFrames[action.ActionSkill] = 28
	chargeFrames[action.ActionBurst] = 28
	chargeFrames[action.ActionDash] = 24
	chargeFrames[action.ActionSwap] = 30
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	if c.armed() {
		c.summonPinion()
	}
	ap := combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 0.8, 3)
	for _, delay := range chargeHitmarks {
		c.Core.Tasks.Add(func() {
			ai := info.AttackInfo{
				ActorIndex: c.Index(),
				Abil:       "Charged Attack",
				AttackTag:  attacks.AttackTagExtra,
				ICDTag:     attacks.ICDTagNormalAttack,
				ICDGroup:   attacks.ICDGroupDefault,
				StrikeType: attacks.StrikeTypeSlash,
				Element:    attributes.Physical,
				Durability: 25,
				Mult:       attack7[c.TalentLvlAttack()],
			}
			if c.armed() {
				ai.Element = attributes.Anemo
				ai.IgnoreInfusion = true
				ai.Abil = "Armed Charged Attack"
			}
			c.Core.QueueAttack(ai, ap, 0, 0)
		}, delay)
	}
	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmarks[0],
		State:           action.ChargeAttackState,
	}, nil
}
