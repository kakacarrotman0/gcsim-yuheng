package sandrone

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

// Plunge windows are the standard claymore template. No Sandrone plunge sheet was used.
// The falling attacks in the ability dump do not set Ice, so plunge stays Physical.
var (
	lowPlungeFrames  []int
	highPlungeFrames []int
)

const (
	lowPlungeHitmark  = 42
	highPlungeHitmark = 44
	collisionHitmark  = 36
)

func init() {
	lowPlungeFrames = frames.InitAbilSlice(70)
	lowPlungeFrames[action.ActionAttack] = 56
	lowPlungeFrames[action.ActionSkill] = 56
	lowPlungeFrames[action.ActionBurst] = 56
	lowPlungeFrames[action.ActionDash] = 64
	lowPlungeFrames[action.ActionSwap] = 58

	highPlungeFrames = frames.InitAbilSlice(72)
	highPlungeFrames[action.ActionAttack] = 58
	highPlungeFrames[action.ActionSkill] = 58
	highPlungeFrames[action.ActionBurst] = 58
	highPlungeFrames[action.ActionDash] = 66
	highPlungeFrames[action.ActionSwap] = 60
}

func (c *char) LowPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)
	if c.Core.Player.Airborne() == player.Grounded {
		return action.Info{}, errors.New("low_plunge can only be used while airborne")
	}
	return c.plunge(p, false), nil
}

func (c *char) HighPlungeAttack(p map[string]int) (action.Info, error) {
	defer c.Core.Player.SetAirborne(player.Grounded)
	if c.Core.Player.Airborne() == player.Grounded {
		return action.Info{}, errors.New("high_plunge can only be used while airborne")
	}
	return c.plunge(p, true), nil
}

func (c *char) plunge(p map[string]int, high bool) action.Info {
	if p["collision"] > 0 {
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Plunge Collision",
			AttackTag:  attacks.AttackTagPlunge,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeBlunt,
			Element:    attributes.Physical,
			Durability: 0,
			Mult:       collision[c.TalentLvlAttack()],
		}
		c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 1), collisionHitmark, collisionHitmark)
	}
	mult := lowPlunge[c.TalentLvlAttack()]
	hitmark := lowPlungeHitmark
	abil := "Low Plunge"
	set := lowPlungeFrames
	if high {
		mult = highPlunge[c.TalentLvlAttack()]
		hitmark = highPlungeHitmark
		abil = "High Plunge"
		set = highPlungeFrames
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
	c.Core.QueueAttack(ai, combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 3.5), hitmark, hitmark)
	return action.Info{
		Frames:          frames.NewAbilFunc(set),
		AnimationLength: set[action.InvalidAction],
		CanQueueAfter:   hitmark,
		State:           action.PlungeAttackState,
	}
}
