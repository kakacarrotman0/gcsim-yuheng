package nefer

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
)

var skillFrames []int

func init() {
	skillFrames = frames.InitAbilSlice(skillAnim)
	skillFrames[action.ActionAttack] = skillHit
	skillFrames[action.ActionCharge] = skillHit
	skillFrames[action.ActionSkill] = skillHit
	skillFrames[action.ActionBurst] = skillHit
	skillFrames[action.ActionDash] = skillHit
	skillFrames[action.ActionJump] = skillHit
	skillFrames[action.ActionSwap] = skillHit
}

func (c *char) Skill(_ map[string]int) (action.Info, error) {
	lvl := c.TalentLvlSkill()
	c.queueCircle(c.hybrid("Senet Strategy: Dance of a Thousand Nights", attacks.AttackTagElementalArt, attacks.ICDTagElementalArt, 25, skillATK[lvl], skillEM[lvl]), skillHit)
	c.SetCDWithDelay(action.ActionSkill, skillCD, skillHit)
	c.enterDance()
	if c.Base.Cons >= 2 && c.Base.Ascension >= 1 {
		c.addVeils(2)
	}
	if c.gleam && c.Base.Ascension >= 1 {
		c.AddStatus(deceitKey, deceitDur, true)
		c.markCores()
	}
	c.syncC4()
	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillHit,
		State:           action.SkillState,
	}, nil
}

func (c *char) enterDance() {
	c.danceSrc = c.Core.F
	c.ppCount = 0
	c.AddStatus(shadowKey, shadowDur, true)
	src := c.danceSrc
	c.QueueCharTask(func() {
		if c.danceSrc != src {
			return
		}
		c.exitDance()
	}, shadowDur)
}

func (c *char) exitDance() {
	if c.danceSrc < 0 && !c.StatusIsActive(shadowKey) {
		return
	}
	c.danceSrc = -1
	c.ppCount = 0
	c.DeleteStatus(shadowKey)
	c.syncC4()
}
