package vodyanitsa

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var burstFrames []int

// APPROXIMATE. The hit is onRemoved of a modifier that covers normalizeEnd 0.95 of the burst state.
const burstHitmark = 76

func init() {
	burstFrames = frames.InitAbilSlice(80)
	burstFrames[action.ActionAttack] = burstHitmark
	burstFrames[action.ActionSkill] = burstHitmark
	burstFrames[action.ActionDash] = burstHitmark
	burstFrames[action.ActionSwap] = 70
}

func (c *char) Burst(p map[string]int) (action.Info, error) {
	c.QueueCharTask(func() {
		ratio := burst[c.TalentLvlBurst()]
		// Ability default DamageRatio is 1.0. The talent param is a paramDelta added to that default,
		// and the song path multiplies Max HP damage by DamageRatio. Without the song the ratio is omitted.
		if c.songActive() {
			ratio *= 1 + burstBonus[c.TalentLvlBurst()]
		}
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Koda: Sink With Thee",
			AttackTag:  attacks.AttackTagElementalBurst,
			ICDTag:     attacks.ICDTagNone,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Hydro,
			Durability: 50,
			FlatDmg:    ratio * c.MaxHP(),
			PoiseDMG:   150,
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, 5),
			0,
			0,
		)
	}, burstHitmark)

	c.SetCD(action.ActionBurst, burstCD)
	c.ConsumeEnergy(0)

	return action.Info{
		Frames:          frames.NewAbilFunc(burstFrames),
		AnimationLength: burstFrames[action.InvalidAction],
		CanQueueAfter:   burstFrames[action.ActionSwap],
		State:           action.BurstState,
	}, nil
}
