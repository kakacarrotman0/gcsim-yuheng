package sandrone

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// Hitmarks are an unverified 3-hit claymore spacing. They are not a frame sheet.
// Normals stay Physical. Charged Attack, Skill, and Burst carry Cryo.
var attackHitmarks = []int{16, 14, 28}

const normalHitNum = 3

var attackFrames [][]int

func init() {
	attackFrames = make([][]int, normalHitNum)
	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 32)
	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 28)
	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 54)
}

func (c *char) Attack(_ map[string]int) (action.Info, error) {
	counter := c.NormalCounter
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       fmt.Sprintf("Normal %v", counter),
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNormalAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       attackMults[counter][c.TalentLvlAttack()],
	}
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), info.Point{Y: 0.8}, 2.2)
	c.Core.QueueAttack(ai, ap, attackHitmarks[counter], attackHitmarks[counter])
	defer c.AdvanceNormalIndex()
	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[counter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[counter],
		State:           action.NormalAttackState,
	}, nil
}
