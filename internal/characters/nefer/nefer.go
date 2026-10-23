package nefer

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

const (
	shadowKey = "nefer-shadow-dance"
	deceitKey = "nefer-deceit"
	shadowDur = 9 * 60
	deceitDur = 15 * 60
	skillCD   = 9 * 60
	burstCD   = 15 * 60
	veilDur   = 9 * 60
	veilDurC2 = 14 * 60
	veilEMDur = 8 * 60
	c4Linger  = 270 // 4.5s
	ppMax     = 3
	hitRadius = 4.0

	// Unmeasured. No frame sheet was found for Nefer.
	skillHit    = 20
	skillAnim   = 45
	burstHit1   = 30
	burstHit2   = 52
	burstAnim   = 70
	caHit       = 40
	caAnim      = 70
	ppN1Frame   = 30
	ppN2Frame   = 45
	shade1Frame = 48
	shade2Frame = 60
	shade3Frame = 72
	ppExtra     = 84
	ppAnim      = 96
	n3Split     = 2
)

type char struct {
	*tmpl.Character
	gleam      bool
	danceSrc   int
	ppCount    int
	veils      []int
	burstBonus float64
	c4Shred    bool
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)
	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5
	c.Moonsign = 1
	c.danceSrc = -1
	c.SetNumCharges(action.ActionSkill, 2)
	w.Character = &c
	return nil
}

func (c *char) Init() error {
	c.gleam = c.Core.Player.GetMoonsignLevel() >= 2
	c.moonsignInit()
	c.a1Init()
	c.buffInit()
	c.c4Init()
	return nil
}

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	if a != action.ActionCharge {
		return c.Character.ActionStam(a, p)
	}
	if c.StatusIsActive(shadowKey) && c.Core.Player.Dew() > 0 {
		return 0
	}
	if c.StatusIsActive(shadowKey) {
		return 25
	}
	return 50
}

func (c *char) AnimationStartDelay(k info.AnimationDelayKey) int {
	return c.Character.AnimationStartDelay(k)
}
