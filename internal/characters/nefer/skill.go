package nefer

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

const (
	particleICDKey = "nefer-particle-icd"
	// One Skill hit can strike several enemies. The wiki generates particles
	// once, when Skill DMG hits at least one enemy, so the rolls share this window.
	particleICD = 18
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
	c.queueCircle(c.hybrid("Senet Strategy: Dance of a Thousand Nights", attacks.AttackTagElementalArt, attacks.ICDTagElementalArt, skillATK[lvl], skillEM[lvl]), skillHit, c.particleCB)
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

// The wiki gameplay note is "When Skill DMG hits at least one enemy, it generates 2.6667 Elemental Particles."
// 2.6667 is 8/3, the usual 2 particles plus a 2/3 chance of a third.
func (c *char) particleCB(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, particleICD, true)
	count := 2.0
	if c.Core.Rand.Float64() < 2.0/3.0 {
		count = 3
	}
	c.Core.QueueParticle(c.Base.Key.String(), count, attributes.Dendro, c.ParticleDelay)
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
