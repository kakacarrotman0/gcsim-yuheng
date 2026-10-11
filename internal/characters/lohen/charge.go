package lohen

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// Two ExtraAttack hits. Spacing is unverified. The default CA→N1 cancel
// stays at 36. The optional cadence profile changes that cancel only so
// the combined N1C loop matches experimentalN1C. The two Joy modifiers are
// different names, so both hits add Joy even inside 0.1s.
// The ICD tag is normal attack, not extra attack: both hits share 洛恩战技
// with the normal string.
var (
	chargeFrames        []int
	chargeFramesCadence []int
	chargeHitmarks      = []int{18, 28}
)

func init() {
	chargeFrames = frames.InitAbilSlice(46)
	chargeFrames[action.ActionAttack] = 36
	chargeFrames[action.ActionSkill] = 36
	chargeFrames[action.ActionBurst] = 36
	chargeFrames[action.ActionDash] = 28
	chargeFrames[action.ActionSwap] = 34
	buildCadenceFrames()
}

// buildCadenceFrames keeps every placeholder hitmark and shortens only the
// N1→CA and CA→N1 recovery. The footage fixes their sum, not each one.
func buildCadenceFrames() {
	lastCA := chargeHitmarks[len(chargeHitmarks)-1]
	remain := experimentalN1C - attackHitmarks[0][0] - lastCA
	if remain < 0 || remain%2 != 0 {
		panic("experimental N1C sum does not fit the placeholder hitmarks")
	}
	tail := remain / 2
	experimentalN1ToCA = attackHitmarks[0][0] + tail
	experimentalCAToN1 = lastCA + tail

	attackFramesCadence = make([][]int, len(attackFrames))
	for i := range attackFrames {
		attackFramesCadence[i] = append([]int(nil), attackFrames[i]...)
	}
	attackFramesCadence[0][action.ActionCharge] = experimentalN1ToCA

	chargeFramesCadence = append([]int(nil), chargeFrames...)
	chargeFramesCadence[action.ActionAttack] = experimentalCAToN1
}

func (c *char) ChargeAttack(p map[string]int) (action.Info, error) {
	ap := combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 0.8, 4)
	for i, delay := range chargeHitmarks {
		hit := i
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
			// ExtraAttackA and ExtraAttackB are separate Joy modifiers.
			joyKey := joyCAAKey
			if hit == 1 {
				joyKey = joyCABKey
			}
			c.Core.QueueAttack(ai, ap, 0, 0, c.hitCB(joy, joyKey, true))
		}, delay)
	}
	// CA→N1 uses the experimental split only when this charged attack starts
	// during Masterstroke. The footage does not show the physical string.
	slice := chargeFrames
	if c.n1cCadence && c.masterActive() {
		slice = chargeFramesCadence
	}
	return action.Info{
		Frames:          frames.NewAbilFunc(slice),
		AnimationLength: slice[action.InvalidAction],
		CanQueueAfter:   chargeHitmarks[0],
		State:           action.ChargeAttackState,
	}, nil
}
