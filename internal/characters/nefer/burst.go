package nefer

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
)

var burstFrames []int

func init() {
	burstFrames = frames.InitAbilSlice(burstAnim)
	burstFrames[action.ActionAttack] = burstHit1
	burstFrames[action.ActionCharge] = burstHit1
	burstFrames[action.ActionSkill] = burstHit1
	burstFrames[action.ActionBurst] = burstHit1
	burstFrames[action.ActionDash] = burstHit1
	burstFrames[action.ActionJump] = burstHit1
	burstFrames[action.ActionSwap] = burstHit1
}

func (c *char) Burst(_ map[string]int) (action.Info, error) {
	c.ConsumeEnergy(burstHit1)
	stacks := c.veilCount()
	c.veils = nil
	c.burstBonus = float64(stacks) * burstVeil[c.TalentLvlBurst()]
	lvl := c.TalentLvlBurst()
	c.queueCircle(c.hybrid("Sacred Vow (1)", attacks.AttackTagElementalBurst, attacks.ICDTagElementalBurst, burst1ATK[lvl], burst1EM[lvl]), burstHit1)
	c.queueCircle(c.hybrid("Sacred Vow (2)", attacks.AttackTagElementalBurst, attacks.ICDTagElementalBurst, burst2ATK[lvl], burst2EM[lvl]), burstHit2)
	c.SetCDWithDelay(action.ActionBurst, burstCD, burstHit1)
	c.QueueCharTask(func() { c.burstBonus = 0 }, burstHit2+1)
	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstHit1,
		State:           action.BurstState,
	}, nil
}
