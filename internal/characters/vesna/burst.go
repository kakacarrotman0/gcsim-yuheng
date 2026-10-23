package vesna

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
)

// The burst gadget thinks once at 0.55s and then removes itself. That is one
// spirit blade. Burst_MagicSword_Duration 6 and Burst_MagicSword_Interval 0.5
// are declared specials and are not referenced by the ability or gadget file,
// so they are not a 12-hit loop.
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
