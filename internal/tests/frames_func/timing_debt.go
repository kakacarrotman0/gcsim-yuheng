package canqueueafter

import (
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/keys"
)

// debtKind names which TestFramesFunc comparison a row excuses.
// These rows are unvalidated timing debt. They are not measured cancels.
type debtKind int

const (
	// The cancel frame is later than AnimationLength, so the action state
	// expires before that transition can be taken.
	debtFramePastAnimation debtKind = iota
	// The cancel frame is earlier than CanQueueAfter, so the queue opens
	// after that transition's recorded frame.
	debtFrameBeforeQueue
)

// timingDebt is one exact character, executed action, transition, and frame
// triple. A different value, character, action, or transition does not match.
type timingDebt struct {
	char       keys.Char
	executed   action.Action
	transition action.Action
	animation  int
	canQueue   int
	frame      int
	kind       debtKind
}

// knownTimingDebt is the only set of frame violations TestFramesFunc allows.
// Each row is an existing unmeasured placeholder or an intentional
// unavailable transition. Do not add a row to silence a new failure.
var knownTimingDebt = []timingDebt{
	// Kachina N4 → Charge. Animation length 51, hitmark and queue open 16.
	// 500 is the sheet sentinel: the charge cancel is not available during N4.
	{
		char: keys.Kachina, executed: action.ActionAttack, transition: action.ActionCharge,
		animation: 51, canQueue: 16, frame: 500, kind: debtFramePastAnimation,
	},
	// Kachina Charge → Charge. Animation length 55, queue open 20.
	// 500 means a second charge is not available during the charge.
	{
		char: keys.Kachina, executed: action.ActionCharge, transition: action.ActionCharge,
		animation: 55, canQueue: 20, frame: 500, kind: debtFramePastAnimation,
	},

	// Linnea Skill. The draft animation length is the unknown 14-frame slice.
	// Queue open is the swap cancel, 13. These cancels are longer than 14,
	// so they are not taken. They are not measured replacements for 14.
	{
		char: keys.Linnea, executed: action.ActionSkill, transition: action.ActionAttack,
		animation: 14, canQueue: 13, frame: 33, kind: debtFramePastAnimation,
	},
	{
		char: keys.Linnea, executed: action.ActionSkill, transition: action.ActionAim,
		animation: 14, canQueue: 13, frame: 44, kind: debtFramePastAnimation,
	},
	{
		char: keys.Linnea, executed: action.ActionSkill, transition: action.ActionSkill,
		animation: 14, canQueue: 13, frame: 20, kind: debtFramePastAnimation,
	},
	{
		char: keys.Linnea, executed: action.ActionSkill, transition: action.ActionBurst,
		animation: 14, canQueue: 13, frame: 15, kind: debtFramePastAnimation,
	},
	{
		char: keys.Linnea, executed: action.ActionSkill, transition: action.ActionWalk,
		animation: 14, canQueue: 13, frame: 32, kind: debtFramePastAnimation,
	},

	// Linnea N3. Hitmark and queue open are 47. The N3 → N1 frame, which is
	// also the animation length, is 43. attack.go still says the frames are
	// unimplemented. Both numbers are placeholders.
	//
	// InitNormalCancelSlice puts the hitmark on aim, skill, burst, dash,
	// jump, and swap. Those 47s are past the 43-frame animation.
	{
		char: keys.Linnea, executed: action.ActionAttack, transition: action.ActionAim,
		animation: 43, canQueue: 47, frame: 47, kind: debtFramePastAnimation,
	},
	{
		char: keys.Linnea, executed: action.ActionAttack, transition: action.ActionSkill,
		animation: 43, canQueue: 47, frame: 47, kind: debtFramePastAnimation,
	},
	{
		char: keys.Linnea, executed: action.ActionAttack, transition: action.ActionBurst,
		animation: 43, canQueue: 47, frame: 47, kind: debtFramePastAnimation,
	},
	{
		char: keys.Linnea, executed: action.ActionAttack, transition: action.ActionDash,
		animation: 43, canQueue: 47, frame: 47, kind: debtFramePastAnimation,
	},
	{
		char: keys.Linnea, executed: action.ActionAttack, transition: action.ActionJump,
		animation: 43, canQueue: 47, frame: 47, kind: debtFramePastAnimation,
	},
	{
		char: keys.Linnea, executed: action.ActionAttack, transition: action.ActionSwap,
		animation: 43, canQueue: 47, frame: 47, kind: debtFramePastAnimation,
	},
}

func init() {
	// The remaining N3 transitions stay at the 43-frame animation length,
	// which is earlier than the 47-frame queue open.
	for _, transition := range []action.Action{
		action.InvalidAction,
		action.ActionAttack,
		action.ActionCharge,
		action.ActionHighPlunge,
		action.ActionLowPlunge,
		action.ActionWalk,
		action.ActionWait,
		action.ActionDelay,
		action.ActionSpecialSkill,
	} {
		knownTimingDebt = append(knownTimingDebt, timingDebt{
			char: keys.Linnea, executed: action.ActionAttack, transition: transition,
			animation: 43, canQueue: 47, frame: 43, kind: debtFrameBeforeQueue,
		})
	}
}

func isKnownTimingDebt(char keys.Char, executed, transition action.Action, animation, canQueue, frame int, kind debtKind) bool {
	for _, d := range knownTimingDebt {
		if d.char == char && d.executed == executed && d.transition == transition &&
			d.animation == animation && d.canQueue == canQueue && d.frame == frame && d.kind == kind {
			return true
		}
	}
	return false
}
