package diluc

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

func addDiluc(t *testing.T, c *core.Core, maxLevel, talent int) {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Diluc, testhelper.TestWeaponKey)
	prof.Base.Level = maxLevel
	prof.Base.MaxLevel = maxLevel
	prof.Talents = info.TalentProfile{Attack: talent, Skill: talent, Burst: talent}
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

func doCharge(t *testing.T, c *core.Core, p map[string]int) {
	t.Helper()
	if err := c.Player.Exec(action.ActionCharge, keys.Diluc, p); err != nil {
		t.Fatal(err)
	}
}

func untilQueue(t *testing.T, c *core.Core) {
	t.Helper()
	for range 800 {
		if c.Player.CanQueueNextAction() {
			return
		}
		advance(c, 1)
	}
	t.Fatal("action never opened the queue")
}

func TestTemperedSwordCharge(t *testing.T) {
	c := makeCore(t)
	addDiluc(t, c, 90, 9)
	if c.Player.ByIndex(0).Base.Ascension < 1 {
		t.Fatal("expected ascension 1 or higher")
	}
	hits := watch(c)
	doCharge(t, c, nil)
	untilQueue(t, c)
	if math.Abs(c.Player.Stam-(player.MaxStam-6)) > 1e-6 {
		t.Fatalf("stamina %v", c.Player.Stam)
	}
	doCharge(t, c, nil)
	untilQueue(t, c)
	doCharge(t, c, map[string]int{"final": 1})
	for range 80 {
		advance(c, 1)
		if len(*hits) > 0 && (*hits)[len(*hits)-1].info.Abil == "Charge Attack (Finisher)" {
			break
		}
	}
	var cyclic, finisher int
	for _, h := range *hits {
		switch h.info.Abil {
		case "Charge Attack":
			cyclic++
			if h.info.Mult != charge[8] || h.info.PoiseDMG != 60 || h.info.Element != attributes.Physical {
				t.Fatalf("cyclic %+v", h.info)
			}
			if h.info.AttackTag != attacks.AttackTagExtra || h.info.Durability != 25 || h.info.StrikeType != attacks.StrikeTypeBlunt {
				t.Fatalf("class %+v", h.info)
			}
		case "Charge Attack (Finisher)":
			finisher++
			if h.info.Mult != chargeFinal[8] || h.info.PoiseDMG != 120 || h.info.Element != attributes.Physical {
				t.Fatalf("finisher %+v", h.info)
			}
		}
	}
	if cyclic != 2 || finisher != 1 {
		t.Fatalf("cyclic %d finisher %d", cyclic, finisher)
	}
}

func TestRelentlessStaminaAndDuration(t *testing.T) {
	plainDrop, plainEnd := spinUntilFinisher(t, 20)
	buffedDrop, buffedEnd := spinUntilFinisher(t, 40)
	if plainDrop != 12 || buffedDrop != 6 {
		t.Fatalf("stamina drop plain %v buffed %v", plainDrop, buffedDrop)
	}
	if plainEnd < tmpl.ClaymoreSpinDuration || plainEnd > tmpl.ClaymoreSpinDuration+90 {
		t.Fatalf("a0 cap landed at %d", plainEnd)
	}
	if buffedEnd < tmpl.ClaymoreSpinDuration+3*60 || buffedEnd > tmpl.ClaymoreSpinDuration+3*60+120 {
		t.Fatalf("a1 cap landed at %d", buffedEnd)
	}
}

func spinUntilFinisher(t *testing.T, maxLevel int) (float64, int) {
	t.Helper()
	c := makeCore(t)
	addDiluc(t, c, maxLevel, 9)
	asc := c.Player.ByIndex(0).Base.Ascension
	if maxLevel == 20 && asc != 0 {
		t.Fatalf("ascension %d", asc)
	}
	if maxLevel == 40 && asc != 1 {
		t.Fatalf("ascension %d", asc)
	}
	hits := watch(c)
	startStam := c.Player.Stam
	start := c.F
	doCharge(t, c, nil)
	untilQueue(t, c)
	drop := startStam - c.Player.Stam
	var finisher int
	limit := tmpl.ClaymoreSpinDuration + 3*60 + 180
	for c.F < start+limit {
		untilQueue(t, c)
		for i := range *hits {
			h := &(*hits)[i]
			if h.info.Abil == "Charge Attack (Finisher)" {
				finisher = h.frame
			}
		}
		if finisher != 0 {
			break
		}
		doCharge(t, c, nil)
	}
	if finisher == 0 {
		t.Fatal("no finisher")
	}
	return drop, finisher - start
}

func TestDawnInfusesCharge(t *testing.T) {
	c := makeCore(t)
	addDiluc(t, c, 90, 9)
	hits := watch(c)
	doCharge(t, c, map[string]int{"final": 1})
	for range 120 {
		advance(c, 1)
		if len(*hits) > 0 {
			break
		}
	}
	if (*hits)[0].info.Element != attributes.Physical {
		t.Fatalf("before burst %+v", (*hits)[0].info)
	}

	c.Player.ByIndex(0).Energy = c.Player.ByIndex(0).EnergyMax
	if err := c.Player.Exec(action.ActionBurst, keys.Diluc, nil); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c)
	before := len(*hits)
	doCharge(t, c, nil)
	untilQueue(t, c)
	var saw bool
	for _, h := range (*hits)[before:] {
		if h.info.Abil != "Charge Attack" {
			continue
		}
		saw = true
		if h.info.Element != attributes.Pyro || h.info.AttackTag != attacks.AttackTagExtra {
			t.Fatalf("infused %+v", h.info)
		}
	}
	if !saw {
		t.Fatal("no charged hit during Dawn")
	}
}
