package lohen

import (
	"fmt"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var burstFrames []int

func init() {
	// Unverified cancel windows. Hit times follow the gadget's 0.1s counter.
	burstFrames = frames.InitAbilSlice(90)
	burstFrames[action.ActionAttack] = 80
	burstFrames[action.ActionSkill] = 80
	burstFrames[action.ActionDash] = 78
	burstFrames[action.ActionSwap] = 82
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	if c.Base.Cons >= 4 && c.masterActive() {
		c.will = c.willMax()
	}
	snapped := c.will
	lvl := c.TalentLvlBurst()
	ratio := c.damageRatio(snapped, burstWillRatio[lvl])
	if !c.c6KeepWill() {
		c.will = 0
	}
	if c.masterActive() {
		c.extendMaster(burstExtend)
	}
	c.tryHexerei(snapped)
	if c.Core.F < c.c4RefundUntil {
		c.AddEnergy("lohen-c4", c4Energy)
		c.c4RefundUntil = 0
	}

	mult := burstDMG[lvl] * ratio
	for i, delay := range burstHitmarks {
		hit := i
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       fmt.Sprintf("Manifest Judgment %v", hit+1),
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagElementalBurst,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Cryo,
			Durability: 25,
			Mult:       mult,
		}
		// Gadget hitbox was not extracted.
		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 4)
		last := hit == len(burstHitmarks)-1
		c.Core.Tasks.Add(func() {
			c.Core.QueueAttack(ai, ap, 0, 0, func(a info.AttackCB) {
				if a.Target.Type() != info.TargettableEnemy || !last {
					return
				}
				// Burst handler onRemoved. The state-exit frame is not extracted,
				// so this is tied to the last gadget hit.
				c.armEvilsbane()
				c.etchedFinal(snapped)
			})
		}, delay)
	}

	c.SetCD(action.ActionBurst, burstCDDur)
	c.ConsumeEnergy(4)
	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionDash],
		State:           action.BurstState,
	}, nil
}
