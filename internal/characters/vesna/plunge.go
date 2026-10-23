package vesna

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

// Plunge windows are an unverified sword spacing. Armed Falling Anthem is one
// Wind blunt hit. Its ratio matches the normal-attack plunge row. Gauge is 1U
// on the impact. Collision stays 0-gauge so the impact is the only application.
var (
	highPlungeFrames []int
	lowPlungeFrames  []int
)

const (
	lowPlungeHitmark  = 36
	highPlungeHitmark = 38
	collisionHitmark  = 30
)

func init() {
	lowPlungeFrames = frames.InitAbilSlice(60)
	lowPlungeFrames[action.ActionAttack] = 48
	lowPlungeFrames[action.ActionSkill] = 48
	lowPlungeFrames[action.ActionBurst] = 48
	lowPlungeFrames[action.ActionDash] = 50
	lowPlungeFrames[action.ActionSwap] = 52

	highPlungeFrames = frames.InitAbilSlice(62)
	highPlungeFrames[action.ActionAttack] = 50
	highPlungeFrames[action.ActionSkill] = 50
	highPlungeFrames[action.ActionBurst] = 50
	highPlungeFrames[action.ActionDash] = 52
	highPlungeFrames[action.ActionSwap] = 54
}

func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)
	switch c.Core.Player.Airborne() {
	case player.AirborneXianyun:
		return c.plunge(p, false), nil
	default:
		return action.Info{}, errors.New("low_plunge can only be used while airborne")
	}
}

func (c *char) HighPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)
	switch c.Core.Player.Airborne() {
	case player.AirborneXianyun:
		return c.plunge(p, true), nil
	default:
		return action.Info{}, errors.New("high_plunge can only be used while airborne")
	}
}

func (c *char) plunge(p map[string]int, high bool) action.Info {
	// The wind hit is chosen when the plunge starts. Its cancel is after the impact.
	armed := c.armed()
	if armed {
		c.summonPinion()
	}
	if p["collision"] > 0 {
		c.plungeCollision(armed)
	}
	lvl := c.TalentLvlAttack()
	mult := lowPlunge[lvl]
	hitmark := lowPlungeHitmark
	abil := "Low Plunge"
	frameset := lowPlungeFrames
	if high {
		mult = highPlunge[lvl]
		hitmark = highPlungeHitmark
		abil = "High Plunge"
		frameset = highPlungeFrames
	}
	ele := attributes.Physical
	ignore := false
	if armed {
		ele = attributes.Anemo
		ignore = true
		abil = "Armed " + abil
	}
	ai := info.AttackInfo{
		ActorIndex:     c.Index(),
		Abil:           abil,
		AttackTag:      attacks.AttackTagPlunge,
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeBlunt,
		Element:        ele,
		Durability:     25,
		Mult:           mult,
		IgnoreInfusion: ignore,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), info.Point{Y: 1}, 3),
		hitmark,
		hitmark,
	)
	return action.Info{
		Frames:          frames.NewAbilFunc(frameset),
		AnimationLength: frameset[action.InvalidAction],
		CanQueueAfter:   frameset[action.ActionAttack],
		State:           action.PlungeAttackState,
	}
}

func (c *char) plungeCollision(armed bool) {
	ele := attributes.Physical
	ignore := false
	if armed {
		ele = attributes.Anemo
		ignore = true
	}
	ai := info.AttackInfo{
		ActorIndex:     c.Index(),
		Abil:           "Plunge Collision",
		AttackTag:      attacks.AttackTagPlunge,
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeBlunt,
		Element:        ele,
		Durability:     0,
		Mult:           collision[c.TalentLvlAttack()],
		IgnoreInfusion: ignore,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), info.Point{Y: 1}, 1),
		collisionHitmark,
		collisionHitmark,
	)
}
