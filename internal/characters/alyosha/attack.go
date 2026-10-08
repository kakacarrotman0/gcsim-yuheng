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
	attackFrames [][]int
	chargeFrames []int
	// Sheet hitmark medians. N3-2 is 15 frames after N3-1 on every trial.
	attackHitmark         = [][]int{{13}, {14}, {17, 17 + 15}, {17}}
	attackHitlagHaltFrame = [][]float64{{0.06}, {0.06}, {0, 0}, {0}}
	attackDefHalt         = [][]bool{{true}, {true}, {false, false}, {false}}
	attackHitboxes        = [][][]float64{{{2}}, {{2, 4}}, {{1.8}, {2, 4}}, {{2.4, 6}}}
	attackOffsets         = [][][]float64{{{0, 0.6}}, {{0.2, 0}}, {{0, 0.8}, {0, -0.5}}, {{0, 0.5}}}
)

func init() {
	// Cancels are sheet medians. Uncounted skill/burst/dash/jump/swap cells stay at the hitmark.
	attackFrames = make([][]int, normalHitNum)
	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmark[0][0], 34) // walk 34/34/33
	attackFrames[0][action.ActionAttack] = 18                               // 18/18/17
	attackFrames[0][action.ActionCharge] = 24

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmark[1][0], 38)
	attackFrames[1][action.ActionAttack] = 23 // midpoint of 24 and 21; trial 3 excluded on the sheet
	attackFrames[1][action.ActionCharge] = 26 // 25/26/26

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmark[2][1], 65) // walk 66/65/65
	attackFrames[2][action.ActionAttack] = 53                               // 53/54/53
	attackFrames[2][action.ActionCharge] = 54

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmark[3][0], 62)
	attackFrames[3][action.ActionAttack] = 56 // 57/56/56

	// CA hitmark median is 22 (trials 23/22/22). Dash and jump are "anytime".
	chargeFrames = frames.InitAbilSlice(65) // walk
	chargeFrames[action.ActionAttack] = 63  // 63/63/64
	chargeFrames[action.ActionSkill] = 62
	chargeFrames[action.ActionBurst] = 63
	chargeFrames[action.ActionDash] = 22
	chargeFrames[action.ActionJump] = 22
	chargeFrames[action.ActionSwap] = 61 // 60/61/61
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
			Element:    attributes.Physical,
			Durability: 25,
			Mult:       mult[c.TalentLvlAttack()],
		}
		if attackHitlagHaltFrame[counter][i] > 0 || attackDefHalt[counter][i] {
			ai.HitlagFactor = 0.01
			ai.HitlagHaltFrames = attackHitlagHaltFrame[counter][i] * 60
			ai.CanBeDefenseHalted = attackDefHalt[counter][i]
		}
		offset := info.Point{X: attackOffsets[counter][i][0], Y: attackOffsets[counter][i][1]}
		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), offset, attackHitboxes[counter][i][0])
		ai.StrikeType = attacks.StrikeTypeSlash
		if len(attackHitboxes[counter][i]) == 2 {
			ai.StrikeType = attacks.StrikeTypeSpear
			ap = combat.NewBoxHitOnTarget(
				c.Core.Combat.Player(),
				offset,
				attackHitboxes[counter][i][0],
				attackHitboxes[counter][i][1],
			)
		}
		var cb info.AttackCBFunc
		if counter == normalHitNum-1 && i == len(attack[counter])-1 {
			cb = func(a info.AttackCB) {
				c.markHit(a.Target, true, true)
			}
		}
		c.Core.QueueAttack(ai, ap, attackHitmark[counter][i], attackHitmark[counter][i], cb)
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
		ActorIndex:         c.Index(),
		Abil:               "Charge",
		AttackTag:          attacks.AttackTagExtra,
		ICDTag:             attacks.ICDTagExtraAttack,
		ICDGroup:           attacks.ICDGroupPoleExtraAttack,
		StrikeType:         attacks.StrikeTypeSpear,
		Element:            attributes.Physical,
		Durability:         25,
		Mult:               charge[c.TalentLvlAttack()],
		HitlagHaltFrames:   0,
		HitlagFactor:       0.01,
		CanBeDefenseHalted: true,
		IsDeployable:       true,
	}
	// Bullet radius 0.8 on the primary target.
	ap := combat.NewCircleHit(c.Core.Combat.Player(), c.Core.Combat.PrimaryTarget(), nil, 0.8)
	c.Core.QueueAttack(ai, ap, 0, 22)
	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   22,
		State:           action.ChargeAttackState,
	}, nil
}
