package scrolloftheheroofcindercity

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/kaeya"
	_ "github.com/genshinsim/gcsim/internal/characters/lisa"
	_ "github.com/genshinsim/gcsim/internal/characters/sandrone"
	_ "github.com/genshinsim/gcsim/internal/characters/sucrose"
	_ "github.com/genshinsim/gcsim/internal/characters/xiangling"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestWeapon()
}

func TestRealStellarSwirlGrantsAnemoAndCryoDamage(t *testing.T) {
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

	wearer := testhelper.DefaultProfile(keys.Sucrose, testhelper.TestWeaponKey)
	wearer.Sets[keys.ScrollOfTheHeroOfCinderCity] = 4
	if _, err := c.AddChar(wearer); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AddChar(testhelper.DefaultProfile(keys.Sandrone, testhelper.TestWeaponKey)); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	before := partyBonus(c)
	e.AttachOrRefill(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25, ActorIndex: 1},
	})
	e.React(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Anemo, Durability: 25, ActorIndex: 0},
	})
	for i, name := range []string{"sucrose", "sandrone"} {
		ch := c.Player.ByIndex(i)
		if got := ch.Stat(attributes.AnemoP) - before[i][attributes.AnemoP]; math.Abs(got-0.12) > 1e-9 {
			t.Fatalf("%s anemo damage bonus %v", name, got)
		}
		if got := ch.Stat(attributes.CryoP) - before[i][attributes.CryoP]; math.Abs(got-0.12) > 1e-9 {
			t.Fatalf("%s cryo damage bonus %v", name, got)
		}
	}
}

func TestStellarSwirlGrantsAnemoAndCryoDamage(t *testing.T) {
	c, trg := makeParty(t, keys.Sucrose, keys.Kaeya, keys.ScrollOfTheHeroOfCinderCity)
	before := partyBonus(c)
	emit(c, trg, event.OnStellarSwirl, 0)

	for i, name := range []string{"sucrose", "kaeya"} {
		ch := c.Player.ByIndex(i)
		if got := ch.Stat(attributes.AnemoP) - before[i][attributes.AnemoP]; math.Abs(got-0.12) > 1e-9 {
			t.Fatalf("%s anemo damage bonus %v", name, got)
		}
		if got := ch.Stat(attributes.CryoP) - before[i][attributes.CryoP]; math.Abs(got-0.12) > 1e-9 {
			t.Fatalf("%s cryo damage bonus %v", name, got)
		}
		if got := ch.Stat(attributes.ElectroP) - before[i][attributes.ElectroP]; math.Abs(got) > 1e-9 {
			t.Fatalf("%s electro damage bonus %v", name, got)
		}
		if got := ch.Stat(attributes.PyroP) - before[i][attributes.PyroP]; math.Abs(got) > 1e-9 {
			t.Fatalf("%s pyro damage bonus %v", name, got)
		}
	}
}

func TestStellarSwirlIgnoresAnUnrelatedWearer(t *testing.T) {
	c, trg := makeParty(t, keys.Xiangling, keys.Kaeya, keys.ScrollOfTheHeroOfCinderCity)
	emit(c, trg, event.OnStellarSwirl, 0)
	if got := c.Player.ByIndex(0).Stat(attributes.AnemoP); math.Abs(got) > 1e-9 {
		t.Fatalf("pyro wearer anemo damage %v", got)
	}
	if got := c.Player.ByIndex(1).Stat(attributes.CryoP); math.Abs(got) > 1e-9 {
		t.Fatalf("ally cryo damage %v", got)
	}
}

func TestStellarConductStillGrantsElectroAndCryoDamage(t *testing.T) {
	c, trg := makeParty(t, keys.Lisa, keys.Kaeya, keys.ScrollOfTheHeroOfCinderCity)
	before := partyBonus(c)
	emit(c, trg, event.OnStellarConduct, 0)
	for i, name := range []string{"lisa", "kaeya"} {
		ch := c.Player.ByIndex(i)
		if got := ch.Stat(attributes.ElectroP) - before[i][attributes.ElectroP]; math.Abs(got-0.12) > 1e-9 {
			t.Fatalf("%s electro damage bonus %v", name, got)
		}
		if got := ch.Stat(attributes.CryoP) - before[i][attributes.CryoP]; math.Abs(got-0.12) > 1e-9 {
			t.Fatalf("%s cryo damage bonus %v", name, got)
		}
	}
}

func partyBonus(c *core.Core) [][attributes.EndStatType]float64 {
	out := make([][attributes.EndStatType]float64, len(c.Player.Chars()))
	for i, ch := range c.Player.Chars() {
		out[i][attributes.AnemoP] = ch.Stat(attributes.AnemoP)
		out[i][attributes.CryoP] = ch.Stat(attributes.CryoP)
		out[i][attributes.ElectroP] = ch.Stat(attributes.ElectroP)
		out[i][attributes.PyroP] = ch.Stat(attributes.PyroP)
	}
	return out
}

func emit(c *core.Core, trg *enemy.Enemy, evt event.Event, actor int) {
	ae := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: actor}}
	c.Events.Emit(evt, trg, ae)
}

func makeParty(t *testing.T, wearer, ally keys.Char, set keys.Set) (*core.Core, *enemy.Enemy) {
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

	wp := testhelper.DefaultProfile(wearer, testhelper.TestWeaponKey)
	wp.Sets[set] = 4
	al := testhelper.DefaultProfile(ally, testhelper.TestWeaponKey)
	if _, err := c.AddChar(wp); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AddChar(al); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	return c, e
}
