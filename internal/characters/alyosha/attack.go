package alyosha

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const normalHitNum = 4

var (
	attackFrames  [][]int
	chargeFrames  []int
	attackHitmark = [][]int{{12}, {16, 22}, {28}, {34}}
)

func init() {
	// Hitmarks are approximate. No Alyosha frame sheet was available.
	attackFrames = make([][]int, normalHitNum)
	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmark[0][0], 32)
	attackFrames[0][action.ActionAttack] = 20
	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmark[1][1], 48)
	attackFrames[1][action.ActionAttack] = 36
	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmark[2][0], 50)
	attackFrames[2][action.ActionAttack] = 40
	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmark[3][0], 62)
	attackFrames[3][action.ActionAttack] = 55

	chargeFrames = frames.InitAbilSlice(52)
	chargeFrames[action.ActionAttack] = 40
	chargeFrames[action.ActionSkill] = 40
	chargeFrames[action.ActionBurst] = 40
	chargeFrames[action.ActionDash] = 36
	chargeFrames[action.ActionJump] = 36
	chargeFrames[action.ActionSwap] = 38
}

func (c *char) Attack(_ map[string]int) (action.Info, error) {
	counter := c.NormalCounter
	for i, mult := range attack[counter] {
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       fmt.Sprintf("Normal %v", counter),
			AttackTag:  attacks.AttackTagNormal,
			ICDTag:     attacks.ICDTagNormalAttack,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeSpear,
			Element:    attributes.Physical,
			Durability: 25,
			Mult:       mult[c.TalentLvlAttack()],
		}
		hitmark := attackHitmark[counter][i]
		var cb info.AttackCBFunc
		if counter == normalHitNum-1 && i == len(attack[counter])-1 {
			cb = func(a info.AttackCB) {
				c.markHit(a.Target, true, true)
			}
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2.2),
			hitmark,
			hitmark,
			cb,
		)
	}
	defer c.AdvanceNormalIndex()
	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[counter][action.InvalidAction],
		CanQueueAfter:   attackHitmark[counter][len(attackHitmark[counter])-1],
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) ChargeAttack(_ map[string]int) (action.Info, error) {
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Charge",
		AttackTag:  attacks.AttackTagExtra,
		ICDTag:     attacks.ICDTagExtraAttack,
		ICDGroup:   attacks.ICDGroupPoleExtraAttack,
		StrikeType: attacks.StrikeTypeSpear,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       charge[c.TalentLvlAttack()],
	}
	c.Core.QueueAttack(ai, combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 2.5, 6), 18, 18)
	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   18,
		State:           action.ChargeAttackState,
	}, nil
}
