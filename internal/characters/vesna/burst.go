package vesna

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

// SkillObj_Vesna_MagicSword_Big_03 thinks once at 0.55s, deals one hit, and
// removes the modifier. Burst_MagicSword_Duration 6 and Burst_MagicSword_Interval
// 0.5 are declared specials with no other reference, so they are not a 12-hit
// loop. The talent table is one Spirit Blade percentage, not ×N. The description
// "multiple Spirit Blades" uses the same glossary as Lv2, Lv3, and Transpose.
var burstFrames []int

const (
	burstBladeDelay = 33
	burstRadius     = 6.5
)

func init() {
	burstFrames = frames.InitAbilSlice(50)
	burstFrames[action.ActionAttack] = 40
	burstFrames[action.ActionSkill] = 40
	burstFrames[action.ActionDash] = 36
	burstFrames[action.ActionSwap] = 42
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	// Essence is added on burst start, including while Armed is down.
	// It stays until an Armed exit clears it.
	c.addEssence(essenceBurst)
	c.addDisc()
	c.queueSpirit("For the Tsaritsa!", burstDMG[c.TalentLvlBurst()], burstBladeDelay, true, burstRadius)
	c.SetCD(action.ActionBurst, burstCD)
	c.ConsumeEnergy(1)
	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}
