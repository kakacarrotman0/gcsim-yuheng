package lohen

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// Hitmarks are an unverified polearm spacing. They are not a frame sheet.
// N3's three hits are 8 frames apart so they sit outside the 0.1s unique Joy modifier.
var (
	attackFrames   [][]int
	attackHitmarks = [][]int{{12}, {9}, {12, 20, 28}, {16}, {22, 34}}
)

const normalHitNum = 5

func init() {
	attackFrames = make([][]int, normalHitNum)
	attackFrames[0] = frames.InitNormalCancelSlice(attackHitmarks[0][0], 22)
	attackFrames[1] = frames.InitNormalCancelSlice(attackHitmarks[1][0], 20)
	attackFrames[2] = frames.InitNormalCancelSlice(attackHitmarks[2][2], 40)
	attackFrames[3] = frames.InitNormalCancelSlice(attackHitmarks[3][0], 32)
	attackFrames[4] = frames.InitNormalCancelSlice(attackHitmarks[4][1], 58)
}

func (c *char) normalMults() []float64 {
	lvl := c.TalentLvlAttack()
	if c.masterActive() {
		lvl = c.skillLvl()
		switch c.NormalCounter {
		case 0:
			return []float64{enhanced1[lvl]}
		case 1:
			return []float64{enhanced2[lvl]}
		case 2:
			return []float64{enhanced3[lvl], enhanced3[lvl], enhanced3[lvl]}
		case 3:
			return []float64{enhanced4[lvl]}
		default:
			return []float64{enhanced5[lvl], enhanced6[lvl]}
		}
	}
	switch c.NormalCounter {
	case 0:
		return []float64{attack1[lvl]}
	case 1:
		return []float64{attack2[lvl]}
	case 2:
		return []float64{attack3[lvl], attack3[lvl], attack3[lvl]}
	case 3:
		return []float64{attack4[lvl]}
	default:
		return []float64{attack5[lvl], attack6[lvl]}
	}
}

func (c *char) Attack(p map[string]int) (action.Info, error) {
	counter := c.NormalCounter
	mults := c.normalMults()
	joy := 0.0
	if c.masterActive() {
		joy = joyOnNormal[c.skillLvl()]
	}
	for i, mult := range mults {
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       fmt.Sprintf("Normal %v", counter),
			AttackTag:  attacks.AttackTagNormal,
			ICDTag:     attacks.ICDTagNormalAttack,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeSpear,
			Element:    attributes.Physical,
			Durability: 25,
			Mult:       mult,
		}
		if c.masterActive() {
			ai.Abil = fmt.Sprintf("Masterstroke Normal %v", counter)
		}
		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2.2)
		delay := attackHitmarks[counter][i]
		c.QueueCharTask(func() {
			c.Core.QueueAttack(ai, ap, 0, 0, c.hitCB(joy, true))
		}, delay)
	}
	defer c.AdvanceNormalIndex()
	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, attackFrames),
		AnimationLength: attackFrames[counter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[counter][0],
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) hitCB(joy float64, canEvilsbane bool) info.AttackCBFunc {
	return func(a info.AttackCB) {
		if a.Target.Type() != info.TargettableEnemy {
			return
		}
		if joy > 0 {
			c.addJoy(joy)
		}
		if canEvilsbane {
			c.tryEvilsbane()
		}
		if a.AttackEvent != nil && a.AttackEvent.Info.Element == attributes.Cryo {
			c.tryParticle()
		}
	}
}
