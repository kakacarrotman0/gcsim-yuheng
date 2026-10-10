package lohen

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

// Plunge cancel windows are the standard polearm template. No Lohen plunge sheet
// was used. Masterstroke uses the Ice FallingAnthem attacks; the ratios on the
// skill talent match the normal-attack plunge ratios.
var (
	highPlungeFrames []int
	lowPlungeFrames  []int
)

const (
	lowPlungeHitmark  = 45
	highPlungeHitmark = 46
	collisionHitmark  = lowPlungeHitmark - 6
)

func init() {
	lowPlungeFrames = frames.InitAbilSlice(77)
	lowPlungeFrames[action.ActionAttack] = 59
	lowPlungeFrames[action.ActionSkill] = 59
	lowPlungeFrames[action.ActionBurst] = 58
	lowPlungeFrames[action.ActionDash] = 75
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
		return c.plungeXY(p, false), nil
	default:
		return action.Info{}, errors.New("low_plunge can only be used while airborne")
	}
}

func (c *char) HighPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)
	switch c.Core.Player.Airborne() {
	case player.AirborneXianyun:
		return c.plungeXY(p, true), nil
	default:
		return action.Info{}, errors.New("high_plunge can only be used while airborne")
	}
}

func (c *char) plungeXY(p map[string]int, high bool) action.Info {
	if p["collision"] > 0 {
		c.plungeCollision(collisionHitmark)
	}
	// Plunge ratios match across the normal and skill sheets at the same
	// level. Masterstroke reads the skill-talent row, which High Spirits and
	// C3 can raise above the normal-attack level.
	lvl := c.TalentLvlAttack()
	if c.masterActive() {
		lvl = c.skillLvl()
	}
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
	if c.masterActive() {
		ele = attributes.Cryo
		ignore = true
		abil = "Masterstroke " + abil
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
		c.hitCB(0, false),
	)
	return action.Info{
		Frames:          frames.NewAbilFunc(frameset),
		AnimationLength: frameset[action.InvalidAction],
		CanQueueAfter:   frameset[action.ActionBurst],
		State:           action.PlungeAttackState,
	}
}

func (c *char) plungeCollision(delay int) {
	// FallingAnthem loop is Physical outside Masterstroke and Ice, with no
	// gauge, while the mode is active. The Ice hit still counts as a
	// FallingAttack for the particle handler.
	lvl := c.TalentLvlAttack()
	ele := attributes.Physical
	ignore := false
	if c.masterActive() {
		lvl = c.skillLvl()
		ele = attributes.Cryo
		ignore = true
	}
	ai := info.AttackInfo{
		ActorIndex:     c.Index(),
		Abil:           "Plunge Collision",
		AttackTag:      attacks.AttackTagPlunge,
		ICDTag:         attacks.ICDTagNone,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeSlash,
		Element:        ele,
		Durability:     0,
		Mult:           collision[lvl],
		IgnoreInfusion: ignore,
	}
	var cb info.AttackCBFunc
	if ele == attributes.Cryo {
		cb = c.hitCB(0, false)
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), info.Point{Y: 1}, 1), delay, delay, cb)
}
