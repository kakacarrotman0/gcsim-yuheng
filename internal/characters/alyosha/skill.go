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
	// Tap and hold cancels are medians from the caramielle skill sheet.
	skillFrames = make([][]int, 2)
	skillFrames[0] = frames.InitAbilSlice(31) // walk 31/29/31
	skillFrames[0][action.ActionAttack] = 30  // 30/29/30
	skillFrames[0][action.ActionBurst] = 30
	skillFrames[0][action.ActionDash] = 31 // 29/31/31
	skillFrames[0][action.ActionJump] = 31
	skillFrames[0][action.ActionSwap] = 30 // 30/28/30

	skillFrames[1] = frames.InitAbilSlice(142) // walk 142/142/141
	skillFrames[1][action.ActionAttack] = 138
	skillFrames[1][action.ActionBurst] = 138 // 138/137/138
	skillFrames[1][action.ActionDash] = 138  // 138/139/138
	skillFrames[1][action.ActionJump] = 138  // 138/138/137
	skillFrames[1][action.ActionSwap] = 137  // 137/137/136
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	hold := 0
	if p["hold"] > 0 {
		hold = 1
	}
	hitmark := skillTapHitmark
	cdDelay := skillTapCDDelay
	mult := skillTap[c.TalentLvlSkill()]
	tag := attacks.AttackTagElementalArt
	abil := "Thunderbolt Strike"
	// Tap is the ability box, x 2.5 by z 10.
	pattern := combat.NewBoxHitOnTarget(c.Core.Combat.Player(), nil, 2.5, 10)
	if hold == 1 {
		hitmark = skillHoldHitmark
		cdDelay = skillHoldCDDelay
		mult = skillHold[c.TalentLvlSkill()]
		// Hold is Elemental_Art in the ability graph. gcsim records the hold as its own tag.
		tag = attacks.AttackTagElementalArtHold
		abil = "Thunderbolt Strike (Hold)"
		// Hold pattern is a 150 degree fan of radius 15.
		pattern = combat.NewCircleHitOnTargetFanAngle(c.Core.Combat.Player(), nil, 15, 150)
	}
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       abil,
		AttackTag:  tag,
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
	c.SetCDWithDelay(action.ActionSkill, skillCD, cdDelay)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames[hold]),
		AnimationLength: skillFrames[hold][action.InvalidAction],
		CanQueueAfter:   skillFrames[hold][action.ActionSwap],
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
