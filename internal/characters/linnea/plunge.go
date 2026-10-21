package linnea

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

// Bow plunge cancel windows are approximate. The upstream Linnea draft has no
// plunge file, and no Linnea plunge sheet was used here.
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
	lowPlungeFrames = frames.InitAbilSlice(70)
	lowPlungeFrames[action.ActionAttack] = 55
	lowPlungeFrames[action.ActionAim] = 55
	lowPlungeFrames[action.ActionSkill] = 55
	lowPlungeFrames[action.ActionBurst] = 55
	lowPlungeFrames[action.ActionDash] = 60
	lowPlungeFrames[action.ActionSwap] = 58

	highPlungeFrames = frames.InitAbilSlice(72)
	highPlungeFrames[action.ActionAttack] = 56
	highPlungeFrames[action.ActionAim] = 56
	highPlungeFrames[action.ActionSkill] = 56
	highPlungeFrames[action.ActionBurst] = 56
	highPlungeFrames[action.ActionDash] = 62
	highPlungeFrames[action.ActionSwap] = 60
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
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       abil,
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeBlunt,
		Element:    attributes.Physical,
		Durability: 25,
		Mult:       mult,
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
