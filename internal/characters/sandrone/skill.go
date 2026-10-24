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
	c.notePowerLoss(skillInstant)
	c.addPower(-skillInstant)
	c.stopShooting()
	c.repairGen++
	gen := c.repairGen
	c.Core.Tasks.Add(func() { c.repairTick(gen) }, skillTick)
	c.Core.Tasks.Add(func() { c.prism(false, boosted) }, prismFirst)
	c.Core.Tasks.Add(func() { c.prism(true, boosted) }, prismSecond)
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

func (c *char) prism(second, boosted bool) {
	if !c.onField() || c.Core.Player.CurrentState() != action.SkillState {
		return
	}
	lvl := c.TalentLvlSkill()
	mult := skillOrdinary[lvl]
	if second {
		rad := c.getRadiance()
		// Avatar_MarionetteNew_Funnel_Bullet_Enhanced multiplies P1_UpRatio only on the stellar factors.
		// The ordinary Elemental_Art shot uses DamageRatio alone. The 400% talent sentence sits under the Radiance prefix.
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
