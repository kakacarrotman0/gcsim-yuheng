package kachina

import (
	"github.com/genshinsim/gcsim/internal/template/character"
	nightsoultemplate "github.com/genshinsim/gcsim/internal/template/nightsoul"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	playercharacter "github.com/genshinsim/gcsim/pkg/core/player/character"
)

const (
	twirlyKey     = "kachina-turbo-twirly"
	fieldKey      = "kachina-turbo-drill-field"
	particleICD   = "kachina-particle-icd"
	nightsoulMax  = 60
	slamCost      = 10.0
	dismountCost  = 2.0
	pointEpsilon  = 0.2
	firstSlam     = 132 // 2.0s think + 0.2s attack delay
	slamInterval  = 120 // 2.0s
	independentR  = 4.0
	fieldSlamR    = 5.2
	burstHitR     = 6.0
	skillCD       = 20 * 60
	fieldDuration = 12 * 60
	burstCD       = 18 * 60
	burstHitmark  = 35
	a4Ratio       = 0.2
)

type char struct {
	*character.Character
	nightsoulState *nightsoultemplate.State
	twirlySrc      int
	mounted        bool
	fieldCenter    info.Point
	c6Seen         map[string]bool
	twirlyCon      *twirlyConstruct
	constructSeq   int
}

func NewChar(s *core.Core, w *playercharacter.CharWrapper, _ info.CharacterProfile) error {
	c := char{
		Character: character.NewWithWrapper(s, w),
		twirlySrc: -1,
		c6Seen:    make(map[string]bool),
	}
	c.nightsoulState = nightsoultemplate.New(c.Core, c.CharWrapper)
	c.nightsoulState.MaxPoints = nightsoulMax
	c.EnergyMax = 70
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5
	w.Character = &c
	return nil
}

func (c *char) Init() error {
	c.initA1()
	c.initC1()
	c.initC4()
	c.initC6()
	return nil
}

func (c *char) Condition(fields []string) (any, error) {
	if len(fields) > 0 && fields[0] == "nightsoul" {
		return c.nightsoulState.Condition(fields)
	}
	if len(fields) > 0 && fields[0] == "mounted" {
		return c.mounted, nil
	}
	return c.Character.Condition(fields)
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && c.StatusIsActive(twirlyKey) {
		return true, action.NoFailure
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	if c.mounted {
		return 0
	}
	return c.Character.ActionStam(a, p)
}

func (c *char) startTwirly(points float64, mounted bool) {
	if !c.nightsoulState.HasBlessing() {
		c.nightsoulState.EnterBlessing(points)
	} else if points > 0 {
		c.nightsoulState.GeneratePoints(points)
	}
	c.AddStatus(twirlyKey, -1, true)
	c.mounted = mounted
	c.twirlySrc = c.Core.F
	c.spawnConstruct()
	if !mounted {
		c.queueIndependent(c.twirlySrc, firstSlam)
	}
}

func (c *char) killTwirlyGadget() {
	c.DeleteStatus(twirlyKey)
	c.mounted = false
	c.twirlySrc = -1
	c.despawnConstruct()
}

func (c *char) endTwirly() {
	c.killTwirlyGadget()
	if c.nightsoulState.HasBlessing() {
		c.nightsoulState.ExitBlessing()
	}
	c.nightsoulState.ClearPoints()
}

func (c *char) consumeTwirlyPoints(amount float64) {
	if !c.nightsoulState.HasBlessing() {
		return
	}
	c.nightsoulState.ConsumePoints(amount)
	if c.nightsoulState.Points() <= pointEpsilon {
		c.endTwirly()
	}
}

func (c *char) canPaySlam() bool {
	return c.nightsoulState.HasBlessing() && c.nightsoulState.Points() > pointEpsilon
}

func (c *char) twirlyAttackInfo(abil string, mult float64, poise float64) info.AttackInfo {
	ai := info.AttackInfo{
		ActorIndex:     c.Index(),
		Abil:           abil,
		AttackTag:      attacks.AttackTagElementalArt,
		AdditionalTags: []attacks.AttackTag{attacks.AttackTagNightsoul},
		ICDTag:         attacks.ICDTagElementalArt,
		ICDGroup:       attacks.ICDGroupDefault,
		StrikeType:     attacks.StrikeTypeBlunt,
		PoiseDMG:       poise,
		Element:        attributes.Geo,
		Durability:     25,
		Mult:           mult,
		UseDef:         true,
		IgnoreInfusion: true,
	}
	if c.Base.Ascension >= 4 {
		ai.FlatDmg += c.TotalDef(false) * a4Ratio
	}
	return ai
}

func (c *char) twirlyRadius() float64 {
	if c.StatusIsActive(fieldKey) {
		return fieldSlamR
	}
	return independentR
}
