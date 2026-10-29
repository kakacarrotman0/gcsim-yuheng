package character

import (
	"errors"

	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// Neutral claymore spin playback. Dehya and Xinyan have no published hitmarks,
// so these four frame values are Noelle's neutral schedule and are not a
// measurement of the character that plays them. 40/s and 5s are the shared
// talent values. The drain steps every 0.3s, which is the existing claymore tick.
const (
	ClaymoreSpinWindup     = 56
	ClaymoreSpinHit        = 23
	ClaymoreSpinFinalHit   = 26
	ClaymoreSpinTransition = 13
	ClaymoreSpinDuration   = 5 * 60

	claymoreSpinStamina = 40.0
	claymoreSpinTick    = 0.3
)

type SpinChargeState struct {
	StartF int
	Active bool
	Ended  bool
	Error  bool
}

type SpinChargeSpec struct {
	Windup     int
	SpinHit    int
	FinalHit   int
	Transition int
	Duration   int
	Adjust     func(int, float64) int
	Hit        func(final bool)
}

func (c *Character) SpinChargeAttack(st *SpinChargeState, spec SpinChargeSpec, p map[string]int) (action.Info, error) {
	if c.Core.Player.CurrentState() != action.ChargeAttackState {
		*st = SpinChargeState{}
	}
	if st.Error {
		return action.Info{}, errors.New("use of delays between charge attack hits is invalid")
	}

	final := st.Ended || p["final"] != 0
	windup := spec.Windup
	if st.Active && c.Core.Player.CurrentState() == action.ChargeAttackState {
		windup = 0
	}
	atkspd := c.Stat(attributes.AtkSpd)
	scale := func(f int) int {
		if spec.Adjust == nil {
			return f
		}
		return spec.Adjust(f, atkspd)
	}
	windup = scale(windup)

	transFrames := 0
	if !final || st.Active {
		transFrames = scale(spec.Transition)
	}
	hitmark := scale(spec.SpinHit)
	spinFrames := 0
	if final {
		hitmark = transFrames + scale(spec.FinalHit)
	} else {
		spinFrames = hitmark
	}

	act := action.Info{
		Frames: func(action.Action) int { return windup + hitmark },
		OnRemoved: func(next action.AnimationState) {
			if next != action.ChargeAttackState {
				*st = SpinChargeState{}
			}
		},
		AnimationLength: windup + hitmark,
		CanQueueAfter:   windup + hitmark,
		Segmented:       st.Active,
		State:           action.ChargeAttackState,
	}
	if !final {
		act.AnimationLength = windup + spinFrames + transFrames + scale(spec.FinalHit)
	}
	act.QueueAction(func() { st.Error = !final }, act.CanQueueAfter+1)
	act.QueueAction(func() {
		c.Core.Tasks.Add(func() { spec.Hit(final) }, 0)
	}, windup+hitmark)

	finalStart := windup + spinFrames + transFrames
	act.QueueAction(func() { *st = SpinChargeState{} }, finalStart)
	if !final {
		act.QueueAction(func() {
			c.Core.Tasks.Add(func() { spec.Hit(true) }, 0)
		}, act.AnimationLength)
	}

	if !st.Active && !final {
		src := c.Core.F
		st.StartF = src
		st.Active = true
		if p["no_limit"] == 0 {
			dur := spec.Duration
			c.Core.Tasks.Add(func() {
				if st.Active && st.StartF == src {
					st.Ended = true
				}
			}, dur)
		}
		act.QueueAction(func() { c.spinStamina(src, st) }, windup)
	}

	return act, nil
}

func (c *Character) spinStamina(src int, st *SpinChargeState) {
	c.QueueCharTask(func() {
		if !st.Active || src != st.StartF || st.Ended {
			return
		}
		if c.Core.Player.Stam == 0 {
			st.Ended = true
			return
		}
		r := 1 + c.Core.Player.StamPercentMod(action.ActionCharge)
		if r < 0 {
			r = 0
		}
		c.Core.Player.UseStam((claymoreSpinStamina*r)*claymoreSpinTick, action.ActionCharge)
		c.spinStamina(src, st)
	}, int(60*claymoreSpinTick))
}

func (c *Character) ClaymoreSpinAttack(final bool, cyclic, finisher float64, cb info.AttackCBFunc) {
	ai := info.AttackInfo{
		ActorIndex:         c.Index(),
		Abil:               "Charge Attack",
		AttackTag:          attacks.AttackTagExtra,
		ICDTag:             attacks.ICDTagNormalAttack,
		ICDGroup:           attacks.ICDGroupDefault,
		StrikeType:         attacks.StrikeTypeBlunt,
		PoiseDMG:           60,
		Element:            attributes.Physical,
		Durability:         25,
		Mult:               cyclic,
		HitlagFactor:       0.01,
		HitlagHaltFrames:   0.03 * 60,
		CanBeDefenseHalted: true,
	}
	offset := 0.3
	radius := 3.0
	if final {
		ai.Abil = "Charge Attack (Finisher)"
		ai.PoiseDMG = 120
		ai.Mult = finisher
		ai.HitlagHaltFrames = 0.15 * 60
		offset = 0
		radius = 3.5
	}
	c.Core.QueueAttack(
		ai,
		combat.NewCircleHitOnTarget(c.Core.Combat.Player(), info.Point{Y: offset}, radius),
		0,
		0,
		cb,
	)
}
