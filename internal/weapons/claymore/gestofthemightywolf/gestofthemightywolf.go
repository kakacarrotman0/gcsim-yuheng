package gestofthemightywolf

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/glog"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

type Weapon struct {
	Index               int
	char                *character.CharWrapper
	core                *core.Core
	refine              int
	stacks              int
	mod                 []float64
	constructing        action.Action
	isHexereiSecretRite bool
}

const (
	gestAttackSpeedKey = "gest-of-the-mighty-wolf-atkspd"
	gestStacksKey      = "gest-of-the-mighty-wolf-stacks"
)

func (w *Weapon) SetIndex(idx int) { w.Index = idx }

func (w *Weapon) Init() error {
	w.isHexereiSecretRite = w.core.Player.GetHexereiCount() >= 2
	return nil
}

func NewWeapon(c *core.Core, char *character.CharWrapper, p info.WeaponProfile) (info.Weapon, error) {
	w := &Weapon{
		char:   char,
		core:   c,
		refine: p.Refine,
		mod:    make([]float64, attributes.EndStatType),
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.AtkSpd] = 0.1
	// Exec emits the action, builds it, then stores the new state. Attack speed
	// is read during the build, so CurrentState still describes the previous
	// action. The constructing flag is that in-between sample: a Normal Attack
	// being built receives the 10%, and any other action being built does not.
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase(gestAttackSpeedKey, -1),
		AffectedStat: attributes.AtkSpd,
		Amount: func() []float64 {
			switch w.constructing {
			case action.ActionAttack:
				return m
			case action.InvalidAction:
				if c.Player.CurrentState() == action.NormalAttackState {
					return m
				}
			}
			return nil
		},
	})
	w.watchConstruction(c, char)

	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != char.Index() {
			return
		}
		if w.core.Player.Active() != char.Index() {
			return
		}
		if atk.Info.AttackTag != attacks.AttackTagNormal {
			return
		}
		w.addStacks("normal", 1)
	}, fmt.Sprintf("gest-of-the-mighty-wolf-on-normal-attack-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnStateChange, func(args ...any) {
		next := args[1].(action.AnimationState)
		segmented := args[2].(bool)

		if w.core.Player.Active() != char.Index() {
			return
		}
		if next == action.ChargeAttackState && !segmented {
			w.addStacks("charge", 2)
		}
	}, fmt.Sprintf("gest-of-the-mighty-wolf-on-charge-%v", char.Base.Key.String()))

	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		if w.core.Player.Active() != char.Index() {
			return
		}
		w.addStacks("skill", 2)
	}, fmt.Sprintf("gest-of-the-mighty-wolf-on-skill-%v", char.Base.Key.String()))

	return w, nil
}

func (w *Weapon) watchConstruction(c *core.Core, char *character.CharWrapper) {
	marks := []struct {
		ev  event.Event
		act action.Action
	}{
		{event.OnAttack, action.ActionAttack},
		{event.OnChargeAttack, action.ActionCharge},
		{event.OnSkill, action.ActionSkill},
		{event.OnBurst, action.ActionBurst},
		{event.OnDash, action.ActionDash},
		{event.OnPlunge, action.ActionHighPlunge},
		{event.OnAimShoot, action.ActionAim},
	}
	for _, mark := range marks {
		c.Events.Subscribe(mark.ev, func(args ...any) {
			if c.Player.Active() != char.Index() {
				return
			}
			w.constructing = mark.act
		}, fmt.Sprintf("gest-of-the-mighty-wolf-construct-%v-%v", mark.act, char.Base.Key.String()))
	}
	c.Events.Subscribe(event.OnStateChange, func(args ...any) {
		if c.Player.Active() != char.Index() {
			return
		}
		w.constructing = action.InvalidAction
	}, fmt.Sprintf("gest-of-the-mighty-wolf-construct-clear-%v", char.Base.Key.String()))
}

func (w *Weapon) addStacks(src string, amt int) {
	if !w.char.StatModIsActive(gestStacksKey) {
		w.stacks = 0
	}

	w.char.AddStatMod(character.StatMod{
		Base: modifier.NewBase(gestStacksKey, 4*60),
		Amount: func() []float64 {
			w.mod[attributes.DmgP] = (0.055 + 0.02*float64(w.refine)) * float64(w.stacks)

			if !w.isHexereiSecretRite {
				return w.mod
			}

			w.mod[attributes.CD] = (0.055 + 0.02*float64(w.refine)) * float64(w.stacks)

			return w.mod
		},
	})

	w.stacks = min(4, w.stacks+amt)
	if w.core.Flags.LogDebug {
		w.core.Log.NewEvent(fmt.Sprintf("gest adding %v-stack", src), glog.LogWeaponEvent, w.char.Index()).
			Write("stacks", w.stacks)
	}
}
