package beidou

import (
	"math"
	"strings"
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

func addBeidou(t *testing.T, c *core.Core, maxLevel, talent int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Beidou, testhelper.TestWeaponKey)
	prof.Base.Level = maxLevel
	prof.Base.MaxLevel = maxLevel
	prof.Talents = info.TalentProfile{Attack: talent, Skill: talent, Burst: talent}
	if _, err := c.AddChar(prof); err != nil {
		t.Fatal(err)
	}
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	return c.Player.ByIndex(0)
}

func advance(c *core.Core, n int) {
	for range n {
		c.F++
		c.Tick()
	}
}

func TestCounterIsTheA1Maximum(t *testing.T) {
	// Talent 9 is the KQM row: 206.72%, 478.72%, 750.72% for 0, 1, and 2 stacks.
	// A perfect parry is the 2-stack row, including 4 particles.
	cases := []struct {
		counter int
		mult    float64
		parts   float64
	}{
		{0, 2.0672, 2},
		{1, 4.7872, 3},
		{2, 7.5072, 4},
		{9, 7.5072, 4},
	}
	for _, tc := range cases {
		c := makeCore(t)
		ch := addBeidou(t, c, 40, 9)
		if ch.Base.Ascension != 1 {
			t.Fatalf("ascension %d", ch.Base.Ascension)
		}
		var mults []float64
		c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
			ai := args[1].(*info.AttackEvent).Info
			if strings.HasPrefix(ai.Abil, "Tidecaller") {
				mults = append(mults, ai.Mult)
			}
		}, "hits")
		var particles []float64
		c.Events.Subscribe(event.OnParticleReceived, func(args ...any) {
			particles = append(particles, args[0].(character.Particle).Num)
		}, "particles")
		if err := c.Player.Exec(action.ActionSkill, keys.Beidou, map[string]int{"counter": tc.counter}); err != nil {
			t.Fatal(err)
		}
		lvl := ch.TalentLvlSkill()
		stacks := tc.counter
		if stacks > 2 {
			stacks = 2
		}
		want := skillbase[lvl] + skillbonus[lvl]*float64(stacks)
		advance(c, skillHitmark+ch.ParticleDelay)
		if len(mults) != 1 || math.Abs(mults[0]-want) > 1e-9 || math.Abs(mults[0]-tc.mult) > 1e-9 {
			t.Fatalf("counter %d mult %v want %v and %v", tc.counter, mults, want, tc.mult)
		}
		if len(particles) != 1 || particles[0] != tc.parts {
			t.Fatalf("counter %d particles %v", tc.counter, particles)
		}
		if math.Abs(ch.Stat(attributes.AtkSpd)) > 1e-9 {
			t.Fatalf("a4 atk speed at ascension 1: %v", ch.Stat(attributes.AtkSpd))
		}
	}
}

func TestA4FollowsMaximumCounterOnly(t *testing.T) {
	c := makeCore(t)
	ch := addBeidou(t, c, 90, 9)
	if ch.Base.Ascension < 4 {
		t.Fatalf("ascension %d", ch.Base.Ascension)
	}
	if err := c.Player.Exec(action.ActionSkill, keys.Beidou, map[string]int{"counter": 1}); err != nil {
		t.Fatal(err)
	}
	if math.Abs(ch.Stat(attributes.AtkSpd)) > 1e-9 {
		t.Fatalf("one stack armed a4: %v", ch.Stat(attributes.AtkSpd))
	}

	c = makeCore(t)
	ch = addBeidou(t, c, 90, 9)
	if err := c.Player.Exec(action.ActionSkill, keys.Beidou, map[string]int{"counter": 2}); err != nil {
		t.Fatal(err)
	}
	if math.Abs(ch.Stat(attributes.AtkSpd)-0.15) > 1e-9 {
		t.Fatalf("a4 atk speed %v", ch.Stat(attributes.AtkSpd))
	}
}
