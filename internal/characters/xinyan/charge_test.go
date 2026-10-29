package xinyan

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestWeapon()
}

func makeCore(t *testing.T) *core.Core {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
	e := enemy.New(c, info.EnemyProfile{
		Level:  100,
		Resist: make(map[attributes.Element]float64),
		Pos:    info.Coord{X: 0, Y: 0, R: 1},
	})
	c.Combat.AddEnemy(e)
	c.Combat.DefaultTarget = e.Key()
	c.Player.SetActive(0)
	return c
}

func addXinyan(t *testing.T, c *core.Core, cons int) {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Xinyan, testhelper.TestWeaponKey)
	prof.Base.Cons = cons
	prof.Talents.Attack = 9
	prof.Stats[attributes.CR] = 1
	if _, err := c.AddChar(prof); err != nil {
		t.Fatal(err)
	}
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
}

type hit struct {
	frame int
	info  info.AttackInfo
}

func watch(c *core.Core) *[]hit {
	var hits []hit
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hits = append(hits, hit{frame: c.F, info: args[1].(*info.AttackEvent).Info})
	}, "hits")
	return &hits
}

func advance(c *core.Core, n int) {
	for range n {
		c.F++
		_ = c.Tick()
	}
}

func untilQueue(t *testing.T, c *core.Core) {
	t.Helper()
	for range 400 {
		if c.Player.CanQueueNextAction() {
			return
		}
		advance(c, 1)
	}
	t.Fatal("action never opened the queue")
}

func chargeHitFrame(t *testing.T, cons int) int {
	t.Helper()
	c := makeCore(t)
	addXinyan(t, c, cons)
	hits := watch(c)
	if err := c.Player.Exec(action.ActionAttack, keys.Xinyan, nil); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c)
	if err := c.Player.Exec(action.ActionCharge, keys.Xinyan, nil); err != nil {
		t.Fatal(err)
	}
	started := c.F
	untilQueue(t, c)
	for i := range *hits {
		h := &(*hits)[i]
		if h.info.Abil == "Charge Attack" {
			return h.frame - started
		}
	}
	t.Fatal("no cyclic hit")
	return 0
}

func TestDanceOnFireCharge(t *testing.T) {
	c := makeCore(t)
	addXinyan(t, c, 0)
	hits := watch(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Xinyan, nil); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c)
	if len(*hits) != 1 {
		t.Fatalf("hits %d", len(*hits))
	}
	got := (*hits)[0].info
	if got.Abil != "Charge Attack" || got.Mult != charge[8] || got.PoiseDMG != 60 || got.Durability != 25 {
		t.Fatalf("cyclic %+v", got)
	}
	if got.AttackTag != attacks.AttackTagExtra || got.Element != attributes.Physical || got.StrikeType != attacks.StrikeTypeBlunt {
		t.Fatalf("class %+v", got)
	}
	if math.Abs(c.Player.Stam-(player.MaxStam-12)) > 1e-6 {
		t.Fatalf("stamina %v", c.Player.Stam)
	}

	if err := c.Player.Exec(action.ActionCharge, keys.Xinyan, map[string]int{"final": 1}); err != nil {
		t.Fatal(err)
	}
	for range 80 {
		advance(c, 1)
		if (*hits)[len(*hits)-1].info.Abil == "Charge Attack (Finisher)" {
			break
		}
	}
	fin := (*hits)[len(*hits)-1].info
	if fin.Abil != "Charge Attack (Finisher)" || fin.Mult != chargeFinal[8] || fin.PoiseDMG != 120 {
		t.Fatalf("finisher %+v", fin)
	}
}

func TestFatalAccelerationSpeedsChargedAttack(t *testing.T) {
	c0 := chargeHitFrame(t, 0)
	c1 := chargeHitFrame(t, 1)
	if c1 >= c0 {
		t.Fatalf("c1 delay %d c0 delay %d", c1, c0)
	}
}

func TestRockinInAFlamingWorld(t *testing.T) {
	c0 := makeCore(t)
	addXinyan(t, c0, 0)
	c6 := makeCore(t)
	addXinyan(t, c6, 6)
	ch := c6.Player.ByIndex(0)
	bonus := ch.Stat(attributes.ATK) - c0.Player.ByIndex(0).Stat(attributes.ATK)
	if math.Abs(bonus-0.5*ch.TotalDef(false)) > 1e-6 {
		t.Fatalf("atk bonus %v def %v", bonus, ch.TotalDef(false))
	}

	if err := c6.Player.Exec(action.ActionCharge, keys.Xinyan, nil); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c6)
	if math.Abs(c6.Player.Stam-(player.MaxStam-8.4)) > 1e-6 {
		t.Fatalf("c6 stamina %v", c6.Player.Stam)
	}
}
