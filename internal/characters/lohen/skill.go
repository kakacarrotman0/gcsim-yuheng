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

var skillFrames []int

func init() {
	// Unverified cancel windows.
	skillFrames = frames.InitAbilSlice(32)
	skillFrames[action.ActionAttack] = 24
	skillFrames[action.ActionCharge] = 24
	skillFrames[action.ActionBurst] = 24
	skillFrames[action.ActionDash] = 22
	skillFrames[action.ActionSwap] = 26
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if p["etched"] > 0 {
		if !c.masterActive() || c.joy < joyCap || c.extraE >= c.etchedMax() {
			return action.Info{}, fmt.Errorf("etched into bone and soul is not available")
		}
		c.etched()
		return action.Info{
			Frames:          frames.NewAbilFunc(skillFrames),
			AnimationLength: skillFrames[action.InvalidAction],
			CanQueueAfter:   skillFrames[action.ActionDash],
			State:           action.SkillState,
		}, nil
	}

	c.enterMasterstroke()
	c.SetCDWithDelay(action.ActionSkill, skillCD, 1)
	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

func (c *char) enterMasterstroke() {
	// CountExtraE_Report zeroes the etched counter when the mode starts.
	// GrandHandler onAdded zeroes Joy and Will and clears the C6 mark on the
	// previous handler's removal.
	c.extraE = 0
	c.joy = 0
	c.will = 0
	c.c6Mark = false
	dur := int(masterDuration[c.skillLvl()] * 60)
	c.masterUntil = c.Core.F + dur
	c.AddStatus(masterKey, dur, true)
	c.highSpirits()
	c.c4OnEnter()
}

func (c *char) etched() {
	lvl := c.skillLvl()
	snapped := c.will
	ratio := c.damageRatio(snapped, willRatio[lvl])
	c.extraE++
	c.joy = 0
	if c.c6Mark {
		c.c6Mark = false
		c.extendMaster(c6Extend)
	}
	if !c.c6KeepWill() {
		c.will = 0
	}
	mult := raid[lvl] * ratio
	// 0.086, 0.2095, 0.4074, 0.5316 of the unverified 60-frame raid state.
	delays := []int{5, 13, 24, 32}
	for i, delay := range delays {
		hit := i
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       fmt.Sprintf("Etched Into Bone and Soul %v", hit+1),
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeSpear,
			Element:    attributes.Cryo,
			Durability: 25,
			Mult:       mult,
		}
		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 2.5)
		c.Core.Tasks.Add(func() {
			c.Core.QueueAttack(ai, ap, 0, 0, func(a info.AttackCB) {
				if a.Target.Type() != info.TargettableEnemy {
					return
				}
				c.tryParticle()
				if hit == 3 {
					c.etchedFinal(snapped)
				}
			})
		}, delay)
	}
}

func (c *char) etchedFinal(snapped float64) {
	c.tryHexerei(snapped)
	c.armEvilsbane()
	if c.Base.Cons < 6 || !c.masterActive() || c.Core.F < c.c6ICDUntil {
		return
	}
	if c.extraE > c.etchedMax()-1 {
		return
	}
	c.joy = joyCap
	c.c6Mark = true
	c.c6ICDUntil = c.Core.F + c6ICD
}

func (c *char) tryParticle() {
	if c.Core.F < c.particleUntil {
		return
	}
	c.particleUntil = c.Core.F + particleICD
	c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Cryo, c.ParticleDelay)
}
