package kachina

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var burstFrames []int

func init() {
	// Cancel windows follow the caramielle community sheet. The strike itself is
	// attached at 0.65 of the burst clip; 35 is that sheet's hitmark, not a new capture.
	burstFrames = frames.InitAbilSlice(75)
	burstFrames[action.ActionAttack] = 56
	burstFrames[action.ActionCharge] = 56
	burstFrames[action.ActionSkill] = 53
	burstFrames[action.ActionDash] = 53
	burstFrames[action.ActionJump] = 53
	burstFrames[action.ActionSwap] = 50
}

func (c *char) Burst(_ map[string]int) (action.Info, error) {
	wasMounted := c.mounted && c.StatusIsActive(twirlyKey)
	c.killTwirlyGadget()
	if c.Base.Cons >= 2 {
		if c.nightsoulState.HasBlessing() {
			c.nightsoulState.GeneratePoints(c2Points)
		} else {
			c.nightsoulState.EnterBlessing(c2Points)
		}
	}

	c.fieldCenter = c.Core.Combat.Player().Pos()
	c.QueueCharTask(func() {
		c.AddStatus(fieldKey, fieldDuration, true)
		c.applyC4(c.Core.Player.Active())
		c.resumeTwirlyAfterBurst(wasMounted)
	}, burstHitmark)

	ai := info.AttackInfo{
		ActorIndex:     c.Index(),
		Abil:           "Time to Get Serious!",
		AttackTag:      attacks.AttackTagElementalBurst,
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeBlunt,
		PoiseDMG:       150,
		Element:        attributes.Geo,
		Durability:     25,
		Mult:           burst[c.TalentLvlBurst()],
		UseDef:         true,
		IgnoreInfusion: true,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, burstHitR),
		burstHitmark,
		burstHitmark,
	)
	c.ConsumeEnergy(burstHitmark)
	c.SetCDWithDelay(action.ActionBurst, burstCD, 1)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap],
		State:           action.BurstState,
	}, nil
}

func (c *char) resumeTwirlyAfterBurst(wasMounted bool) {
	if !c.canPaySlam() {
		if c.nightsoulState.HasBlessing() && c.nightsoulState.Points() <= pointEpsilon {
			c.endTwirly()
		}
		return
	}
	c.AddStatus(twirlyKey, -1, true)
	c.spawnConstruct()
	if wasMounted {
		c.mounted = true
		c.twirlySrc = -1
		return
	}
	c.mounted = false
	c.twirlySrc = c.Core.F
	c.queueIndependent(c.twirlySrc, firstSlam)
}
