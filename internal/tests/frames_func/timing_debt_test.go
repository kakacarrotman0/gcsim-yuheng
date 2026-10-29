package canqueueafter

import (
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/keys"
)

func TestKnownTimingDebtMatchesOnlyItsExactRow(t *testing.T) {
	seen := map[timingDebt]bool{}
	for _, d := range knownTimingDebt {
		if seen[d] {
			t.Fatalf("duplicate debt %+v", d)
		}
		seen[d] = true
		if !isKnownTimingDebt(d.char, d.executed, d.transition, d.animation, d.canQueue, d.frame, d.kind) {
			t.Fatalf("recorded debt was not recognized: %+v", d)
		}
		switch d.kind {
		case debtFramePastAnimation:
			if d.frame <= d.animation {
				t.Fatalf("%+v does not extend past the animation", d)
			}
		case debtFrameBeforeQueue:
			if d.frame >= d.canQueue {
				t.Fatalf("%+v is not earlier than the queue open", d)
			}
		default:
			t.Fatalf("unknown kind %d", d.kind)
		}
	}
}

func TestChangedTimingDebtValuesFail(t *testing.T) {
	for _, d := range knownTimingDebt {
		cases := []struct {
			name string
			edit func(*timingDebt)
		}{
			{"frame", func(row *timingDebt) { row.frame++ }},
			{"animation", func(row *timingDebt) { row.animation++ }},
			{"queue", func(row *timingDebt) { row.canQueue++ }},
			{"character", func(row *timingDebt) { row.char = keys.Zhongli }},
			{"executed", func(row *timingDebt) {
				if row.executed == action.ActionAttack {
					row.executed = action.ActionBurst
				} else {
					row.executed = action.ActionAttack
				}
			}},
			{"transition", func(row *timingDebt) {
				row.transition = unusedTransition(*row)
			}},
			{"kind", func(row *timingDebt) {
				if row.kind == debtFramePastAnimation {
					row.kind = debtFrameBeforeQueue
				} else {
					row.kind = debtFramePastAnimation
				}
			}},
		}
		for _, tc := range cases {
			row := d
			tc.edit(&row)
			if isKnownTimingDebt(row.char, row.executed, row.transition, row.animation, row.canQueue, row.frame, row.kind) {
				t.Fatalf("%s change still matched: %+v", tc.name, row)
			}
		}
	}
}

func TestLinneaN3DebtListsEveryTransition(t *testing.T) {
	got := map[action.Action]timingDebt{}
	for _, d := range knownTimingDebt {
		if d.char != keys.Linnea || d.executed != action.ActionAttack || d.animation != 43 || d.canQueue != 47 {
			continue
		}
		if _, ok := got[d.transition]; ok {
			t.Fatalf("two N3 rows for %s", d.transition)
		}
		got[d.transition] = d
	}
	for act := range action.EndActionType {
		d, ok := got[act]
		if !ok {
			t.Fatalf("N3 has no debt row for %s", act)
		}
		switch act {
		case action.ActionAim, action.ActionSkill, action.ActionBurst, action.ActionDash, action.ActionJump, action.ActionSwap:
			if d.frame != 47 || d.kind != debtFramePastAnimation {
				t.Fatalf("%s row %+v", act, d)
			}
		default:
			if d.frame != 43 || d.kind != debtFrameBeforeQueue {
				t.Fatalf("%s row %+v", act, d)
			}
		}
	}
}

func TestSyntheticDebtRowsAndMutations(t *testing.T) {
	if act, ok := firstUnexpectedFrame(keys.Kachina, action.ActionAttack, framesInfo(51, 16, map[action.Action]int{
		action.ActionCharge: 500,
	})); ok {
		t.Fatalf("Kachina N4 debt reported %s", act)
	}
	if act, ok := firstUnexpectedFrame(keys.Kachina, action.ActionCharge, framesInfo(55, 20, map[action.Action]int{
		action.ActionCharge: 500,
	})); ok {
		t.Fatalf("Kachina charge debt reported %s", act)
	}
	if _, ok := firstUnexpectedFrame(keys.Kachina, action.ActionAttack, framesInfo(51, 16, map[action.Action]int{
		action.ActionCharge: 499,
	})); !ok {
		t.Fatal("Kachina charge frame 499 was accepted")
	}
	if _, ok := firstUnexpectedFrame(keys.Zhongli, action.ActionAttack, framesInfo(51, 16, map[action.Action]int{
		action.ActionCharge: 500,
	})); !ok {
		t.Fatal("Zhongli inherited Kachina's N4 exception")
	}

	skill := map[action.Action]int{
		action.ActionAttack: 33,
		action.ActionAim:    44,
		action.ActionSkill:  20,
		action.ActionBurst:  15,
		action.ActionWalk:   32,
	}
	if act, ok := firstUnexpectedFrame(keys.Linnea, action.ActionSkill, framesInfo(14, 13, skill)); ok {
		t.Fatalf("Linnea skill debt reported %s", act)
	}
	skill[action.ActionAttack] = 34
	if _, ok := firstUnexpectedFrame(keys.Linnea, action.ActionSkill, framesInfo(14, 13, skill)); !ok {
		t.Fatal("Linnea skill attack frame 34 was accepted")
	}

	n3 := map[action.Action]int{}
	for act := range action.EndActionType {
		n3[act] = 43
	}
	for _, act := range []action.Action{action.ActionAim, action.ActionSkill, action.ActionBurst, action.ActionDash, action.ActionJump, action.ActionSwap} {
		n3[act] = 47
	}
	if act, ok := firstUnexpectedFrame(keys.Linnea, action.ActionAttack, framesInfo(43, 47, n3)); ok {
		t.Fatalf("Linnea N3 debt reported %s", act)
	}
	n3[action.ActionDash] = 46
	if _, ok := firstUnexpectedFrame(keys.Linnea, action.ActionAttack, framesInfo(43, 47, n3)); !ok {
		t.Fatal("Linnea N3 dash frame 46 was accepted")
	}
	n3[action.ActionDash] = 47
	n3[action.InvalidAction] = 42
	if _, ok := firstUnexpectedFrame(keys.Linnea, action.ActionAttack, framesInfo(43, 47, n3)); !ok {
		t.Fatal("Linnea N3 invalid frame 42 was accepted")
	}

	if _, ok := firstUnexpectedFrame(keys.Kachina, action.ActionAttack, framesInfo(31, 11, map[action.Action]int{
		action.ActionCharge: 40,
	})); !ok {
		t.Fatal("an ordinary Kachina N1 charge overflow was accepted")
	}
}

func unusedTransition(row timingDebt) action.Action {
	for act := range action.EndActionType {
		if act == row.transition {
			continue
		}
		if !isKnownTimingDebt(row.char, row.executed, act, row.animation, row.canQueue, row.frame, row.kind) {
			return act
		}
	}
	return row.transition
}

func framesInfo(animation, canQueue int, extra map[action.Action]int) action.Info {
	return action.Info{
		AnimationLength: animation,
		CanQueueAfter:   canQueue,
		Frames: func(next action.Action) int {
			if frame, ok := extra[next]; ok {
				return frame
			}
			return animation
		},
	}
}

func firstUnexpectedFrame(char keys.Char, executed action.Action, evt action.Info) (action.Action, bool) {
	for act := range action.EndActionType {
		frame := evt.Frames(act)
		if frame > evt.AnimationLength && !isKnownTimingDebt(char, executed, act, evt.AnimationLength, evt.CanQueueAfter, frame, debtFramePastAnimation) {
			return act, true
		}
		if frame < evt.CanQueueAfter && !isKnownTimingDebt(char, executed, act, evt.AnimationLength, evt.CanQueueAfter, frame, debtFrameBeforeQueue) {
			return act, true
		}
	}
	return 0, false
}
