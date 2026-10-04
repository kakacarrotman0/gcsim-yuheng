package vodyanitsa

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

// APPROXIMATE hitmarks. No public frame sheet was available.
// Hitbox radii are from ConfigAbility_Avatar_Vodyanitsa (N1–N3 circle 0.7, N4 box 2x10).
var attackHitmarks = []int{14, 28, 42, 58}

const normalHitNum = 4

func init() {
	attackFrames = make([][]int, normalHitNum)
	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0], 32)
	attackFrames[0][action.ActionAttack] = 24
	attackFrames[0][action.ActionCharge] = 28

	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1], 46)
	attackFrames[1][action.ActionAttack] = 38
	attackFrames[1][action.ActionCharge] = 36

	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2], 60)
	attackFrames[2][action.ActionAttack] = 52
	attackFrames[2][action.ActionCharge] = 48

	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3], 78)
	attackFrames[3][action.ActionAttack] = 70
	attackFrames[3][action.ActionCharge] = 64
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	hit := c.NormalCounter
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       fmt.Sprintf("Normal %v", hit),
		AttackTag:  attacks.AttackTagNormal,
		ICDTag:     attacks.ICDTagNormalAttack,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       attack[hit][c.TalentLvlAttack()],
	}
	if hit == 0 {
		ai.PoiseDMG = 8.1
	} else if hit == 1 {
		ai.PoiseDMG = 7.5
	} else if hit == 2 {
		ai.PoiseDMG = 9.2
	} else {
		ai.PoiseDMG = 12.3
	}

	var ap info.AttackPattern
	if hit == 3 {
		ap = combat.NewBoxHitOnTarget(c.Core.Combat.Player(), info.Point{Y: 1.4}, 2, 10)
	} else {
		ap = combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 0.7)
	}

	c.QueueCharTask(func() {
		c.Core.QueueAttack(ai, ap, 0, 0)
	}, attackHitmarks[hit])

	defer c.AdvanceNormalIndex()

	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[hit][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[hit],
		State:           action.NormalAttackState,
	}, nil
}
