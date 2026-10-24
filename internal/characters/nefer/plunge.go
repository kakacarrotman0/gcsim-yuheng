package nefer

import (
	"errors"

	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player"
)

var lowPlungeFrames []int
var highPlungeFrames []int

const (
	lowPlungeHitmark  = 42
	highPlungeHitmark = 46
	collisionHitmark  = 36
)

func init() {
	lowPlungeFrames = frames.InitAbilSlice(62)
	lowPlungeFrames[action.ActionAttack] = lowPlungeHitmark
	lowPlungeFrames[action.ActionCharge] = lowPlungeHitmark
	lowPlungeFrames[action.ActionSkill] = lowPlungeHitmark
	lowPlungeFrames[action.ActionBurst] = lowPlungeHitmark
	lowPlungeFrames[action.ActionDash] = lowPlungeHitmark
	lowPlungeFrames[action.ActionJump] = lowPlungeHitmark
	lowPlungeFrames[action.ActionSwap] = lowPlungeHitmark

	highPlungeFrames = frames.InitAbilSlice(68)
	highPlungeFrames[action.ActionAttack] = highPlungeHitmark
	highPlungeFrames[action.ActionCharge] = highPlungeHitmark
	highPlungeFrames[action.ActionSkill] = highPlungeHitmark
	highPlungeFrames[action.ActionBurst] = highPlungeHitmark
	highPlungeFrames[action.ActionDash] = highPlungeHitmark
	highPlungeFrames[action.ActionJump] = highPlungeHitmark
	highPlungeFrames[action.ActionSwap] = highPlungeHitmark
}

func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)
	if collision, ok := p["collision"]; ok && collision > 0 {
		c.plungeCollision(collisionHitmark)
	}
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Low Plunge",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       lowPlunge[c.TalentLvlAttack()],
	}
	c.queueCircle(ai, lowPlungeHitmark)
	return action.Info{
		Frames:          frames.NewAbilFunc(lowPlungeFrames),
		AnimationLength: lowPlungeFrames[action.InvalidAction],
		CanQueueAfter:   lowPlungeHitmark,
		State:           action.PlungeAttackState,
	}, nil
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
	if collision, ok := p["collision"]; ok && collision > 0 {
		c.plungeCollision(collisionHitmark)
	}
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "High Plunge",
		AttackTag:  attacks.AttackTagPlunge,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Dendro,
		Durability: 25,
		Mult:       highPlunge[c.TalentLvlAttack()],
	}
	c.queueCircle(ai, highPlungeHitmark)
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
		Element:    attributes.Dendro,
		Durability: 0,
		Mult:       plunge[c.TalentLvlAttack()],
	}
	c.queueCircle(ai, delay)
}
