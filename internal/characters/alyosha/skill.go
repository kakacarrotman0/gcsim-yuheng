package alyosha

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var skillFrames [][]int

func init() {
	// Approximate cast windows. The live clip length was not captured.
	skillFrames = make([][]int, 2)
	skillFrames[0] = frames.InitAbilSlice(46)
	skillFrames[0][action.ActionBurst] = 28
	skillFrames[0][action.ActionDash] = 30
	skillFrames[0][action.ActionJump] = 30
	skillFrames[0][action.ActionSwap] = 32
	skillFrames[1] = frames.InitAbilSlice(78)
	skillFrames[1][action.ActionBurst] = 56
	skillFrames[1][action.ActionDash] = 60
	skillFrames[1][action.ActionJump] = 60
	skillFrames[1][action.ActionSwap] = 64
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	hold := 0
	if p["hold"] > 0 {
		hold = 1
	}
	hitmark := skillTapHitmark
	mult := skillTap[c.TalentLvlSkill()]
	pattern := combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 2.5, 10)
	if hold == 1 {
		hitmark = skillHoldHitmark
		mult = skillHold[c.TalentLvlSkill()]
	}
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Thunderbolt Strike",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSpear,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       mult,
		PoiseDMG:   125,
	}
	c.Core.QueueAttack(ai, pattern, hitmark, hitmark, func(a info.AttackCB) {
		c.markHit(a.Target, true, true)
		c.skillParticle(a)
	})
	c.SetCDWithDelay(action.ActionSkill, skillCD, 8)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames[hold]),
		AnimationLength: skillFrames[hold][action.InvalidAction],
		CanQueueAfter:   skillFrames[hold][action.ActionBurst],
		State:           action.SkillState,
	}, nil
}

func (c *char) skillParticle(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy || c.StatusIsActive(particleICDKey) {
		return
	}
	c.AddStatus(particleICDKey, particleICD, true)
	c.Core.QueueParticle(c.Base.Key.String(), particleCount, attributes.Electro, c.ParticleDelay)
}
