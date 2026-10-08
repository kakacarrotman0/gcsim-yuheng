package alyosha

import (
	"github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	playercharacter "github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

const (
	fieldKey          = "alyosha-hunting-field"
	particleICDKey    = "alyosha-particle-icd"
	c1ICDKey          = "alyosha-c1-icd"
	markDuration      = 15 * 60
	precisionDuration = 15 * 60
	skillCD           = 15 * 60
	burstCD           = 18 * 60
	burstDuration     = 14 * 60
	c2ExtraDuration   = 6 * 60
	fieldInterval     = 120
	dogDelay          = 18
	fieldRadius       = 6.0
	dogRadius         = 1.0
	huntRadius        = 15.0
	a1HealRatio       = 1.2
	c4HealRatio       = 0.6
	c1Energy          = 15.0
	c1Interval        = 18 * 60
	c6EM              = 100.0
	a4PerER           = 0.35
	a4Cap             = 0.70
	stellarBonusValue = 0.20
	particleCount     = 5.0
	particleICD       = 30
	skillTapHitmark   = 24
	skillHoldHitmark  = 48
	burstHitmark      = 36
)

type char struct {
	*character.Character
	markUntil       map[info.TargetKey]int
	fieldSrc        int
	precisionStacks int
	precisionUntil  int
	stellarBonus    float64
}

func NewChar(s *core.Core, w *playercharacter.CharWrapper, _ info.CharacterProfile) error {
	c := char{
		Character: character.NewWithWrapper(s, w),
		markUntil: make(map[info.TargetKey]int),
		fieldSrc:  -1,
	}
	c.EnergyMax = 70
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5
	w.Character = &c
	return nil
}

func (c *char) Init() error {
	c.initA4()
	c.initPrecision()
	c.initC1()
	return nil
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "marks":
		n := 0
		for _, until := range c.markUntil {
			if until > c.Core.F {
				n++
			}
		}
		return n, nil
	case "precision":
		if c.Core.F >= c.precisionUntil {
			return 0, nil
		}
		return c.precisionStacks, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) marked(key info.TargetKey) bool {
	return c.markUntil[key] > c.Core.F
}

func (c *char) markHit(t info.Target, activate, apply bool) {
	if t == nil || t.Type() != info.TargettableEnemy {
		return
	}
	key := t.Key()
	if c.marked(key) && activate {
		delete(c.markUntil, key)
		c.grantPrecision()
		return
	}
	if apply {
		c.markUntil[key] = c.Core.F + markDuration
	}
}

func (c *char) grantPrecision() {
	if c.Base.Cons >= 6 && c.precisionStacks >= 1 && c.Core.F < c.precisionUntil {
		c.precisionStacks = 2
	} else {
		c.precisionStacks = 1
	}
	c.precisionUntil = c.Core.F + precisionDuration
	if c.StatusIsActive(reactable.PolestarFieldKey) {
		// Snapshotted at activation. C6's second stack doubles the 20%.
		c.stellarBonus = stellarBonusValue * float64(c.precisionStacks)
	} else {
		c.stellarBonus = 0
	}
}

func (c *char) initPrecision() {
	for _, other := range c.Core.Player.Chars() {
		other := other
		other.AddStatMod(playercharacter.StatMod{
			Base:         modifier.NewBase("alyosha-precision", -1),
			AffectedStat: attributes.NoStat,
			Amount: func() []float64 {
				if c.Core.Player.Active() != other.Index() || c.Core.F >= c.precisionUntil || c.precisionStacks == 0 {
					return nil
				}
				m := make([]float64, attributes.EndStatType)
				m[attributes.ATKP] = precisionATK[c.TalentLvlSkill()] * float64(c.precisionStacks)
				if c.precisionStacks >= 2 {
					m[attributes.EM] = c6EM
				}
				return m
			},
		})
		other.AddReactBonusMod(playercharacter.ReactBonusMod{
			Base: modifier.NewBase("alyosha-precision-stellar", -1),
			Amount: func(ai info.AttackInfo) float64 {
				return c.stellarBonusFor(other.Index(), ai.AttackTag)
			},
		})
	}
}

func (c *char) stellarBonusFor(charIndex int, tag attacks.AttackTag) float64 {
	if c.Core.F >= c.precisionUntil || c.precisionStacks == 0 {
		return 0
	}
	if c.Core.Player.Active() != charIndex {
		return 0
	}
	if tag != attacks.AttackTagDirectStellarConduct {
		return 0
	}
	return c.stellarBonus
}

func (c *char) initA4() {
	if c.Base.Ascension < 4 {
		return
	}
	c.AddAttackMod(playercharacter.AttackMod{
		Base: modifier.NewBase("alyosha-a4", -1),
		Amount: func(atk *info.AttackEvent, _ info.Target) []float64 {
			if atk.Info.ActorIndex != c.Index() {
				return nil
			}
			switch atk.Info.AttackTag {
			case attacks.AttackTagElementalArt, attacks.AttackTagElementalBurst:
			default:
				return nil
			}
			bonus := c.Stat(attributes.ER) * a4PerER
			if bonus > a4Cap {
				bonus = a4Cap
			}
			m := make([]float64, attributes.EndStatType)
			m[attributes.DmgP] = bonus
			return m
		},
	})
}

func (c *char) fieldDuration() int {
	if c.Base.Cons >= 2 {
		return burstDuration + c2ExtraDuration
	}
	return burstDuration
}
