package linnea

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var burstFrames []int

const (
	energyDrainDelay = 4
	// Approximate. The burst-summoned Lumi's first hit is not a measured hitmark.
	firstSkillHitDelay = 214
)

func init() {
	burstFrames = frames.InitAbilSlice(110)
	burstFrames[action.ActionSkill] = 96
	burstFrames[action.ActionSwap] = 97
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	switch {
	case c.StatusIsActive(skillStandardPower):
		// Refresh the existing Lumi. Do not replace skillSrc, or the live ticker stops.
		c.AddStatus(skillStandardPower, skillDur, false)
	case c.StatusIsActive(skillSuperPower):
		c.AddStatus(skillSuperPower, skillDur, false)
	default:
		src := c.Core.F
		c.skillSrc = src
		c.AddStatus(skillSuperPower, skillDur, false)
		c.a1OnLumi(src)
		c.advanceSkillIndex() // the first pound pound is skipped right after summoning
		c.Core.Tasks.Add(func() { c.lumiAttack(src) }, firstSkillHitDelay)
	}

	// HealHP on the burst heal modifier's onAdded. The cast animation before that
	// attach is not in the extracted graph, so the initial heal is the next frame.
	// The arrays are named from the pipeline draft: burstInitialFlat holds the DEF
	// ratio and burstInitialDef holds the flat amount.
	c.QueueCharTask(func() {
		lvl := c.TalentLvlBurst()
		heal := burstInitialDef[lvl] + burstInitialFlat[lvl]*c.TotalDef(false)
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index(),
			Target:  -1,
			Message: "Memo: Survival Guide in Extreme Conditions (Initial)",
			Src:     heal,
			Bonus:   c.Stat(attributes.Heal),
		})
	}, 1)

	// Heal_Interval is 1.0s and Heal_Duration is 12s at every talent level.
	// 12 thinks, the first one interval after the modifier is added.
	for i := 1; i <= 12; i++ {
		c.QueueCharTask(func() {
			lvl := c.TalentLvlBurst()
			heal := burstTickDef[lvl] + burstTickFlat[lvl]*c.TotalDef(false)
			c.Core.Player.Heal(info.HealInfo{
				Caller:  c.Index(),
				Target:  -1,
				Message: "Memo: Survival Guide in Extreme Conditions (Tick)",
				Src:     heal,
				Bonus:   c.Stat(attributes.Heal),
			})
		}, i*60)
	}

	c.SetCD(action.ActionBurst, 15*60)
	c.ConsumeEnergy(energyDrainDelay)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap], // earliest cancel
		State:           action.BurstState,
	}, nil
}
