package alyosha

import (
	"errors"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player"
)

var (
	highPlungeFrames []int
	lowPlungeFrames  []int
)

const (
	// Plunge cancels are the standard polearm template. They are approximate for Alyosha.
	lowPlungeHitmark   = 42 + 3
	highPlungeHitmark  = 43 + 3
	collisionHitmark   = lowPlungeHitmark - 6
	lowPlungePoiseDMG  = 100.0
	lowPlungeRadius    = 3.0
	highPlungePoiseDMG = 150.0
	highPlungeRadius   = 5.0
)

func init() {
	lowPlungeFrames = frames.InitAbilSlice(77)
	lowPlungeFrames[action.ActionAttack] = 59
	lowPlungeFrames[action.ActionSkill] = 59
	lowPlungeFrames[action.ActionBurst] = 58
	lowPlungeFrames[action.ActionDash] = 75
	lowPlungeFrames[action.ActionWalk] = 74
	lowPlungeFrames[action.ActionSwap] = 62

	highPlungeFrames = frames.InitAbilSlice(77)
	highPlungeFrames[action.ActionAttack] = 58
	highPlungeFrames[action.ActionSkill] = 60
	highPlungeFrames[action.ActionBurst] = 60
	highPlungeFrames[action.ActionWalk] = 76
	highPlungeFrames[action.ActionSwap] = 63
}

func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)
	switch c.Core.Player.Airborne() {
	case player.AirborneXianyun:
		return c.lowPlungeXY(p), nil
	default:
		return action.Info{}, errors.New("low_plunge can only be used while airborne")
	}
}

func (c *char) lowPlungeXY(p map[string]int) action.Info {
	if p["collision"] > 0 {
		c.plungeCollision(collisionHitmark)
	}
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Low Plunge",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   lowPlungePoiseDMG,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       lowPlunge[c.TalentLvlAttack()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), info.Point{Y: 1}, lowPlungeRadius), lowPlungeHitmark, lowPlungeHitmark)
	return action.Info{
		Frames:          frames.NewAbilFunc(lowPlungeFrames),
		AnimationLength: lowPlungeFrames[action.InvalidAction],
		CanQueueAfter:   lowPlungeFrames[action.ActionBurst],
		State:           action.PlungeAttackState,
	}
}

func (c *char) HighPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)
	switch c.Core.Player.Airborne() {
	case player.AirborneXianyun:
		return c.highPlungeXY(p), nil
	default:
		return action.Info{}, errors.New("high_plunge can only be used while airborne")
	}
}

func (c *char) highPlungeXY(p map[string]int) action.Info {
	if p["collision"] > 0 {
		c.plungeCollision(collisionHitmark)
	}
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "High Plunge",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		PoiseDMG:   highPlungePoiseDMG,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       highPlunge[c.TalentLvlAttack()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), info.Point{Y: 1}, highPlungeRadius), highPlungeHitmark, highPlungeHitmark)
	return action.Info{
		Frames:          frames.NewAbilFunc(highPlungeFrames),
		AnimationLength: highPlungeFrames[action.InvalidAction],
		CanQueueAfter:   highPlungeFrames[action.ActionAttack],
		State:           action.PlungeAttackState,
	}
}

func (c *char) plungeCollision(delay int) {
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Plunge Collision",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeSlash,
		Element:    attributes.Physical,
		Durability: 0,
		Mult:       collision[c.TalentLvlAttack()],
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), info.Point{Y: 1}, 1), delay, delay)
}
