package wriothesley

import (
	"slices"

	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	c1Status         = "wriothesley-c1"
	c1ICD            = 2.5 * 60
	c1ICDKey         = "wriothesley-c1-icd"
	c1SkillExtension = 4 * 60
	c4Status         = "wriothesley-c4-spd"
)

func (c *char) c1Ready() bool {
	return c.c1N5Proc || (c.CurrentHPRatio() < 0.6 && !c.StatusIsActive(c1ICDKey))
}

// The Gracious Rebuke from the Passive Talent "There Shall Be a Plea for Justice" is changed to this:
// When Wriothesley's HP is less than 60% or while he is in the Chilling Penalty state caused by Icefang Rush,
// when the fifth attack of Repelling Fists hits, it will create a Gracious Rebuke.
// 1 Gracious Rebuke effect can be obtained every 2.5s.
// Additionally, Rebuke: Vaulting Fist will obtain the following enhancement:
// - The DMG Bonus gained will be further increased to 200%.
// - When it hits while Wriothesley is in the Chilling Penalty state, that state's duration is extended by 4s. 1 such extension can occur per 1 Chilling Penalty duration.
// You must first unlock the Passive Talent "There Shall Be a Plea for Justice."
func (c *char) c1(ai *info.AttackInfo, snap *info.Snapshot) (info.AttackCBFunc, bool) {
	if !c.c1Ready() {
		return nil, false
	}
	c.c1N5Proc = false

	// add status that is removed on consumption
	c.AddStatus(c1Status, -1, false)

	// adjust ai
	ai.Abil = "Rebuke: Vaulting Fist"
	ai.HitlagFactor = 0.03
	ai.HitlagHaltFrames = 0.12 * 60

	// 200% increased DMG
	dmg := 2.0
	snap.Stats[attributes.DmgP] += dmg
	c.Core.Log.NewEvent("adding c1", glog.LogCharacterEvent, c.Index()).Write("dmg%", dmg)

	// add c6 if applicable
	c.addC6Buff(snap)

	// return callback to heal, extend E, remove C1 and apply 2.5s cd
	return func(a info.AttackCB) {
		if a.Target.Type() != info.TargettableEnemy {
			return
		}
		// do not proc if C1 not active
		if !c.StatusIsActive(c1Status) {
			return
		}
		// remove C1 and apply CD
		c.DeleteStatus(c1Status)
		c.AddStatus(c1ICDKey, c1ICD, true)

		c.c1SkillExtension()

		// heal
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index(),
			Target:  c.Index(),
			Message: a1HealName,
			Src:     c.caHeal * c.MaxHP(),
			Bonus:   c.Stat(attributes.Heal),
		})
	}, c.Base.Cons >= 6
}

func (c *char) makeC1N5CB(naCounter int) info.AttackCBFunc {
	if c.Base.Cons < 1 || naCounter != 4 {
		return nil
	}
	return func(a info.AttackCB) {
		if a.Target.Type() != info.TargettableEnemy {
			return
		}
		if c.revelationConduct() && c.skillBuffActive() {
			c.DeleteReactBonusMod(c1N5Bonus)
			c.addStellarHitBonus(c1LusterBonus, attacks.AttackTagWrioSSCCA)
			return
		}
		if !c.StatusIsActive(skillKey) || c.StatusIsActive(c1ICDKey) {
			return
		}
		c.c1N5Proc = true
		c.Core.Log.NewEvent("gained Gracious Rebuke from C1 N5", glog.LogCharacterEvent, c.Index())
	}
}

const (
	c1LusterBonus = c1Status + "-revelation-ca"
	c1N5Bonus     = c1Status + "-revelation-n5"
)

func (c *char) addStellarHitBonus(key string, tag attacks.AttackTag) {
	c.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBaseWithHitlag(key, 5*60),
		Amount: func(ai info.AttackInfo) float64 {
			if slices.Contains(ai.AdditionalTags, tag) {
				return 0.5
			}
			return 0
		},
	})
}

func (c *char) c1SkillExtension() {
	if !c.c1SkillExtensionProc && c.StatusIsActive(skillKey) {
		c.ExtendStatus(skillKey, c1SkillExtension)
		c.c1SkillExtensionProc = true
		c.Core.Log.NewEvent("c1: skill duration is extended", glog.LogCharacterEvent, c.Index())
	}
}

func (c *char) c1LusterCB(a info.AttackCB) {
	if c.Base.Cons < 1 || a.Target.Type() != info.TargettableEnemy || !c.revelationConduct() {
		return
	}
	c.DeleteReactBonusMod(c1LusterBonus)
	c.addStellarHitBonus(c1N5Bonus, attacks.AttackTagWrioSSCN5)
	c.c1SkillExtension()
}

func (c *char) resetC1SkillExtension() {
	if c.Base.Cons < 1 {
		return
	}
	c.c1SkillExtensionProc = false
}

// When using Darkgold Wolfbite, each Prosecution Edict stack from the Passive Talent
// "There Shall Be a Reckoning for Sin" will increase said ability's DMG dealt by 40%.
// You must first unlock the Passive Talent "There Shall Be a Reckoning for Sin."
func (c *char) c2(snap *info.Snapshot) {
	if c.Base.Ascension < 4 {
		return
	}
	if c.Base.Cons < 2 {
		return
	}
	if !c.StatusIsActive(skillKey) {
		return
	}

	dmg := 0.4 * float64(c.a4Stack)
	snap.Stats[attributes.DmgP] += dmg
	c.Core.Log.NewEvent("adding c2", glog.LogCharacterEvent, c.Index()).Write("dmg%", dmg)
}

func (c *char) c2ConductReady() bool {
	return c.Base.Cons >= 2 && c.revelation && c.Base.Ascension >= 4 && c.a4Stack >= 5 && c.StatusIsActive(skillKey)
}

func (c *char) c2ConductFraction() float64 {
	if !c.c2ConductReady() || !c.inConductRadiance() {
		return 0
	}
	return 0.5
}

func (c *char) c2OutsideConductNA() float64 {
	if !c.c2ConductReady() || c.inConductRadiance() {
		return 0
	}
	return 0.25
}

func (c *char) c2OutsideConductCA() float64 {
	if !c.c2ConductReady() || c.inConductRadiance() {
		return 0
	}
	return 0.3
}

// The HP restored to Wriothesley through Rebuke: Vaulting Fist will be increased to 50%
// of his Max HP. You must first unlock the Passive Talent "There Shall Be a Plea for Justice."
// Additionally, when Wriothesley is healed, if the amount of healing overflows, the following
// effects will occur depending on whether his is on the field or not. If he is on the field,
// his ATK SPD will be increased by 20% for 4s. If he is off-field, all party members' ATK SPD
// will be increased by 10% for 6s. These two ATK SPD increasing methods cannot stack.
func (c *char) c4() {
	if c.Base.Cons < 4 {
		c.caHeal = 0.3
		return
	}
	c.caHeal = 0.5

	m := make([]float64, attributes.EndStatType)
	c.Core.Events.Subscribe(event.OnHeal, func(args ...any) {
		if c.revelationConduct() {
			return
		}
		index := args[1].(int)
		amount := args[2].(float64)
		overheal := args[3].(float64)
		if index != c.Index() || amount <= 0 || overheal <= 0 {
			return
		}

		chars := c.Core.Player.Chars()
		for _, char := range chars {
			char.DeleteStatus(c4Status)
		}

		if c.Core.Player.Active() == c.Index() {
			c.AddStatMod(character.StatMod{
				Base:         modifier.NewBaseWithHitlag(c4Status, 4*60),
				AffectedStat: attributes.AtkSpd,
				Amount: func() []float64 {
					m[attributes.AtkSpd] = 0.2
					return m
				},
			})
		} else {
			for _, char := range chars {
				char.AddStatMod(character.StatMod{
					Base:         modifier.NewBaseWithHitlag(c4Status, 6*60),
					AffectedStat: attributes.AtkSpd,
					Amount: func() []float64 {
						m[attributes.AtkSpd] = 0.1
						return m
					},
				})
			}
		}
	}, "wriothesley-c4-heal")

	if !c.revelation {
		return
	}
	for _, ch := range c.Core.Player.Chars() {
		amt := 0.1
		if ch.Index() == c.Index() {
			amt = 0.2
		}
		spd := make([]float64, attributes.EndStatType)
		spd[attributes.AtkSpd] = amt
		this := ch
		this.AddStatMod(character.StatMod{
			Base:         modifier.NewBase(c4Status+"-revelation", -1),
			AffectedStat: attributes.AtkSpd,
			Amount: func() []float64 {
				if this.StatusIsActive(c4Status) || !c.inConductRadiance() {
					return nil
				}
				return spd
			},
		})
	}
}

// The CRIT Rate of Rebuke: Vaulting Fist will be increased by 10%, and its CRIT DMG by 80%.
func (c *char) addC6Buff(snap *info.Snapshot) {
	if c.Base.Cons < 6 {
		return
	}
	cr := 0.1
	cd := 0.8
	snap.Stats[attributes.CR] += cr
	snap.Stats[attributes.CD] += cd
	c.Core.Log.NewEvent("adding c6", glog.LogCharacterEvent, c.Index()).Write("cr", cr).Write("cd", cd)
}
