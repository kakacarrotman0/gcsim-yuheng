package lohen

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) willInit() {
	// Teammate_Attack_Monitor is part of the skill, not Moratorium on
	// Questioning. The extra 60 Will is that ascension-1 passive.
	// The monitor is attached to each other avatar within 40m. gcsim has one
	// player position, so that distance cannot fail. See docs/yuheng/lohen.md.
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		if !c.masterActive() {
			return
		}
		atk := args[1].(*info.AttackEvent)
		src := atk.Info.ActorIndex
		if src == c.Index() {
			return
		}
		dmg := args[2].(float64)
		base := c.Stat(attributes.BaseATK)
		lvl := c.skillLvl()
		eff := c.willEfficiency()
		gate := c.willGate[src]
		// Success_Pre, Fail_Pre, and Success_Extra_Pre are separate unique 0.1s
		// modifiers on the attacker, so one character's gate does not block another.
		if dmg >= base*willAtkRatio[lvl] {
			if c.Core.F >= gate.success {
				c.addWill(willSuccess[lvl] * eff)
				gate.success = c.Core.F + willWindowF
			}
		} else if dmg >= 0 {
			if c.Core.F >= gate.fail {
				c.addWill(willFailGain * eff)
				gate.fail = c.Core.F + willWindowF
			}
		}
		if c.Base.Ascension >= 1 && dmg >= base*a1ExtraRatio && c.Core.F >= gate.extra {
			c.addWill(a1ExtraWill * eff)
			gate.extra = c.Core.F + willWindowF
		}
		c.willGate[src] = gate
	}, "lohen-will")
}

func (c *char) a4Init() {
	if c.Base.Ascension < 4 {
		return
	}
	hook := func(args ...any) {
		if !c.masterActive() || c.Base.Ascension < 4 {
			return
		}
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex == c.Index() {
			return
		}
		m := make([]float64, attributes.EndStatType)
		m[attributes.ATKP] = a4ATK
		src := c.Core.Player.ByIndex(atk.Info.ActorIndex)
		src.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(a4Key, a4Duration),
			AffectedStat: attributes.ATKP,
			Amount: func() []float64 {
				return m
			},
		})
		c.AddStatMod(character.StatMod{
			Base:         modifier.NewBaseWithHitlag(a4Key, a4Duration),
			AffectedStat: attributes.ATKP,
			Amount: func() []float64 {
				return m
			},
		})
	}
	// PermanentSkill_2 lists Melt, Freeze, Superconductor, SwirlIce,
	// StarSwirlIce, CrystallizeIce, and StarSuperconductor.
	// OnStellarSwirl is only emitted for Cryo or Frozen. OnStellarConduct is
	// the Cryo and Electro Stellar reaction.
	for _, evt := range []event.Event{
		event.OnMelt,
		event.OnFrozen,
		event.OnSuperconduct,
		event.OnSwirlCryo,
		event.OnStellarSwirl,
		event.OnCrystallizeCryo,
		event.OnStellarConduct,
	} {
		c.Core.Events.Subscribe(evt, hook, "lohen-a4")
	}
}

func (c *char) highSpirits() {
	// When the Mood Strikes is proud skill 12923. AvatarSkillDepot 12901
	// omits its promote level, while 12921 is promote 1 and 12922 is promote
	// 4. An omitted level is 0, the same encoding as an always-unlocked
	// utility passive, so High Spirits is not ascension-gated.
	if c.Core.F < c.spiritsICDUntil {
		return
	}
	c.spiritsICDUntil = c.Core.F + spiritsICD
	dur := spiritsBase
	skillLevel := c.TalentLvlSkill() + 1
	for _, char := range c.Core.Player.Chars() {
		if char.Index() == c.Index() {
			continue
		}
		if char.TalentLvlAttack()+1 >= skillLevel || char.TalentLvlSkill()+1 >= skillLevel || char.TalentLvlBurst()+1 >= skillLevel {
			dur += spiritsExtra
			break
		}
	}
	c.AddStatus(spiritsKey, dur, true)
}

func (c *char) tryHexerei(snapped float64) {
	if !c.IsHexerei || c.Core.Player.GetHexereiCount() < 2 {
		return
	}
	if snapped < hexThreshold*c.willMax() {
		return
	}
	m := make([]float64, attributes.EndStatType)
	m[attributes.DmgP] = hexBonus
	c.AddAttackMod(character.AttackMod{
		Base: modifier.NewBaseWithHitlag(hexKey, hexDuration),
		Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			switch atk.Info.AttackTag {
			case attacksTagNormal():
				// NormalAttack_06 is the second hit of the fifth normal. The
				// Hexenzirkel mixin lists NormalAttack_01 through _05 and ExtraAttack.
				if len(atk.Info.Abil) >= 2 && atk.Info.Abil[len(atk.Info.Abil)-2:] == " B" {
					return nil
				}
				return m
			case attacksTagExtra():
				return m
			default:
				return nil
			}
		},
	})
}
