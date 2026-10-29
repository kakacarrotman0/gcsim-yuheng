package illuga

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
	"github.com/genshinsim/gcsim/pkg/modifier"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestWeapon()
}

func TestSkillFlatDamageUsesTheHitmarkSnapshot(t *testing.T) {
	cases := []struct {
		name     string
		hold     int
		hitmark  int
		emRatio  float64
		defRatio float64
		abil     string
	}{
		{
			name:     "tap",
			hitmark:  skillTapHitmark,
			emRatio:  4.8256,
			defRatio: 2.4128,
			abil:     "Dawnbearing Songbird (Tap)",
		},
		{
			name:     "hold",
			hold:     1,
			hitmark:  skillHoldHitmark,
			emRatio:  6.032,
			defRatio: 3.016,
			abil:     "Dawnbearing Songbird (Hold)",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, ch := makeCore(t)
			em := ch.Stat(attributes.EM)
			def := ch.TotalDef(false)

			var hits []info.AttackInfo
			c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
				hits = append(hits, args[1].(*info.AttackEvent).Info)
			}, "hits")

			param := map[string]int{}
			if tc.hold == 1 {
				param["hold"] = 1
			}
			if err := c.Player.Exec(action.ActionSkill, keys.Illuga, param); err != nil {
				t.Fatal(err)
			}

			advance(c, 10)
			const (
				lateEM  = 1000.0
				lateDEF = 400.0
			)
			bonus := make([]float64, attributes.EndStatType)
			bonus[attributes.EM] = lateEM
			bonus[attributes.DEF] = lateDEF
			ch.AddStatMod(character.StatMod{
				Base: modifier.NewBase("illuga-late-snapshot", -1),
				Amount: func() []float64 {
					return bonus
				},
			})
			advance(c, tc.hitmark-10)

			var got []info.AttackInfo
			for _, hit := range hits {
				if hit.Abil == tc.abil {
					got = append(got, hit)
				}
			}
			if len(got) != 1 {
				t.Fatalf("skill hits %d", len(got))
			}
			if c.F != tc.hitmark {
				t.Fatalf("hit frame %d", c.F)
			}
			want := tc.emRatio*(em+lateEM) + tc.defRatio*(def+lateDEF)
			if math.Abs(got[0].FlatDmg-want) > 1e-6 {
				t.Fatalf("flat damage %v, want %v from the hitmark EM and DEF", got[0].FlatDmg, want)
			}
		})
	}
}

func makeCore(t *testing.T) (*core.Core, *character.CharWrapper) {
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

	prof := testhelper.DefaultProfile(keys.Illuga, testhelper.TestWeaponKey)
	prof.Base.Cons = 0
	prof.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	if _, err := c.AddChar(prof); err != nil {
		t.Fatal(err)
	}
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	return c, c.Player.ByIndex(0)
}

func advance(c *core.Core, n int) {
	for range n {
		c.F++
		c.Tick()
	}
}
