package lohen

import (
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

func (c *char) a1Init() {
	if c.Base.Ascension < 1 {
		return
	}
	// Teammate_Attack_Monitor is attached to every avatar except Lohen while
	// Masterstroke's think is running. Distance 40 is not simulated.
	c.Core.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		if !c.masterActive() {
			return
		}
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex == c.Index() {
			return
		}
		dmg := args[2].(float64)
		base := c.Stat(attributes.BaseATK)
		lvl := c.skillLvl()
		eff := c.willEfficiency()
		if dmg >= base*willAtkRatio[lvl] {
			c.addWill(willSuccess[lvl] * eff)
		} else if dmg >= 0 {
			c.addWill(willFailGain * eff)
		}
		if c.Base.Ascension >= 1 && dmg >= base*a1ExtraRatio {
			c.addWill(a1ExtraWill * eff)
		}
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
	// PermanentSkill_2 lists Melt, Freeze, Superconductor, Cryo Swirl, and Cryo Crystallize.
	// Stellar-Swirl and Stellar-Superconductor are in that list and are not wired:
	// the engine events do not identify the ice variant.
	for _, evt := range []event.Event{
		event.OnMelt,
		event.OnFrozen,
		event.OnSuperconduct,
		event.OnSwirlCryo,
		event.OnCrystallizeCryo,
	} {
		c.Core.Events.Subscribe(evt, hook, "lohen-a4")
	}
}

func (c *char) highSpirits() {
	if c.Base.Ascension < 1 || c.Core.F < c.spiritsICDUntil {
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
			case attacksTagNormal(), attacksTagExtra():
				return m
			default:
				return nil
			}
		},
	})
}
