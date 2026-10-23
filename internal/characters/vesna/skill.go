package vesna

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// Unverified cancel windows. The entry hit, Lv1 slash, and Lv2 slash are
// placeholders. The 30-frame gap from the Lv2 slash to its spirit blade is
// the gadget think interval of 0.5s. Lv3 emits four hits plus a final because
// the talent row says ×4; those five frames are not in the ability JSON.
var skillFrames []int

const (
	skillHitmark  = 14
	lv1Hitmark    = 12
	lv2Slash      = 12
	lv2Blade      = 42
	pinionHitmark = 12
	skillRadius   = 5.0
	bladeRadius   = 3.0
	pinionRadius  = 0.6
)

var lv3Hitmarks = []int{14, 24, 34, 44, 58}

func init() {
	skillFrames = frames.InitAbilSlice(windborneEnd)
	skillFrames[action.ActionAttack] = 20
	skillFrames[action.ActionCharge] = 20
	skillFrames[action.ActionSkill] = 20
	skillFrames[action.ActionBurst] = 20
	skillFrames[action.ActionDash] = 18
	skillFrames[action.ActionSwap] = 24
}

func framesAbil() func(action.Action) int {
	return frames.NewAbilFunc(skillFrames)
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	if c.transposeReady() {
		c.transpose()
		return c.skillInfo(), nil
	}
	if c.armed() {
		if !c.canWindborne() {
			return c.notReady("windborne sword is not available")
		}
		c.windborne()
		return c.skillInfo(), nil
	}
	if c.AvailableCDCharge[action.ActionSkill] <= 0 {
		return c.notReady("the art of victory is on cooldown")
	}
	c.enterArmed()
	c.artHit("The Art of Victory", skillDMG[c.TalentLvlSkill()], skillHitmark, skillRadius, attacks.StrikeTypeSlash)
	c.SetCD(action.ActionSkill, skillCD)
	return c.skillInfo(), nil
}

func (c *char) windborne() {
	lvl := c.TalentLvlSkill()
	cast := c.wbLevel
	if cast < 1 {
		cast = 1
	}
	// Cost and the level-up are on the attack-effect onAdded, at the start of the state.
	if c.c1Free && cast >= wbLevelMax {
		c.c1Free = false
	} else {
		c.addEssence(-essenceCost)
	}
	if c.wbLevel < wbLevelMax {
		c.wbLevel++
	}
	if cast >= wbLevelMax {
		c.e3Uses++
		// Further Windborne is blocked as soon as the capping use is cast.
		// Armed itself ends at the placeholder animation length, so a cancel
		// into Transpose can still see the mode and summon a pinion.
		if c.e3Uses >= c.e3Limit {
			c.closing = true
			c.scheduleArmedExit(windborneEnd)
		}
		c.openTranspose()
	}
	c.addDisc()

	switch cast {
	case 1:
		c.artHit("Windborne Sword Lv1", skillDMG[lvl], lv1Hitmark, skillRadius, attacks.StrikeTypePierce)
	case 2:
		c.artHit("Windborne Sword Lv2", wbLv2[lvl], lv2Slash, skillRadius, attacks.StrikeTypeSlash)
		c.queueSpirit("Spirit Blade Lv2", wbLv2Blade[lvl], lv2Blade, false, bladeRadius)
	default:
		for i := range 4 {
			c.queueSpirit("Spirit Blade Lv3", wbLv3[lvl], lv3Hitmarks[i], false, bladeRadius)
		}
		c.queueSpirit("Spirit Blade Lv3 Final", wbLv3Final[lvl], lv3Hitmarks[4], false, bladeRadius)
	}
}

func (c *char) artHit(name string, mult float64, delay int, radius float64, strike attacks.StrikeType) {
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       name,
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: strike,
		Element:    attributes.Anemo,
		Durability: 25,
		Mult:       mult,
	}
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, radius)
	c.Core.Tasks.Add(func() {
		c.Core.QueueAttack(ai, ap, 0, 0, c.tryParticle)
	}, delay)
}

// queueSpirit builds the hit when it lands. Disciplinary stacks and Radiance
// are read then, which is when the gadget's damage modifier fires.
func (c *char) queueSpirit(name string, ratio float64, delay int, burst bool, radius float64) {
	c.Core.Tasks.Add(func() {
		ai := c.spiritInfo(name, ratio, burst)
		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, radius)
		c.Core.QueueAttack(ai, ap, 0, 0, c.tryParticle)
	}, delay)
}

func (c *char) spiritInfo(name string, ratio float64, burst bool) info.AttackInfo {
	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       name,
		Element:    attributes.Anemo,
		Mult:       ratio * c.discFactor(),
		StrikeType: attacks.StrikeTypeSlash,
	}
	// During Radiance the gadget uses a reaction attack: no gauge, no DEF,
	// Direct Stellar Swirl. Outside Radiance it is an ordinary Anemo talent hit.
	// The talent sheet uses the same percentage for both rows.
	if c.inRadiance() {
		ai.AttackTag = attacks.AttackTagDirectStellarSwirl
		ai.ICDTag = attacks.ICDTagNone
		ai.ICDGroup = attacks.ICDGroupDefault
		ai.Durability = 0
		ai.IgnoreDefPercent = 1
		if c.Base.Cons >= 6 {
			ai.Elevation = c6Elevation
		}
		return ai
	}
	ai.Durability = 25
	ai.ICDGroup = attacks.ICDGroupDefault
	if burst {
		ai.AttackTag = attacks.AttackTagElementalBurst
		ai.ICDTag = attacks.ICDTagElementalBurst
		return ai
	}
	ai.AttackTag = attacks.AttackTagElementalArt
	ai.ICDTag = attacks.ICDTagElementalArt
	return ai
}

// summonPinion fires one small-sword bullet and adds 5 essence.
// The extracted small-sword hit is 1U Anemo. A star-mode bullet id exists, and
// the particle handler lists SmallSword_Star, but that bullet's attack event
// was not in the extracted handler. The pinion stays an ordinary Anemo hit
// in both states so an unverified tag is not turned into Direct Stellar Swirl.
func (c *char) summonPinion() {
	if !c.armed() {
		return
	}
	c.addEssence(essencePinion)
	lvl := c.TalentLvlSkill()
	c.Core.Tasks.Add(func() {
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Wind Pinion",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypePierce,
			Element:    attributes.Anemo,
			Durability: 25,
			Mult:       pinion[lvl],
		}
		ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, pinionRadius)
		c.Core.QueueAttack(ai, ap, 0, 0, c.tryParticle)
	}, pinionHitmark)
}
