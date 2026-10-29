package razor

import (
	"errors"
	"math"
	"strings"
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

func addRazor(t *testing.T, c *core.Core) *characterWait {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Razor, testhelper.TestWeaponKey)
	prof.Base.Level = 90
	prof.Base.MaxLevel = 90
	prof.Talents = info.TalentProfile{Attack: 9, Skill: 9, Burst: 9}
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
	if err := w.c.Player.Exec(action.ActionCharge, keys.Razor, p); err != nil {
		t.Fatal(err)
	}
}

func (w *characterWait) untilQueue(t *testing.T) {
	t.Helper()
	for range 800 {
		if w.c.Player.CanQueueNextAction() {
			return
		}
		advance(w.c, 1)
	}
	t.Fatal("action never opened the queue")
}

func TestSteelFangCharge(t *testing.T) {
	c := makeCore(t)
	w := addRazor(t, c)
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
	for _, h := range *hits {
		if strings.HasPrefix(h.info.Abil, "The Wolf Within") {
			t.Fatal("wolf companion followed a charged attack")
		}
	}
	if cyclic[1].frame-cyclic[0].frame >= tmpl.ClaymoreSpinWindup {
		t.Fatalf("chained slash kept the windup: %d then %d", cyclic[0].frame, cyclic[1].frame)
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

func TestSteelFangChargeEndsAtFiveSeconds(t *testing.T) {
	c := makeCore(t)
	w := addRazor(t, c)
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

func TestSteelFangChargeEndsWhenStaminaRunsOut(t *testing.T) {
	c := makeCore(t)
	w := addRazor(t, c)
	hits := w.watch()
	start := c.F
	w.charge(t, map[string]int{"no_limit": 1})
	var finisher int
	for c.F < start+2000 {
		w.untilQueue(t)
		for _, h := range *hits {
			if h.info.Abil == "Charge Attack (Finisher)" {
				finisher = h.frame
			}
		}
		if finisher != 0 {
			break
		}
		w.charge(t, map[string]int{"no_limit": 1})
	}
	if finisher == 0 || finisher <= start+tmpl.ClaymoreSpinDuration {
		t.Fatalf("finisher frame %d start %d", finisher, start)
	}
	if c.Player.Stam != 0 {
		t.Fatalf("stamina %v", c.Player.Stam)
	}
}

func TestSteelFangChargeCancelDropsTheFinisher(t *testing.T) {
	c := makeCore(t)
	w := addRazor(t, c)
	hits := w.watch()
	start := c.F
	w.charge(t, nil)
	w.untilQueue(t)
	if (*hits)[0].frame-start != tmpl.ClaymoreSpinWindup+tmpl.ClaymoreSpinHit {
		t.Fatalf("first hit %d", (*hits)[0].frame-start)
	}
	if err := c.Player.Exec(action.ActionDash, keys.Razor, nil); err != nil {
		t.Fatal(err)
	}
	w.untilQueue(t)
	advance(c, 40)
	for _, h := range *hits {
		if h.info.Abil == "Charge Attack (Finisher)" {
			t.Fatal("cancelled spin still finished")
		}
	}
	start = c.F
	w.charge(t, nil)
	w.untilQueue(t)
	var second int
	for _, h := range *hits {
		if h.info.Abil == "Charge Attack" && h.frame >= start {
			second = h.frame - start
		}
	}
	if second != tmpl.ClaymoreSpinWindup+tmpl.ClaymoreSpinHit {
		t.Fatalf("charge after cancel queued in %d frames", second)
	}
}

func TestLightningFangDisablesCharge(t *testing.T) {
	c := makeCore(t)
	w := addRazor(t, c)
	hits := w.watch()
	c.Player.ByIndex(0).Energy = c.Player.ByIndex(0).EnergyMax
	if err := c.Player.Exec(action.ActionBurst, keys.Razor, nil); err != nil {
		t.Fatal(err)
	}
	w.untilQueue(t)
	if got := c.Player.ByIndex(0).Stat(attributes.AtkSpd); got != burstATKSpeed[8] {
		t.Fatalf("burst attack speed %v", got)
	}
	err := c.Player.ByIndex(0).NextQueueItemIsValid(keys.Razor, action.ActionCharge, nil)
	if !errors.Is(err, errLightningFangCharge) {
		t.Fatalf("queue check: %v", err)
	}
	err = c.Player.Exec(action.ActionCharge, keys.Razor, nil)
	if !errors.Is(err, errLightningFangCharge) {
		t.Fatalf("exec: %v", err)
	}
	advance(c, 5)
	for _, h := range *hits {
		if h.info.Abil == "Charge Attack" || h.info.Abil == "Charge Attack (Finisher)" {
			t.Fatalf("charge during Lightning Fang: %+v", h.info)
		}
	}

	if err := c.Player.Exec(action.ActionAttack, keys.Razor, nil); err != nil {
		t.Fatal(err)
	}
	w.untilQueue(t)
	advance(c, 3)
	sawNormal := false
	sawWolf := false
	for _, h := range *hits {
		if h.info.Abil == "Normal 0" {
			sawNormal = true
		}
		if h.info.Abil == "The Wolf Within 0" {
			sawWolf = true
			if h.info.Element != attributes.Electro || h.info.AttackTag != attacks.AttackTagElementalBurst {
				t.Fatalf("wolf %+v", h.info)
			}
		}
	}
	if !sawNormal || !sawWolf {
		t.Fatalf("normal %v wolf %v", sawNormal, sawWolf)
	}
}
