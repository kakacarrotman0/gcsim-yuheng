package sandrone

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
)

// Bombardment follows MarionetteNewBurstAttacker: one hit on attach, then every 0.2s while the count is <= 2.
// That is three hits, at frames 0, 12, and 24. The ray is the removal of MarionetteNewBurstAttackerLaser, whose 0.85s think removes it.
// SetStarModeDelay snapshots _IS_IN_STARMODE at 0.28s. The ray reads that snapshot, not radiance at 0.85s.
// 0.28s is 16.8 frames and is rounded to 17. These are ability intervals, not measured burst hitmarks.
var burstFrames []int

func init() {
	burstFrames = frames.InitAbilSlice(70)
	burstFrames[action.ActionDash] = 58
	burstFrames[action.ActionSwap] = 62
	burstFrames[action.ActionAttack] = 64
	burstFrames[action.ActionSkill] = 64
	burstFrames[action.ActionCharge] = 64
}

func (c *char) Burst(_ map[string]int) (action.Info, error) {
	c.SetCD(action.ActionBurst, burstCDFrames)
	// The drain frame is not on a measured burst sheet. Consume on cast.
	c.ConsumeEnergy(0)
	stacks := 0
	if c.Base.Ascension >= 1 {
		// ElementalBurstEnhance is triggered with only the A1 talent predicate.
		// It zeros FrostReduce and removes the 60s timer on cast, including outside Radiance.
		// The talent sentence names Radiance: Stellar Glimmer. The stored ratio is still
		// applied only when the 0.28s ray snapshot is a stellar ray.
		stacks = c.tactics()
		c.frostReduce = 0
		c.tacticsUntil = c.Core.F
	}
	for _, delay := range []int{0, 12, 24} {
		d := delay
		c.Core.Tasks.Add(func() { c.bombard() }, d)
	}
	c.Core.Tasks.Add(func() {
		snap := c.getRadiance()
		c.Core.Tasks.Add(func() { c.ray(stacks, snap) }, burstRayAt-burstStarSnap)
	}, burstStarSnap)
	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}

func (c *char) bombard() {
	lvl := c.TalentLvlBurst()
	ai := c.cryoInfo("Bombardment", attacks.AttackTagElementalBurst, attacks.ICDTagElementalBurst, attacks.ICDGroupDefault, burstBombard[lvl], 25)
	ai.StrikeType = attacks.StrikeTypeDefault
	c.queue(ai, 7.3, 0, false)
}

func (c *char) ray(stacks int, rad radianceState) {
	lvl := c.TalentLvlBurst()
	mult := 1.0
	// BurstP1RatioUp is stored at cast. Both stellar ray factors multiply it.
	// The ordinary BurstLaserRatio path does not. Stacks are already cleared.
	if rad != radianceNone {
		mult = 1 + a1RayPer*float64(stacks)
	}
	switch rad {
	case radianceStellarConduct:
		ai := c.stellarInfo("Convective Inhibition Ray"+stellarConductText, attacks.AttackTagDirectStellarConduct, burstRayConduct[lvl]*mult)
		c.queue(ai, 7.5, 0, false)
	case radianceStellarSwirl:
		ai := c.stellarInfo("Convective Inhibition Ray"+stellarSwirlText, attacks.AttackTagDirectStellarSwirl, burstRaySwirl[lvl]*mult)
		c.queue(ai, 7.5, 0, false)
	default:
		ai := c.cryoInfo("Convective Inhibition Ray", attacks.AttackTagElementalBurst, attacks.ICDTagElementalBurst, attacks.ICDGroupDefault, burstRayOrdinary[lvl], 25)
		c.queue(ai, 7.5, 0, false)
	}
}
