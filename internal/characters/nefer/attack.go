package nefer

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var attackFrames [][]int

var attackHitmarks = []int{15, 18, 28, 40}

const normalHitNum = 4

func init() {
	attackFrames = make([][]int, normalHitNum)
	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 32)
	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 36)
	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 58)
	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3], 68)
}

func (c *char) Attack(_ map[string]int) (action.Info, error) {
	idx := c.NormalCounter
	ratios := [][]float64{attack1, attack2, attack3, attack4}
	hits := 1
	if idx == 2 {
		hits = 2
	}
	for i := range hits {
		delay := attackHitmarks[idx]
		if i > 0 {
			delay += n3Split
		}
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       fmt.Sprintf("Normal %v", idx),
			AttackTag:  attacks.AttackTagNormal,
			ICDTag:     attacks.ICDTagNormalAttack,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Dendro,
			Durability: 25,
			Mult:       ratios[idx][c.TalentLvlAttack()],
		}
		c.queueCircle(ai, delay)
	}
	defer c.AdvanceNormalIndex()
	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[idx][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[idx],
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) queueCircle(ai info.AttackInfo, delay int, cb ...info.AttackCBFunc) {
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.PrimaryTarget(), nil, hitRadius)
	c.Core.QueueAttack(ai, ap, delay, delay, cb...)
}
