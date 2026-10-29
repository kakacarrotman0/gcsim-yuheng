package dehya

import (
	"math"
	"testing"

	tmpl "github.com/genshinsim/gcsim/internal/template/character"
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

func addDehya(t *testing.T, c *core.Core, talent int) *characterWait {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Dehya, testhelper.TestWeaponKey)
	prof.Talents.Attack = talent
	if _, err := c.AddChar(prof); err != nil {
		t.Fatal(err)
	}
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	return &characterWait{c: c}
}

type characterWait struct{ c *core.Core }

type hit struct {
	frame int
	info  info.AttackInfo
}

func (w *characterWait) watch() *[]hit {
	var hits []hit
	w.c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hits = append(hits, hit{frame: w.c.F, info: args[1].(*info.AttackEvent).Info})
	}, "hits")
	return &hits
}

func advance(c *core.Core, n int) {
	for range n {
		c.F++
		_ = c.Tick()
	}
}

func (w *characterWait) charge(t *testing.T, p map[string]int) {
	t.Helper()
	if err := w.c.Player.Exec(action.ActionCharge, keys.Dehya, p); err != nil {
		t.Fatal(err)
	}
}

func (w *characterWait) untilQueue(t *testing.T) {
	t.Helper()
	for range 400 {
		if w.c.Player.CanQueueNextAction() {
			return
		}
		advance(w.c, 1)
	}
	t.Fatal("charge never opened the queue")
}

func TestSandstormAssaultCharge(t *testing.T) {
	c := makeCore(t)
	w := addDehya(t, c, 9)
	hits := w.watch()
	w.charge(t, nil)
	w.untilQueue(t)
	if math.Abs(c.Player.Stam-(player.MaxStam-12)) > 1e-6 {
		t.Fatalf("stamina %v", c.Player.Stam)
	}
	w.charge(t, nil)
	w.untilQueue(t)

	var cyclic []hit
	for _, h := range *hits {
		if h.info.Abil == "Charge Attack" {
			cyclic = append(cyclic, h)
		}
		if h.info.Abil == "Charge Attack (Finisher)" {
			t.Fatal("finisher landed before release")
		}
	}
	if len(cyclic) != 2 {
		t.Fatalf("cyclic hits %d", len(cyclic))
	}
	want := charge[8]
	for _, h := range cyclic {
		if h.info.Mult != want || h.info.PoiseDMG != 60 || h.info.Durability != 25 {
			t.Fatalf("cyclic %+v", h.info)
		}
		if h.info.AttackTag != attacks.AttackTagExtra || h.info.Element != attributes.Physical || h.info.StrikeType != attacks.StrikeTypeBlunt {
			t.Fatalf("class %+v", h.info)
		}
		if h.info.ICDTag != attacks.ICDTagNormalAttack || h.info.ICDGroup != attacks.ICDGroupDefault {
			t.Fatalf("icd %+v", h.info)
		}
	}

	w.charge(t, map[string]int{"final": 1})
	for range 80 {
		advance(c, 1)
		if len(*hits) > 2 && (*hits)[len(*hits)-1].info.Abil == "Charge Attack (Finisher)" {
			break
		}
	}
	last := (*hits)[len(*hits)-1].info
	if last.Abil != "Charge Attack (Finisher)" || last.Mult != chargeFinal[8] || last.PoiseDMG != 120 {
		t.Fatalf("finisher %+v", last)
	}
}

func TestSandstormAssaultChargeEndsAtFiveSeconds(t *testing.T) {
	c := makeCore(t)
	w := addDehya(t, c, 9)
	hits := w.watch()
	start := c.F
	w.charge(t, nil)
	var finisher int
	for c.F < start+tmpl.ClaymoreSpinDuration+180 {
		w.untilQueue(t)
		for _, h := range *hits {
			if h.info.Abil == "Charge Attack (Finisher)" {
				finisher = h.frame
			}
		}
		if finisher != 0 {
			break
		}
		w.charge(t, nil)
	}
	if finisher < start+tmpl.ClaymoreSpinDuration || finisher > start+tmpl.ClaymoreSpinDuration+90 {
		t.Fatalf("finisher frame %d start %d", finisher, start)
	}
}

func TestSandstormAssaultNormalAndBurstStillRun(t *testing.T) {
	c := makeCore(t)
	w := addDehya(t, c, 9)
	hits := w.watch()
	if err := c.Player.Exec(action.ActionAttack, keys.Dehya, nil); err != nil {
		t.Fatal(err)
	}
	w.untilQueue(t)
	if (*hits)[0].info.Abil != "Normal 0" || (*hits)[0].info.Mult != attack[0][8] {
		t.Fatalf("normal %+v", (*hits)[0].info)
	}

	c.Player.ByIndex(0).Energy = c.Player.ByIndex(0).EnergyMax
	if err := c.Player.Exec(action.ActionBurst, keys.Dehya, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 110)
	before := len(*hits)
	if err := c.Player.Exec(action.ActionCharge, keys.Dehya, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 40)
	for _, h := range (*hits)[before:] {
		if h.info.Abil == "Charge Attack" || h.info.Abil == "Charge Attack (Finisher)" {
			t.Fatalf("charge during burst: %+v", h.info)
		}
	}
	sawFist := false
	for _, h := range (*hits)[before:] {
		if h.info.Abil == "Flame-Mane's Fist" {
			sawFist = true
		}
	}
	if !sawFist {
		t.Fatal("burst charge did not punch")
	}
}
