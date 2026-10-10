package lohen

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

const (
	masterKey   = "lohen-masterstroke"
	spiritsKey  = "lohen-high-spirits"
	hexKey      = "lohen-hexerei"
	a4Key       = "lohen-a4"
	c2EMKey     = "lohen-c2-em"
	particleKey = "lohen-particle"

	joyCap       = 100.0
	willCapBase  = 100.0
	etchedBase   = 3
	willFailGain = 1.0
	// Talent text. ProudSkill rows for these passives are not in ConfigTalent;
	// the talent file only binds %1/%2/%3 onto the specials below.
	a1ExtraRatio = 30.0
	a1ExtraWill  = 60.0
	a4ATK        = 0.15
	a4Duration   = 8 * 60
	spiritsBase  = 9 * 60
	spiritsExtra = 6 * 60
	spiritsICD   = 18 * 60
	hexThreshold = 0.5
	hexBonus     = 0.40
	hexDuration  = 6 * 60
	c1Efficiency = 5.0
	c1MaxRatio   = 3.0
	c2Ratio      = 5.0
	c2EM         = 200.0
	c2Duration   = 8 * 60
	c2Window     = 4 * 60
	c4Energy     = 15.0
	c4Window     = 15 * 60
	c6ICD        = 7 * 60
	c6Crit       = 1.75
	c6Extend     = 75 // 1.25s, AddElementDurability on the C6 handler
	burstExtend  = 99 // 1.65s, AddElementDurability on the burst handler
	masterCap    = 20 * 60
	particleICD  = 2 * 60
	skillCD      = 18 * 60
	burstCDDur   = 15 * 60
)

// Burst gadget count steps, assuming the 0.1s think starts immediately.
// FirstAttack, four Attack steps, LastAttack. Blank is not a hit.
var burstHitmarks = []int{12, 36, 42, 48, 54, 78}

type char struct {
	*tmpl.Character
	joy             float64
	will            float64
	extraE          int
	masterUntil     int
	c6Mark          bool
	c6ICDUntil      int
	c2Until         int
	c2ICDUntil      int
	c4RefundUntil   int
	spiritsICDUntil int
	particleUntil   int
}

func NewChar(s *core.Core, w *character.CharWrapper, p info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)
	c.EnergyMax = 60
	c.NormalHitNum = normalHitNum
	c.SkillCon = 3
	c.BurstCon = 5
	// Witch's Homework is a quest. Unlike Fischl, leave Hexerei off unless set.
	if p.Params["hexerei"] == 1 {
		c.IsHexerei = true
	}
	w.Character = &c
	return nil
}

func (c *char) Init() error {
	c.willInit()
	c.a4Init()
	c.c6Init()
	// GrandHandler onAvatarOut removes the modifier. onRemoved then zeroes
	// Joy, Will, and the C6 mark.
	c.Core.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		prev := args[0].(int)
		if prev != c.Index() {
			return
		}
		c.endMasterstroke()
	}, "lohen-swap")
	return nil
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "joy":
		return c.joy, nil
	case "will":
		return c.will, nil
	case "masterstroke":
		if c.masterActive() {
			return 1, nil
		}
		return 0, nil
	case "etched":
		return c.extraE, nil
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) ActionReady(a action.Action, p map[string]int) (bool, action.Failure) {
	if a == action.ActionSkill && p["etched"] > 0 {
		if c.masterActive() && c.joy >= joyCap && c.extraE < c.etchedMax() {
			return true, action.NoFailure
		}
		return false, action.SkillCD
	}
	return c.Character.ActionReady(a, p)
}

func (c *char) ActionStam(a action.Action, p map[string]int) float64 {
	if a == action.ActionCharge && c.masterActive() {
		return enhancedStamina[c.skillLvl()]
	}
	return c.Character.ActionStam(a, p)
}

func (c *char) skillLvl() int {
	lvl := c.TalentLvlSkill()
	if c.StatusIsActive(spiritsKey) {
		lvl++
	}
	if lvl > 14 {
		return 14
	}
	return lvl
}

func (c *char) masterActive() bool {
	return c.Core.F < c.masterUntil
}

func (c *char) endMasterstroke() {
	if !c.masterActive() {
		return
	}
	c.masterUntil = c.Core.F
	c.DeleteStatus(masterKey)
	c.joy = 0
	c.will = 0
	c.c6Mark = false
}

func (c *char) etchedMax() int {
	if c.Base.Cons >= 6 {
		return etchedBase + 2
	}
	return etchedBase
}

func (c *char) willMax() float64 {
	ratio := 1.0
	if c.Base.Cons >= 1 {
		ratio = c1MaxRatio
	}
	return willCapBase * ratio
}

func (c *char) willEfficiency() float64 {
	if c.Base.Cons >= 1 {
		return c1Efficiency
	}
	return 1
}

func (c *char) addWill(amt float64) {
	if !c.masterActive() || amt == 0 {
		return
	}
	c.will += amt
	if c.will > c.willMax() {
		c.will = c.willMax()
	}
	if c.will < 0 {
		c.will = 0
	}
}

func (c *char) addJoy(amt float64) {
	if !c.masterActive() {
		return
	}
	// Joy stops once etched uses have reached the cap. The graph compares the
	// counter with MaxCount-1 before a new point is added.
	if c.extraE > c.etchedMax()-1 {
		return
	}
	c.joy += amt
	if c.joy > joyCap {
		c.joy = joyCap
	}
}

func (c *char) extendMaster(frames int) {
	if !c.masterActive() || frames <= 0 {
		return
	}
	remain := c.masterUntil - c.Core.F + frames
	if remain > masterCap {
		remain = masterCap
	}
	c.masterUntil = c.Core.F + remain
	c.AddStatus(masterKey, remain, true)
}

func (c *char) c6KeepWill() bool {
	return c.Base.Cons >= 6 && c.masterActive() && c.Core.F >= c.c6ICDUntil
}

func (c *char) damageRatio(will float64, perPoint float64) float64 {
	return 1 + will*perPoint
}
