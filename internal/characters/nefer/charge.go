package nefer

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

var chargeFrames []int
var ppFrames []int

func init() {
	chargeFrames = frames.InitAbilSlice(caAnim)
	chargeFrames[action.ActionAttack] = caHit
	chargeFrames[action.ActionCharge] = caHit
	chargeFrames[action.ActionSkill] = caHit
	chargeFrames[action.ActionBurst] = caHit
	chargeFrames[action.ActionDash] = caHit
	chargeFrames[action.ActionJump] = caHit
	chargeFrames[action.ActionSwap] = caHit

	ppFrames = frames.InitAbilSlice(ppAnim)
	ppFrames[action.ActionAttack] = shade1Frame
	ppFrames[action.ActionCharge] = shade1Frame
	ppFrames[action.ActionSkill] = shade1Frame
	ppFrames[action.ActionBurst] = shade1Frame
	ppFrames[action.ActionDash] = shade1Frame
	ppFrames[action.ActionJump] = shade1Frame
	ppFrames[action.ActionSwap] = shade1Frame
}

func (c *char) ChargeAttack(_ map[string]int) (action.Info, error) {
	if c.StatusIsActive(shadowKey) && c.Core.Player.Dew() > 0 {
		return c.phantasm(), nil
	}
	c.absorbSeeds()
	ai := c.hybrid("Charged Attack", attacks.AttackTagExtra, attacks.ICDTagNone, 25, charged[c.TalentLvlAttack()], 0)
	c.queueCircle(ai, caHit)
	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   caHit,
		State:           action.ChargeAttackState,
	}, nil
}

func (c *char) phantasm() action.Info {
	c.absorbSeeds()
	c.ppCount++
	lvl := c.TalentLvlSkill()
	c.queueCircle(c.withVeil(c.hybrid("Phantasm Performance (Nefer 1)", attacks.AttackTagExtra, attacks.ICDTagExtraAttack, 25, ppN1ATK[lvl], ppN1EM[lvl])), ppN1Frame)
	if c.Base.Cons >= 6 {
		c.queueLunar("Phantasm Performance C6 (Nefer 2)", 0.85, ppN2Frame)
		c.queueLunar("Phantasm Performance C6 (Extra)", 1.2, ppExtra)
	} else {
		c.queueCircle(c.withVeil(c.hybrid("Phantasm Performance (Nefer 2)", attacks.AttackTagExtra, attacks.ICDTagExtraAttack, 25, ppN2ATK[lvl], ppN2EM[lvl])), ppN2Frame)
	}
	c.queueLunar("Phantasm Performance (Shade 1)", shade1[lvl], shade1Frame)
	c.queueLunar("Phantasm Performance (Shade 2)", shade2[lvl], shade2Frame)
	c.queueLunar("Phantasm Performance (Shade 3)", shade3[lvl], shade3Frame)
	c.QueueCharTask(func() { c.Core.Player.ConsumeDew(1) }, shade1Frame)
	if c.ppCount >= ppMax {
		src := c.danceSrc
		c.QueueCharTask(func() {
			if c.danceSrc != src {
				return
			}
			c.exitDance()
		}, ppAnim)
	}
	return action.Info{
		Frames:          frames.NewAbilFunc(ppFrames),
		AnimationLength: ppFrames[action.InvalidAction],
		CanQueueAfter:   shade1Frame,
		State:           action.ChargeAttackState,
	}
}

func (c *char) hybrid(name string, tag attacks.AttackTag, icd attacks.ICDTag, dur info.Durability, atk, em float64) info.AttackInfo {
	return info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       name,
		AttackTag:  tag,
		ICDTag:     icd,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Dendro,
		Durability: dur,
		Mult:       atk,
		FlatDmg:    em * c.Stat(attributes.EM),
	}
}

func (c *char) queueLunar(name string, ratio float64, delay int) {
	ai := info.AttackInfo{
		ActorIndex:       c.Index(),
		Abil:             name,
		AttackTag:        attacks.AttackTagDirectLunarBloom,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeDefault,
		Element:          attributes.Dendro,
		UseEM:            true,
		IgnoreDefPercent: 1,
		Mult:             ratio,
	}
	if c.Base.Cons >= 1 {
		ai.Mult += 0.6
	}
	if c.Base.Cons >= 6 && c.gleam {
		ai.Elevation = 0.15
	}
	c.queueCircle(c.withVeil(ai), delay)
}

// withVeil applies Veil of Falsehood as a Base DMG Multiplier on Phantasm Performance.
// Direct Lunar-Bloom multiplies Mult*EM by (1+BaseDmgBonus) and ignores DMG Bonus.
// An ordinary hit adds FlatDmg after that product, so the EM flat is scaled here
// and the 8% is not also written into DMG Bonus.
func (c *char) withVeil(ai info.AttackInfo) info.AttackInfo {
	bonus := 0.08 * float64(c.veilCount())
	if bonus == 0 {
		return ai
	}
	ai.BaseDmgBonus += bonus
	if ai.FlatDmg != 0 {
		ai.FlatDmg *= 1 + bonus
	}
	return ai
}
