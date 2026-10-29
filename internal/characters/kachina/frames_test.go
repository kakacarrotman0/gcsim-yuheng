package kachina

import (
	"testing"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
)

// 500 is the sheet sentinel for a charge transition that does not exist
// during the action. N4 and a charge itself keep that value. A following
// charge stays locked until the animation state expires, and N1 still uses
// its own charge cancel.
func TestChargeTransitionsStayUnavailableUntilTheActionExpires(t *testing.T) {
	if attackFrames[3][action.ActionCharge] != 500 {
		t.Fatalf("N4 charge frame %d", attackFrames[3][action.ActionCharge])
	}
	if chargeFrames[action.ActionCharge] != 500 {
		t.Fatalf("charge chain frame %d", chargeFrames[action.ActionCharge])
	}

	n4c, n4 := newKachina(t)
	n4.NormalCounter = 3
	if err := n4c.Player.Exec(action.ActionAttack, keys.Kachina, nil); err != nil {
		t.Fatal(err)
	}
	assertChargeLocked(t, n4c, attackFrames[3][action.InvalidAction], "N4")

	cac, _ := newKachina(t)
	if err := cac.Player.Exec(action.ActionAttack, keys.Kachina, nil); err != nil {
		t.Fatal(err)
	}
	ready := firstChargeTick(t, cac, attackFrames[0][action.InvalidAction])
	if ready != attackFrames[0][action.ActionCharge] {
		t.Fatalf("N1 charge ready at %d, cancel is %d", ready, attackFrames[0][action.ActionCharge])
	}
	if err := cac.Player.Exec(action.ActionCharge, keys.Kachina, nil); err != nil {
		t.Fatal(err)
	}
	assertChargeLocked(t, cac, chargeFrames[action.InvalidAction], "charge")
}

func newKachina(t *testing.T) (*core.Core, *character.CharWrapper) {
	t.Helper()
	c, _ := makeCore(1)
	ch := addKachina(t, c, 0, 1)
	return c, ch
}

func assertChargeLocked(t *testing.T, c *core.Core, animation int, label string) {
	t.Helper()
	for i := 1; i <= animation; i++ {
		advance(c, 1)
		if err := c.Player.ReadyCheck(action.ActionCharge, keys.Kachina, nil); err == nil {
			t.Fatalf("%s charge ready at tick %d, animation %d", label, i, animation)
		}
	}
	advance(c, 1)
	if err := c.Player.ReadyCheck(action.ActionCharge, keys.Kachina, nil); err != nil {
		t.Fatalf("%s charge locked after animation %d expired: %v", label, animation, err)
	}
}

func firstChargeTick(t *testing.T, c *core.Core, limit int) int {
	t.Helper()
	for i := 1; i <= limit; i++ {
		advance(c, 1)
		if c.Player.ReadyCheck(action.ActionCharge, keys.Kachina, nil) == nil {
			return i
		}
	}
	t.Fatalf("charge never ready within %d", limit)
	return 0
}
