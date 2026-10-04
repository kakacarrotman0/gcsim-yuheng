package vodyanitsa

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

// APPROXIMATE hitmarks. Radii and gauges are from her FallingAnthem ability:
// collision radius 1.5 gauge 0, low radius 3 gauge 1U, high radius 3.5 gauge 1U.
const (
	lowPlungeHitmark  = 42
	highPlungeHitmark = 48
	collisionHitmark  = 36
	lowPlungeRadius   = 3.0
	highPlungeRadius  = 3.5
)

func init() {
	lowPlungeFrames = frames.InitAbilSlice(68)
	lowPlungeFrames[action.ActionAttack] = 56
	lowPlungeFrames[action.ActionCharge] = 54
	lowPlungeFrames[action.ActionSkill] = lowPlungeHitmark
	lowPlungeFrames[action.ActionBurst] = lowPlungeHitmark
	lowPlungeFrames[action.ActionDash] = lowPlungeHitmark
	lowPlungeFrames[action.ActionSwap] = 50

	highPlungeFrames = frames.InitAbilSlice(74)
	highPlungeFrames[action.ActionAttack] = 62
	highPlungeFrames[action.ActionCharge] = 60
	highPlungeFrames[action.ActionSkill] = highPlungeHitmark
	highPlungeFrames[action.ActionBurst] = highPlungeHitmark
	highPlungeFrames[action.ActionDash] = highPlungeHitmark
	highPlungeFrames[action.ActionJump] = 72
	highPlungeFrames[action.ActionSwap] = 54
}

func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)
	switch c.Core.Player.Airborne() {
	case player.AirborneXianyun, player.AirborneStellarSwirl:
		return c.lowPlunge(p), nil
	default:
		return action.Info{}, errors.New("low_plunge can only be used while airborne")
	}
}

func (c *char) lowPlunge(p map[string]int) action.Info {
	if p["collision"] > 0 {
		c.plungeCollision(collisionHitmark)
	}
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Low Plunge",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       lowPlunge[c.TalentLvlAttack()],
		PoiseDMG:   50,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), info.Point{Y: -0.5}, lowPlungeRadius),
		lowPlungeHitmark,
		lowPlungeHitmark,
	)
	return action.Info{
		Frames:          frames.NewAbilFunc(lowPlungeFrames),
		AnimationLength: lowPlungeFrames[action.InvalidAction],
		CanQueueAfter:   lowPlungeHitmark,
		State:           action.PlungeAttackState,
	}
}

func (c *char) HighPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)
	switch c.Core.Player.Airborne() {
	case player.AirborneXianyun, player.AirborneStellarSwirl:
		return c.highPlunge(p), nil
	default:
		return action.Info{}, errors.New("high_plunge can only be used while airborne")
	}
}

func (c *char) highPlunge(p map[string]int) action.Info {
	if p["collision"] > 0 {
		c.plungeCollision(collisionHitmark)
	}
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "High Plunge",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		Mult:       highPlunge[c.TalentLvlAttack()],
		PoiseDMG:   100,
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), info.Point{Y: -0.5}, highPlungeRadius),
		highPlungeHitmark,
		highPlungeHitmark,
	)
	return action.Info{
		Frames:          frames.NewAbilFunc(highPlungeFrames),
		AnimationLength: highPlungeFrames[action.InvalidAction],
		CanQueueAfter:   highPlungeHitmark,
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
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 0,
		Mult:       collision[c.TalentLvlAttack()],
		PoiseDMG:   5,
	}
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 1.5), delay, delay)
}
