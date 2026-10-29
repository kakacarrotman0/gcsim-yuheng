package xingqiu

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
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
	a := avatar.New(c, info.Point{}, 1)
	c.Combat.SetPlayer(a)
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

func addXingqiu(t *testing.T, c *core.Core, cons, maxLevel int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Xingqiu, testhelper.TestWeaponKey)
	prof.Base.Cons = cons
	prof.Base.Level = maxLevel
	prof.Base.MaxLevel = maxLevel
	prof.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	if _, err := c.AddChar(prof); err != nil {
		t.Fatal(err)
	}
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	return c.Player.ByIndex(0)
}

func asXingqiu(ch *character.CharWrapper) *char {
	return ch.Character.(*char)
}

func advance(c *core.Core, n int) {
	for range n {
		c.F++
		c.Tick()
	}
}

func skill(t *testing.T, c *core.Core) {
	t.Helper()
	if err := c.Player.Exec(action.ActionSkill, keys.Xingqiu, nil); err != nil {
		t.Fatal(err)
	}
}

func doBurst(t *testing.T, c *core.Core) {
	t.Helper()
	if err := c.Player.Exec(action.ActionBurst, keys.Xingqiu, nil); err != nil {
		t.Fatal(err)
	}
}

type heal struct {
	frame  int
	amount float64
	msg    string
}

func watchHeals(c *core.Core) *[]heal {
	var heals []heal
	c.Events.Subscribe(event.OnHeal, func(args ...any) {
		hi := args[0].(*info.HealInfo)
		heals = append(heals, heal{frame: c.F, amount: args[4].(float64), msg: hi.Message})
	}, "heals")
	return &heals
}

func hitPlayer(c *core.Core, poise, mult float64) {
	c.Events.Emit(event.OnPlayerHit, c.Player.Active(), &info.AttackEvent{
		Info: info.AttackInfo{Abil: "test hit", PoiseDMG: poise, Mult: mult},
	})
}

func TestRainSwordsHealOnExpiry(t *testing.T) {
	c := makeCore(t)
	ch := addXingqiu(t, c, 0, 40)
	if ch.Base.Ascension != 1 {
		t.Fatalf("ascension %d", ch.Base.Ascension)
	}
	heals := watchHeals(c)
	skill(t, c)
	if asXingqiu(ch).rainSwords != 3 {
		t.Fatalf("swords %d", asXingqiu(ch).rainSwords)
	}
	advance(c, 853)
	if len(*heals) != 0 {
		t.Fatalf("early heals %+v", *heals)
	}
	advance(c, 988-853)
	want := 0.06 * ch.MaxHP()
	if len(*heals) != 3 {
		t.Fatalf("heals %+v", *heals)
	}
	for _, h := range *heals {
		if h.frame != 988 || h.msg != "Hydropathic" || math.Abs(h.amount-want) > 1e-6 {
			t.Fatalf("heal %+v want %v at 988", h, want)
		}
	}
	if asXingqiu(ch).rainSwords != 0 {
		t.Fatalf("swords left %d", asXingqiu(ch).rainSwords)
	}
}

func TestC1AddsAFourthExpiryHeal(t *testing.T) {
	c := makeCore(t)
	ch := addXingqiu(t, c, 1, 40)
	heals := watchHeals(c)
	skill(t, c)
	if asXingqiu(ch).rainSwords != 4 {
		t.Fatalf("swords %d", asXingqiu(ch).rainSwords)
	}
	advance(c, 988)
	if len(*heals) != 4 {
		t.Fatalf("heals %d", len(*heals))
	}
	want := 0.06 * ch.MaxHP()
	for _, h := range *heals {
		if math.Abs(h.amount-want) > 1e-6 {
			t.Fatalf("amount %v want %v", h.amount, want)
		}
	}
}

func TestNoHealBeforeA1(t *testing.T) {
	c := makeCore(t)
	ch := addXingqiu(t, c, 0, 20)
	if ch.Base.Ascension != 0 {
		t.Fatalf("ascension %d", ch.Base.Ascension)
	}
	heals := watchHeals(c)
	var orbitals int
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		if args[1].(*info.AttackEvent).Info.Abil == "Xingqiu Orbital" {
			orbitals++
		}
	}, "orbital")
	skill(t, c)
	hitPlayer(c, 100, 1)
	advance(c, 988)
	if len(*heals) != 0 {
		t.Fatalf("heals %+v", *heals)
	}
	if orbitals == 0 {
		t.Fatal("orbital application stopped")
	}
}

func TestShatterHealsOneSwordAndRespectsICD(t *testing.T) {
	c := makeCore(t)
	ch := addXingqiu(t, c, 0, 40)
	heals := watchHeals(c)
	skill(t, c)
	advance(c, 50)
	hitPlayer(c, 0, 0)
	if len(*heals) != 0 || asXingqiu(ch).rainSwords != 3 {
		t.Fatalf("zero hit heals %d swords %d", len(*heals), asXingqiu(ch).rainSwords)
	}
	hitPlayer(c, 50, 0)
	if len(*heals) != 1 || asXingqiu(ch).rainSwords != 2 {
		t.Fatalf("after hit heals %d swords %d", len(*heals), asXingqiu(ch).rainSwords)
	}
	hitPlayer(c, 50, 0)
	if len(*heals) != 1 || asXingqiu(ch).rainSwords != 2 {
		t.Fatalf("icd heals %d swords %d", len(*heals), asXingqiu(ch).rainSwords)
	}
	advance(c, a1ShatterICD+1)
	hitPlayer(c, 50, 0)
	if len(*heals) != 2 || asXingqiu(ch).rainSwords != 1 {
		t.Fatalf("second shatter heals %d swords %d", len(*heals), asXingqiu(ch).rainSwords)
	}
	advance(c, 988-c.F)
	if len(*heals) != 3 {
		t.Fatalf("expiry should heal the last sword, heals %+v", *heals)
	}
	if math.Abs((*heals)[2].amount-0.06*ch.MaxHP()) > 1e-6 {
		t.Fatalf("last %+v", (*heals)[2])
	}
}

func TestBurstBlocksShatterAndRefillsSwords(t *testing.T) {
	c := makeCore(t)
	ch := addXingqiu(t, c, 0, 40)
	heals := watchHeals(c)
	skill(t, c)
	advance(c, 40)
	hitPlayer(c, 80, 0)
	if asXingqiu(ch).rainSwords != 2 || len(*heals) != 1 {
		t.Fatalf("setup swords %d heals %d", asXingqiu(ch).rainSwords, len(*heals))
	}
	doBurst(t, c)
	if asXingqiu(ch).rainSwords != 3 {
		t.Fatalf("burst swords %d", asXingqiu(ch).rainSwords)
	}
	hitPlayer(c, 80, 1)
	if len(*heals) != 1 || asXingqiu(ch).rainSwords != 3 {
		t.Fatalf("burst shatter heals %d swords %d", len(*heals), asXingqiu(ch).rainSwords)
	}
}
