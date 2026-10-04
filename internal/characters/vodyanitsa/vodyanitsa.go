package vodyanitsa

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

type char struct {
	*tmpl.Character
	skillSrc   int
	lead       int
	chorus     int
	wandering  bool
	alterUntil int
	c4Expiry   []int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)

	c.EnergyMax = burstEnergy
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5

	w.Character = &c

	return nil
}

func (c *char) Init() error {
	c.a1Init()
	c.a4Init()
	c.c1Init()
	c.c2Init()
	c.c4Init()
	c.c6Init()
	return nil
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "lead":
		return c.lead, nil
	case "chorus":
		return c.chorus, nil
	case "c4":
		return c.c4Count(), nil
	case "song":
		return c.StatusIsActive(songKey), nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) songDuration() int {
	// Skill upgrade writes 16s onto a 0 default. C2 paramDelta adds 9s.
	dur := songDur
	if c.Base.Cons >= 2 {
		dur += c2SongExtend
	}
	return dur
}

func (c *char) songActive() bool {
	return c.StatusIsActive(songKey)
}

func (c *char) alterActive() bool {
	return c.wandering || c.Core.F < c.alterUntil
}
