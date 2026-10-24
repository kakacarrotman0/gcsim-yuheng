package sandrone

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
)

// Skill cancel windows are placeholders. Dash at 12 is the early window KQM warns about.
// Attack, Charged Attack, Burst, and Swap at 36 are long enough for the 0.1s repair ticks to finish a full bar.
// The 6s hover, sprint stamina, and steering are not represented.
var skillFrames []int

func init() {
	skillFrames = frames.InitAbilSlice(hoverFrames)
	skillFrames[action.ActionDash] = skillDash
	skillFrames[action.ActionJump] = skillDash
	skillFrames[action.ActionAttack] = skillEarly
	skillFrames[action.ActionCharge] = skillEarly
	skillFrames[action.ActionBurst] = skillEarly
	skillFrames[action.ActionSwap] = skillEarly
}

func (c *char) Skill(_ map[string]int) (action.Info, error) {
	boosted := c.Base.Ascension >= 1 && c.power >= a1Threshold
	// Opt type 1 refreshes the 60s timer even when the bar is already empty.
	// The 0.1s ticks add FrostReduce and do not touch the timer.
	c.refreshTactics()
	c.notePowerLoss(skillInstant)
	c.addPower(-skillInstant)
	c.stopShooting()
	c.repairGen++
	gen := c.repairGen
	c.Core.Tasks.Add(func() { c.repairTick(gen) }, skillTick)
	// Gadget 41133006 is created with lifeByOwnerV2 and its own 1.45s FunnelEff.
	// ElementalArt_Cancel does not kill it, and the bullets are doOffStage.
	// Repair is the state-attached ReduceFrost modifier. The cannon is not.
	c.Core.Tasks.Add(func() { c.prism(false, boosted, radianceNone) }, prismFirst)
	c.Core.Tasks.Add(func() {
		snap := c.getRadiance()
		c.Core.Tasks.Add(func() { c.prism(true, boosted, snap) }, prismSecond-prismStarSnap)
	}, prismStarSnap)
	c.SetCD(action.ActionSkill, skillCDFrames)
	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
		State:           action.SkillState,
		OnRemoved: func(action.AnimationState) {
			if c.repairGen == gen {
				c.repairGen++
			}
		},
	}, nil
}

func (c *char) repairTick(gen int) {
	if c.repairGen != gen {
		return
	}
	if c.Core.Player.CurrentState() != action.SkillState || !c.onField() {
		c.repairGen++
		return
	}
	c.notePowerLoss(skillTickDrain)
	c.addPower(-skillTickDrain)
	c.Core.Tasks.Add(func() { c.repairTick(gen) }, skillTick)
}

func (c *char) prism(second, boosted bool, rad radianceState) {
	lvl := c.TalentLvlSkill()
	mult := skillOrdinary[lvl]
	if second {
		// FunnelStar is written when FunnelShooterR attaches, at the 0.3s think, not when the bullet lands.
		// Avatar_MarionetteNew_Funnel_Bullet_Enhanced multiplies P1_UpRatio only on the stellar factors.
		boost := 1.0
		if boosted {
			boost = a1PrismMult
		}
		switch rad {
		case radianceStellarConduct:
			ai := c.stellarInfo("Prism Shot"+stellarConductText, attacks.AttackTagDirectStellarConduct, skillConduct[lvl]*boost)
			ai.StrikeType = attacks.StrikeTypeDefault
			c.queue(ai, 3, 0, true)
			return
		case radianceStellarSwirl:
			ai := c.stellarInfo("Prism Shot"+stellarSwirlText, attacks.AttackTagDirectStellarSwirl, skillSwirl[lvl]*boost)
			ai.StrikeType = attacks.StrikeTypeDefault
			c.queue(ai, 3, 0, true)
			return
		}
	}
	name := "Prism Shot"
	if second {
		name = "Prism Shot 2"
	}
	ai := c.cryoInfo(name, attacks.AttackTagElementalArt, attacks.ICDTagElementalArt, attacks.ICDGroupDefault, mult, 25)
	ai.StrikeType = attacks.StrikeTypeDefault
	c.queue(ai, 3, 0, true)
}
