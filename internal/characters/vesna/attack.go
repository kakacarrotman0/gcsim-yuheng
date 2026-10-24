package vesna

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// Hitmarks are an unverified sword spacing. They are not a frame sheet.
// KQM's Vesna page has no findings. N3 is two hits and still summons one pinion.
var (
	attackFrames   [][]int
	attackHitmarks = [][]int{{8}, {12}, {10, 20}, {16}, {22}, {28}}
)

const normalHitNum = 6

func init() {
	attackFrames = make([][]int, normalHitNum)
	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 20)
	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 24)
	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][1], 36)
	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 32)
	attackFrames[4] = frames.InitNormalCancelSlice(attackHitmarks[4][0], 40)
	attackFrames[5] = frames.InitNormalCancelSlice(attackHitmarks[5][0], 48)
}

func (c *char) normalMults(counter int) []float64 {
	lvl := c.TalentLvlAttack()
	switch counter {
	case 0:
		return []float64{attack1[lvl]}
	case 1:
		return []float64{attack2[lvl]}
	case 2:
		return []float64{attack3[lvl], attack3[lvl]}
	case 3:
		return []float64{attack4[lvl]}
	case 4:
		return []float64{attack5[lvl]}
	default:
		return []float64{attack6[lvl]}
	}
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	if c.transposeReady() {
		c.transpose()
		return action.Info{
			Frames:          framesAbil(),
			AnimationLength: skillFrames[action.InvalidAction],
			CanQueueAfter:   skillFrames[action.ActionDash],
			State:           action.NormalAttackState,
		}, nil
	}
	// One pinion per normal action, not per hit. Essence is added when the action starts.
	if c.armed() {
		c.summonPinion()
	}
	counter := c.NormalCounter
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2.2)
	for i := range attackHitmarks[counter] {
		hit := i
		c.Core.Tasks.Add(func() {
			// Element is chosen when the hit lands. Armed converts the strike
			// to Anemo and blocks weapon infusion. A swap that already ended
			// the mode leaves the hit Physical.
			ai := info.AttackInfo{
				ActorIndex: c.Index(),
				Abil:       fmt.Sprintf("Normal %v", counter),
				AttackTag:  attacks.AttackTagNormal,
				ICDTag:     attacks.ICDTagNormalAttack,
				ICDGroup:   attacks.ICDGroupDefault,
				StrikeType: attacks.StrikeTypeSlash,
				Element:    attributes.Physical,
				Durability: 25,
				Mult:       c.normalMults(counter)[hit],
			}
			if c.armed() {
				ai.Element = attributes.Anemo
				ai.IgnoreInfusion = true
				ai.Abil = fmt.Sprintf("Armed Normal %v", counter)
			}
			c.Core.QueueAttack(ai, ap, 0, 0)
		}, attackHitmarks[counter][hit])
	}
	defer c.AdvanceNormalIndex()
	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[counter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[counter][0],
		State:           action.NormalAttackState,
	}, nil
}
