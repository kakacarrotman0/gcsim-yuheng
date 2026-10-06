package kachina

import (
	"errors"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var (
	skillFrames       []int
	skillRecastFrames []int
)

func init() {
	// Cancel windows follow the caramielle community sheet cited by gcsim PR #2697.
	// They are not an independent live capture.
	skillFrames = frames.InitAbilSlice(733)
	skillFrames[action.ActionAttack] = 56
	skillFrames[action.ActionBurst] = 30
	skillFrames[action.ActionSkill] = 51
	skillFrames[action.ActionDash] = 55
	skillFrames[action.ActionJump] = 57
	skillFrames[action.ActionWalk] = 45
	skillFrames[action.ActionSwap] = 44

	skillRecastFrames = frames.InitAbilSlice(30)
	skillRecastFrames[action.ActionAttack] = 30
	skillRecastFrames[action.ActionBurst] = 30
	skillRecastFrames[action.ActionDash] = 30
	skillRecastFrames[action.ActionJump] = 30
	skillRecastFrames[action.ActionWalk] = 30
	skillRecastFrames[action.ActionSwap] = 30
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	hold := p["hold"] > 0
	if c.StatusIsActive(twirlyKey) {
		if p["recast"] == 0 && p["hold"] == 0 {
			return c.toggleTwirly(false), nil
		}
		return c.toggleTwirly(hold), nil
	}
	if p["recast"] > 0 {
		return action.Info{}, errors.New("cannot recast E while Turbo Twirly is inactive")
	}

	c.startTwirly(nightsoulMax, hold)
	c.SetCDWithDelay(action.ActionSkill, skillCD, 10)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionBurst],
		State:           action.SkillState,
	}, nil
}

func (c *char) toggleTwirly(forceMount bool) action.Info {
	if forceMount {
		c.mountTwirly()
	} else if c.mounted {
		c.dismountTwirly()
	} else {
		c.mountTwirly()
	}
	return action.Info{
		Frames:          frames.NewAbilFunc(skillRecastFrames),
		AnimationLength: skillRecastFrames[action.InvalidAction],
		CanQueueAfter:   skillRecastFrames[action.ActionAttack],
		State:           action.SkillState,
	}
}

func (c *char) mountTwirly() {
	c.mounted = true
	c.twirlySrc = -1
	c.AddStatus(twirlyKey, -1, true)
}

func (c *char) dismountTwirly() {
	c.consumeTwirlyPoints(dismountCost)
	if !c.canPaySlam() {
		return
	}
	c.mounted = false
	c.twirlySrc = c.Core.F
	c.AddStatus(twirlyKey, -1, true)
	c.queueIndependent(c.twirlySrc, firstSlam)
}

func (c *char) queueIndependent(src, delay int) {
	c.QueueCharTask(func() {
		c.independentSlam(src)
	}, delay)
}

func (c *char) independentSlam(src int) {
	if c.twirlySrc != src || !c.StatusIsActive(twirlyKey) || c.mounted || !c.canPaySlam() {
		if c.twirlySrc == src && !c.mounted && !c.canPaySlam() {
			c.endTwirly()
		}
		return
	}
	ai := c.twirlyAttackInfo("Turbo Twirly", independent[c.TalentLvlSkill()], 50)
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, c.twirlyRadius()),
		0,
		0,
		c.twirlyParticleCB,
	)
	c.consumeTwirlyPoints(slamCost)
	if c.twirlySrc == src && c.StatusIsActive(twirlyKey) && !c.mounted && c.canPaySlam() {
		c.queueIndependent(src, slamInterval)
	}
}

func (c *char) twirlyParticleCB(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy || c.StatusIsActive(particleICD) {
		return
	}
	c.AddStatus(particleICD, 12, true)
	// baseEnergy 3 is one same-element particle. Fischl uses this value for Oz's single particle.
	c.Core.QueueParticle(c.Base.Key.String(), 1, attributes.Geo, c.ParticleDelay)
}
