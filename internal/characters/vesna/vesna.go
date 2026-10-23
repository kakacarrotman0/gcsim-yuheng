package vesna

import (
	"fmt"

	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

const (
	armedKey     = "vesna-armed"
	transposeKey = "vesna-transpose"
	radianceKey  = "vesna-radiance"

	// Ability Avatar_Vesna_MagicPower. These are literals, not unbound specials.
	essenceSkill  = 60.0
	essenceBurst  = 30.0
	essenceMax    = 90.0
	essenceCost   = 30.0
	essencePinion = 5.0

	armedDuration = 15 * 60
	discDuration  = 20 * 60
	discMax       = 6
	discPerStack  = 0.10
	particleICD   = 150 // unique modifier 2.5s
	skillCD       = 18 * 60
	burstCD       = 15 * 60

	wbLevelMax = 3
	e3Base     = 3
	e3C1       = 4

	// Released talent text. The ability specials for these ratios are 0.
	c1React     = 0.20
	c2ATK       = 0.40
	c4Mult      = 3.0
	c6Elevation = 0.20
	c6Window    = 5 * 60
	a4ATK       = 0.06
	a4EM        = 25.0
	jubileePer  = 0.007
	jubileeCap  = 0.14

	transposeTalent = 1.50
	transposeBlade  = 2.00

	// Placeholder animation end for the Lv3 state. The E3 checker on
	// Vesna_ElementalArt_C is 1.5s; the absolute exit frame is not in the JSON.
	windborneEnd = 60
)

type char struct {
	*tmpl.Character
	essence        float64
	armedUntil     int
	armedGen       int
	wbLevel        int
	e3Uses         int
	e3Limit        int
	c1Free         bool
	closing        bool
	discUntil      []int
	transposeUntil int
	particleUntil  int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)
	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5
	w.Character = &c
	return nil
}

func (c *char) Init() error {
	c.stellarInit()
	c.a4Init()
	c.c1Init()
	c.c2Init()
	// UNIQUE_Vesna_ElementalArt_Buff_Handler onAvatarOut removes Armed.
	// That removal clears Sword Essence. ClearBuff also drops Disciplinary
	// stacks, and the C6 window is removed on swap.
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		prev := args[0].(int)
		if prev != c.Index() {
			return
		}
		c.exitArmed()
		c.clearDisc()
		c.endTranspose()
	}, "vesna-swap")
	return nil
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "essence":
		return c.essence, nil
	case "armed":
		if c.armed() {
			return 1, nil
		}
		return 0, nil
	case "windborne":
		return c.wbLevel, nil
	case "discipline":
		return c.discCount(), nil
	case "radiance":
		if c.inRadiance() {
			return 1, nil
		}
		return 0, nil
	case "transpose":
		if c.transposeReady() {
			return 1, nil
		}
		return 0, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill {
		// C6 switches the skill button to Transpose. It ignores the 18s CD and essence.
		if c.transposeReady() {
			return true, action.NoFailure
		}
		if c.armed() {
			if c.canWindborne() {
				return true, action.NoFailure
			}
			return false, action.SkillCD
		}
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) armed() bool {
	return c.Core.F < c.armedUntil
}

func (c *char) inRadiance() bool {
	return c.StatusIsActive(radianceKey)
}

func (c *char) transposeReady() bool {
	return c.Base.Cons >= 6 && c.Core.F < c.transposeUntil
}

func (c *char) endTranspose() {
	c.transposeUntil = c.Core.F
	c.DeleteStatus(transposeKey)
}

func (c *char) openTranspose() {
	if c.Base.Cons < 6 {
		return
	}
	c.transposeUntil = c.Core.F + c6Window
	c.AddStatus(transposeKey, c6Window, false)
}

func (c *char) canWindborne() bool {
	if !c.armed() || c.closing {
		return false
	}
	if c.essence >= essenceCost {
		return true
	}
	// C1's first Lv3 each Armed entry skips the cost. The graph compares the value to 29.5.
	return c.c1Free && c.wbLevel >= wbLevelMax
}

func (c *char) addEssence(n float64) {
	c.essence += n
	if c.essence > essenceMax {
		c.essence = essenceMax
	}
	if c.essence < 0 {
		c.essence = 0
	}
}

func (c *char) enterArmed() {
	c.armedGen++
	// onAdded fires ClearBuff before C2's six PermanentSkill_1 triggers.
	c.clearDisc()
	c.addEssence(essenceSkill)
	c.armedUntil = c.Core.F + armedDuration
	c.AddStatus(armedKey, armedDuration, false)
	c.wbLevel = 1
	c.e3Uses = 0
	c.closing = false
	c.c1Free = false
	c.e3Limit = e3Base
	if c.Base.Cons >= 1 {
		c.e3Limit = e3C1
		c.c1Free = true
	}
	if c.Base.Cons >= 2 && c.Base.Ascension >= 1 {
		for range discMax {
			c.addDisc()
		}
	}
	c.scheduleArmedExit(armedDuration)
}

func (c *char) exitArmed() {
	if !c.armed() && c.wbLevel == 0 {
		c.essence = 0
		return
	}
	c.armedUntil = c.Core.F
	c.DeleteStatus(armedKey)
	// The essence UI modifier is removed with Armed and then clears the value.
	c.essence = 0
	c.wbLevel = 0
	c.e3Uses = 0
	c.e3Limit = 0
	c.c1Free = false
	c.closing = false
}

func (c *char) scheduleArmedExit(delay int) {
	gen := c.armedGen
	// The task runs on the expiry frame. armed() is already false then,
	// because the active check is F < armedUntil, so the cleanup cannot
	// require armed() or Sword Essence would be left behind.
	c.Core.Tasks.Add(func() {
		if c.armedGen != gen {
			return
		}
		c.exitArmed()
	}, delay)
}

func (c *char) addDisc() {
	// Rite of Spring's Procession is promote level 1.
	if c.Base.Ascension < 1 {
		return
	}
	c.pruneDisc()
	if len(c.discUntil) >= discMax {
		return
	}
	c.discUntil = append(c.discUntil, c.Core.F+discDuration)
}

func (c *char) pruneDisc() {
	now := c.Core.F
	kept := c.discUntil[:0]
	for _, exp := range c.discUntil {
		if now < exp {
			kept = append(kept, exp)
		}
	}
	c.discUntil = kept
}

func (c *char) discCount() int {
	c.pruneDisc()
	return len(c.discUntil)
}

func (c *char) clearDisc() {
	c.discUntil = nil
}

func (c *char) discFactor() float64 {
	return 1 + discPerStack*float64(c.discCount())
}

func (c *char) jubileeBonus() float64 {
	bonus := c.TotalAtk() / 100.0 * jubileePer
	if bonus > jubileeCap {
		return jubileeCap
	}
	return bonus
}

func (c *char) skillInfo() action.Info {
	return action.Info{
		Frames:          framesAbil(),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
		State:           action.SkillState,
	}
}

func (c *char) tryParticle(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy {
		return
	}
	if c.Core.F < c.particleUntil {
		return
	}
	c.particleUntil = c.Core.F + particleICD
	c.Core.QueueParticle(c.Base.Key.String(), 1, c.Base.Element, c.ParticleDelay)
}

func (c *char) notReady(msg string) (action.Info, error) {
	return action.Info{}, fmt.Errorf("%s", msg)
}
