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
// The default N1→CA cancel stays at the 22-frame recovery.
// N3's three hits are 8 frames apart so they sit outside the 0.1s unique Joy modifier.
//
// NA and CA share BWiki group 洛恩战技: ICD tag normal attack, 5s reset,
// sequence 1,0 repeated, and at most 6 applications before the reset.
// The same group covers the physical string and Masterstroke.
var (
	attackFrames        [][]int
	attackFramesCadence [][]int
	attackHitmarks      = [][]int{{12}, {9}, {12, 20, 28}, {16}, {22, 34}}
)

// experimentalN1C is the optional n1c_cadence=1 loop, N1→CA plus CA→N1.
// 46 is the median repeated attack-trigger interval in BowTae's 60 FPS
// footage (mode 46, typical range 46–48 after the opening gap). That
// interval is one full cycle between presses at the same point in the
// animation. It still contains hitlag and whatever input delay is common
// to both presses. It is not a measured N1 cancel or a measured CA cancel.
const experimentalN1C = 46

// Filled in once both placeholder tables exist. The split is the frames
// left after the placeholder hitmarks, divided evenly. It is not measured.
var (
	experimentalN1ToCA int
	experimentalCAToN1 int
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

func (c *char) normalMults(counter int) []float64 {
	lvl := c.TalentLvlAttack()
	if c.masterActive() {
		lvl = c.skillLvl()
		switch counter {
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
	switch counter {
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
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2.2)
	for i := range attackHitmarks[counter] {
		hit := i
		c.QueueCharTask(func() {
			// Element is chosen when the hit lands. Masterstroke converts the
			// strike to Cryo and blocks weapon infusion. A swap that already
			// ended the mode leaves the hit Physical.
			mult := c.normalMults(counter)[hit]
			ai := info.AttackInfo{
				ActorIndex: c.Index(),
				Abil:       fmt.Sprintf("Normal %v", counter),
				AttackTag:  attacks.AttackTagNormal,
				ICDTag:     attacks.ICDTagNormalAttack,
				ICDGroup:   attacks.ICDGroupLohenSkill,
				StrikeType: attacks.StrikeTypeSpear,
				Element:    attributes.Physical,
				Durability: 25,
				Mult:       mult,
			}
			joy := 0.0
			if c.masterActive() {
				ai.Element = attributes.Cryo
				ai.IgnoreInfusion = true
				ai.Abil = fmt.Sprintf("Masterstroke Normal %v", counter)
				joy = joyOnNormal[c.skillLvl()]
			}
			// The fifth normal's second hit is NormalAttack_06.
			if counter == 4 && hit == 1 {
				ai.Abil += " B"
			}
			c.Core.QueueAttack(ai, ap, 0, 0, c.hitCB(joy, joyNAKey, true))
		}, attackHitmarks[counter][hit])
	}
	defer c.AdvanceNormalIndex()
	// The 46-frame loop is from Masterstroke footage. Outside that mode the
	// placeholder cancels stay in place. Hitmarks in both slices are approximate.
	slice := attackFrames
	if c.n1cCadence && c.masterActive() {
		slice = attackFramesCadence
	}
	return action.Info{
		Frames:          frames.NewAttackFunc(c.Character, slice),
		AnimationLength: slice[counter][action.InvalidAction],
		CanQueueAfter:   attackHitmarks[counter][0],
		State:           action.NormalAttackState,
	}, nil
}

func (c *char) hitCB(joy float64, joyKey string, canEvilsbane bool) info.AttackCBFunc {
	return func(a info.AttackCB) {
		if a.Target.Type() != info.TargettableEnemy {
			return
		}
		if joy > 0 {
			c.tryJoy(joyKey, joy)
		}
		if canEvilsbane {
			c.tryEvilsbane()
		}
		if a.AttackEvent != nil && a.AttackEvent.Info.Element == attributes.Cryo {
			c.tryParticle()
		}
	}
}
