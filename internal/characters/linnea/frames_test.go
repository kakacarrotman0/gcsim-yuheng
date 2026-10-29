package linnea

import (
	"testing"

	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/keys"
)

// Burst's skill cancel has to be ready on the tick the action loop opens the
// queue. CanQueueAfter is that tick. A skill frame earlier than it is never
// requested.
func TestBurstSkillCancelIsReadyWhenTheQueueOpens(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Linnea, 0, 1)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	ch.Energy = ch.EnergyMax
	if err := c.Player.Exec(action.ActionBurst, keys.Linnea, nil); err != nil {
		t.Fatal(err)
	}

	var skillTick, queueTick int
	for i := 1; i <= burstFrames[action.InvalidAction]+2; i++ {
		advance(c, 1)
		if skillTick == 0 && c.Player.ReadyCheck(action.ActionSkill, keys.Linnea, nil) == nil {
			skillTick = i
		}
		if queueTick == 0 && c.Player.CanQueueNextAction() {
			queueTick = i
		}
	}
	if skillTick == 0 || queueTick == 0 {
		t.Fatalf("skill ready %d, queue open %d", skillTick, queueTick)
	}
	if skillTick != queueTick {
		t.Fatalf("skill ready at tick %d, queue opens at tick %d", skillTick, queueTick)
	}
	if skillTick != burstFrames[action.ActionSwap] {
		t.Fatalf("ready tick %d, swap cancel %d", skillTick, burstFrames[action.ActionSwap])
	}

	infoCore, _ := makeCore(1)
	infoChar := addProfile(t, infoCore, keys.Linnea, 0, 1)
	if err := infoCore.Init(); err != nil {
		t.Fatal(err)
	}
	infoChar.Energy = infoChar.EnergyMax
	evt, err := infoChar.Burst(nil)
	if err != nil {
		t.Fatal(err)
	}
	for act := range action.EndActionType {
		if evt.Frames(act) > evt.AnimationLength {
			t.Fatalf("%s frame %d exceeds animation %d", act, evt.Frames(act), evt.AnimationLength)
		}
		if evt.Frames(act) < evt.CanQueueAfter {
			t.Fatalf("queue opens at %d, later than %s frame %d", evt.CanQueueAfter, act, evt.Frames(act))
		}
	}
}
