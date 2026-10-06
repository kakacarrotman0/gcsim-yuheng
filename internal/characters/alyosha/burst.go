package alyosha

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
	// Approximate. Field and Tugarin intervals below come from the ability graph.
	burstFrames = frames.InitAbilSlice(70)
	burstFrames[action.ActionAttack] = 48
	burstFrames[action.ActionSkill] = 46
	burstFrames[action.ActionDash] = 44
	burstFrames[action.ActionJump] = 44
	burstFrames[action.ActionSwap] = 42
}

func (c *char) Burst(_ map[string]int) (action.Info, error) {
	c.fieldSrc = c.Core.F
	src := c.fieldSrc
	c.QueueCharTask(func() {
		c.AddStatus(fieldKey, c.fieldDuration(), true)
		c.fieldPulse(src)
	}, burstHitmark)
	c.ConsumeEnergy(burstHitmark)
	c.SetCDWithDelay(action.ActionBurst, burstCD, 1)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap],
		State:           action.BurstState,
	}, nil
}

func (c *char) fieldPulse(src int) {
	// Status stays active on its expiry frame. Skip that frame so a 14s field
	// pulses at 0,2,4,6,8,10,12 and not again at 14.
	if c.fieldSrc != src || c.StatusDuration(fieldKey) <= 0 {
		return
	}
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Fulgurite Hunting Field",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupAlyoshaElementalBurst,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       burstField[c.TalentLvlBurst()],
		PoiseDMG:   25,
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, fieldRadius), 0, 0)
	c.QueueCharTask(func() { c.tugarin(src) }, dogDelay)
	c.QueueCharTask(func() { c.fieldPulse(src) }, fieldInterval)
}

func (c *char) tugarin(src int) {
	if c.fieldSrc != src || !c.StatusIsActive(fieldKey) {
		return
	}
	target := c.tugarinTarget()
	if target == nil {
		return
	}
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Tugarin",
		AttackTag:  attacks.AttackTagElementalBurst,
		ICDTag:     attacks.ICDTagElementalBurst,
		ICDGroup:   attacks.ICDGroupAlyoshaElementalBurst,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Electro,
		Durability: 25,
		Mult:       burstDog[c.TalentLvlBurst()],
		PoiseDMG:   35,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(target, nil, dogRadius),
		0,
		0,
		func(a info.AttackCB) {
			if c.Base.Cons >= 2 {
				c.markHit(a.Target, false, true)
			} else {
				c.markHit(a.Target, true, false)
			}
			c.tugarinHeal(a)
		},
	)
}

func (c *char) tugarinTarget() info.Target {
	enemies := c.Core.Combat.EnemiesWithinArea(combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, huntRadius), nil)
	var fallback info.Target
	for _, e := range enemies {
		if fallback == nil {
			fallback = e
		}
		if c.marked(e.Key()) {
			return e
		}
	}
	return fallback
}

func (c *char) tugarinHeal(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy {
		return
	}
	if c.Base.Ascension >= 1 {
		active := c.Core.Player.ActiveChar()
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index(),
			Target:  active.Index(),
			Message: "Awakened by the Baying Hounds",
			Src:     c.TotalAtk() * a1HealRatio,
		})
	}
	if c.Base.Cons >= 4 {
		lowest := c.Core.Player.ActiveChar()
		for _, ch := range c.Core.Player.Chars() {
			if ch.CurrentHPRatio() < lowest.CurrentHPRatio() {
				lowest = ch
			}
		}
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index(),
			Target:  lowest.Index(),
			Message: "Harvest the Spoils",
			Src:     c.TotalAtk() * c4HealRatio,
		})
	}
}
