package fischl

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

func addFischl(t *testing.T, c *core.Core, maxLevel int, params map[string]int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Fischl, testhelper.TestWeaponKey)
	prof.Base.Level = maxLevel
	prof.Base.MaxLevel = maxLevel
	prof.Talents = info.TalentProfile{Attack: 9, Skill: 9, Burst: 9}
	for k, v := range params {
		prof.Params[k] = v
	}
	if _, err := c.AddChar(prof); err != nil {
		t.Fatal(err)
	}
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	return c.Player.ByIndex(0)
}

type landed struct {
	frame int
	info  info.AttackInfo
}

func watch(c *core.Core) *[]landed {
	var hits []landed
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hits = append(hits, landed{frame: c.F, info: args[1].(*info.AttackEvent).Info})
	}, "hits")
	return &hits
}

func advance(c *core.Core, n int) {
	for range n {
		c.F++
		c.Tick()
	}
}

func playerBolts(hits []landed) []landed {
	var out []landed
	for i := range hits {
		if hits[i].info.Abil == "Oz (Skill)" || hits[i].info.Abil == "Oz (Summon)" {
			continue
		}
		out = append(out, hits[i])
	}
	return out
}

func doAim(t *testing.T, c *core.Core, p map[string]int) {
	t.Helper()
	if err := c.Player.Exec(action.ActionAim, keys.Fischl, p); err != nil {
		t.Fatal(err)
	}
}

func summonOz(t *testing.T, c *core.Core) {
	t.Helper()
	if err := c.Player.Exec(action.ActionSkill, keys.Fischl, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillOzSpawn)
}

func TestAimedShotUnchangedWithoutOz(t *testing.T) {
	c := makeCore(t)
	addFischl(t, c, 90, nil)
	hits := watch(c)
	doAim(t, c, nil)
	advance(c, aimedHitmarks[attacks.AimParamLv1]+10)
	if len(*hits) != 1 {
		t.Fatalf("hits %d", len(*hits))
	}
	got := (*hits)[0]
	if got.frame != aimedHitmarks[attacks.AimParamLv1]+10 || got.info.Abil != "Fully-Charged Aimed Shot" {
		t.Fatalf("hit %+v", got)
	}
	if math.Abs(got.info.Mult-fullaim[8]) > 1e-9 || got.info.Element != attributes.Electro {
		t.Fatalf("arrow %+v", got.info)
	}
}

func TestA1BoltReplacesTheArrow(t *testing.T) {
	c := makeCore(t)
	ch := addFischl(t, c, 90, map[string]int{"oz_travel": 30})
	hits := watch(c)
	summonOz(t, c)
	if !ch.StatusIsActive(ozActiveKey) {
		t.Fatal("oz did not spawn")
	}
	start := c.F
	doAim(t, c, map[string]int{"oz": 1, "weakspot": 1, "travel": 10})
	launch := start + aimedHitmarks[attacks.AimParamLv1] + 10
	advance(c, launch-c.F)
	if bolts := playerBolts(*hits); len(bolts) != 0 {
		t.Fatalf("bolt should still be in flight, hits %+v", bolts)
	}
	advance(c, 30)
	bolt := playerBolts(*hits)
	if len(bolt) != 1 {
		t.Fatalf("bolt hits %+v all %+v", bolt, *hits)
	}
	got := bolt[0]
	if got.frame != launch+30 {
		t.Fatalf("frame %d want %d", got.frame, launch+30)
	}
	want := fullaim[8] * a1RetributionRatio
	if math.Abs(got.info.Mult-want) > 1e-9 || math.Abs(want-2.108*1.527) > 1e-9 {
		t.Fatalf("mult %v want %v", got.info.Mult, want)
	}
	if got.info.AttackTag != attacks.AttackTagExtra || got.info.Element != attributes.Electro {
		t.Fatalf("tag %v ele %v", got.info.AttackTag, got.info.Element)
	}
	if got.info.Durability != 25 || got.info.PoiseDMG != 40 || !got.info.HitWeakPoint {
		t.Fatalf("gauge %v poise %v weak %v", got.info.Durability, got.info.PoiseDMG, got.info.HitWeakPoint)
	}
	if got.info.Abil == "Fully-Charged Aimed Shot" {
		t.Fatal("arrow also hit the enemy")
	}
}

func TestA1RequiresOzAscensionAndFullCharge(t *testing.T) {
	c := makeCore(t)
	addFischl(t, c, 20, nil)
	if c.Player.ByIndex(0).Base.Ascension != 0 {
		t.Fatalf("ascension %d", c.Player.ByIndex(0).Base.Ascension)
	}
	hits := watch(c)
	summonOz(t, c)
	start := c.F
	doAim(t, c, map[string]int{"oz": 1})
	advance(c, start+aimedHitmarks[attacks.AimParamLv1]+10-c.F)
	found := false
	for _, h := range *hits {
		if h.info.Abil == "Thundering Retribution (A1)" {
			t.Fatal("a1 fired below ascension 1")
		}
		if h.info.Abil == "Fully-Charged Aimed Shot" {
			found = true
			if h.frame != start+aimedHitmarks[attacks.AimParamLv1]+10 {
				t.Fatalf("frame %d", h.frame)
			}
		}
	}
	if !found {
		t.Fatalf("missing arrow %+v", *hits)
	}

	c = makeCore(t)
	addFischl(t, c, 90, nil)
	hits = watch(c)
	summonOz(t, c)
	start = c.F
	doAim(t, c, map[string]int{"oz": 1, "hold": attacks.AimParamPhys})
	advance(c, start+aimedHitmarks[attacks.AimParamPhys]+10-c.F)
	found = false
	for _, h := range *hits {
		if h.info.Abil == "Thundering Retribution (A1)" {
			t.Fatal("physical shot triggered a1")
		}
		if h.info.Abil == "Aimed Shot" && h.info.Element == attributes.Physical && math.Abs(h.info.Mult-aim[8]) < 1e-9 {
			found = true
		}
	}
	if !found {
		t.Fatalf("physical arrow missing %+v", *hits)
	}
}

func TestA1DropsWhenOzExpiresBeforeTheArrowArrives(t *testing.T) {
	c := makeCore(t)
	ch := addFischl(t, c, 90, nil)
	hits := watch(c)
	summonOz(t, c)
	start := c.F
	doAim(t, c, map[string]int{"oz": 1, "travel": 10})
	advance(c, 1)
	ch.DeleteStatus(ozActiveKey)
	launch := start + aimedHitmarks[attacks.AimParamLv1] + 10
	advance(c, launch-c.F)
	found := false
	for _, h := range *hits {
		if h.info.Abil == "Thundering Retribution (A1)" {
			t.Fatal("bolt fired after oz expired")
		}
		if h.info.Abil == "Fully-Charged Aimed Shot" {
			found = true
			if h.frame != launch {
				t.Fatalf("fallback frame %d want %d", h.frame, launch)
			}
		}
	}
	if !found {
		t.Fatalf("fallback missing %+v", *hits)
	}
}
