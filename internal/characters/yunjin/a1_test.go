package yunjin

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

func makeCore(t *testing.T, seed int64) *core.Core {
	t.Helper()
	c, err := core.New(core.Opt{Seed: seed})
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

func addYunjin(t *testing.T, c *core.Core, cons, maxLevel int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.YunJin, testhelper.TestWeaponKey)
	prof.Base.Cons = cons
	prof.Base.Level = maxLevel
	prof.Base.MaxLevel = maxLevel
	prof.Talents = info.TalentProfile{Attack: 9, Skill: 9, Burst: 9}
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

func particlesOf(c *core.Core) *[]float64 {
	var nums []float64
	c.Events.Subscribe(event.OnParticleReceived, func(args ...any) {
		nums = append(nums, args[0].(character.Particle).Num)
	}, "particles")
	return &nums
}

func TestOpeningFlourishVariants(t *testing.T) {
	cases := []struct {
		name  string
		param map[string]int
		level int
		abil  string
		mult  float64
		dur   info.Durability
		poise float64
		parts float64
		hit   int
	}{
		{"tap", nil, 0, "Opening Flourish (Press)", 2.53504, 50, 100, 2, 13},
		{"hold1", map[string]int{"hold": 1}, 1, "Opening Flourish (Level 1)", 4.43632, 50, 200, -1, 50},
		{"hold2", map[string]int{"hold": 2}, 2, "Opening Flourish (Level 2)", 6.3376, 100, 300, 3, 92},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := makeCore(t, 1)
			ch := addYunjin(t, c, 0, 90)
			hits := watch(c)
			parts := particlesOf(c)
			act, err := ch.Character.(*char).Skill(tc.param)
			if err != nil {
				t.Fatal(err)
			}
			if tc.name == "tap" && act.Frames(action.ActionJump) != skillFrames[0][action.ActionJump] {
				t.Fatalf("tap jump %d", act.Frames(action.ActionJump))
			}
			if tc.name != "tap" && act.Frames(action.ActionDash) != skillFrames[tc.level][action.ActionDash] {
				t.Fatalf("dash %d", act.Frames(action.ActionDash))
			}
			if act.AnimationLength != skillFrames[tc.level][action.InvalidAction] {
				t.Fatalf("animation %d", act.AnimationLength)
			}
			advance(c, tc.hit+ch.ParticleDelay)
			if len(*hits) != 1 {
				t.Fatalf("hits %+v", *hits)
			}
			got := (*hits)[0]
			if got.frame != tc.hit || got.info.Abil != tc.abil {
				t.Fatalf("hit %+v", got)
			}
			if !got.info.UseDef || got.info.Element != attributes.Geo || got.info.Durability != tc.dur || got.info.PoiseDMG != tc.poise {
				t.Fatalf("attack %+v", got.info)
			}
			lvl := ch.TalentLvlSkill()
			if math.Abs(got.info.Mult-skillDmg[tc.level][lvl]) > 1e-9 || math.Abs(got.info.Mult-tc.mult) > 1e-9 {
				t.Fatalf("mult %v want %v", got.info.Mult, tc.mult)
			}
			if tc.parts > 0 && (len(*parts) != 1 || (*parts)[0] != tc.parts) {
				t.Fatalf("particles %v", *parts)
			}
		})
	}
}

func TestPerfectCounterRequiresAscension(t *testing.T) {
	c := makeCore(t, 1)
	ch := addYunjin(t, c, 0, 20)
	if ch.Base.Ascension != 0 {
		t.Fatalf("ascension %d", ch.Base.Ascension)
	}
	hits := watch(c)
	act, err := ch.Character.(*char).Skill(map[string]int{"perfect": 1})
	if err != nil {
		t.Fatal(err)
	}
	if act.Frames(action.ActionJump) != skillFrames[0][action.ActionJump] {
		t.Fatalf("a0 perfect changed press frames to %d", act.Frames(action.ActionJump))
	}
	advance(c, skillHitmarks[0])
	if len(*hits) != 1 || (*hits)[0].info.Abil != "Opening Flourish (Press)" || (*hits)[0].info.Durability != 50 {
		t.Fatalf("a0 perfect %+v", *hits)
	}

	c = makeCore(t, 1)
	ch = addYunjin(t, c, 0, 20)
	hits = watch(c)
	_, err = ch.Character.(*char).Skill(map[string]int{"perfect": 1, "hold": 2})
	if err != nil {
		t.Fatal(err)
	}
	advance(c, skillHitmarks[2])
	if len(*hits) != 1 || (*hits)[0].frame != 92 || (*hits)[0].info.Abil != "Opening Flourish (Level 2)" {
		t.Fatalf("a0 hold still works %+v", *hits)
	}

	c = makeCore(t, 1)
	ch = addYunjin(t, c, 0, 40)
	if ch.Base.Ascension != 1 {
		t.Fatalf("ascension %d", ch.Base.Ascension)
	}
	hits = watch(c)
	parts := particlesOf(c)
	act, err = ch.Character.(*char).Skill(map[string]int{"perfect": 1})
	if err != nil {
		t.Fatal(err)
	}
	if act.Frames(action.ActionJump) != skillFrames[0][action.ActionJump] || act.AnimationLength != skillFrames[0][action.InvalidAction] {
		t.Fatalf("frames jump %d len %d", act.Frames(action.ActionJump), act.AnimationLength)
	}
	advance(c, skillHitmarks[0]+ch.ParticleDelay)
	if len(*hits) != 1 {
		t.Fatalf("hits %+v", *hits)
	}
	got := (*hits)[0]
	if got.frame != 13 || got.info.Abil != "Opening Flourish (Level 2)" {
		t.Fatalf("perfect %+v", got)
	}
	if math.Abs(got.info.Mult-6.3376) > 1e-9 || got.info.Durability != 100 || got.info.PoiseDMG != 300 || got.info.Element != attributes.Geo {
		t.Fatalf("perfect attack %+v", got.info)
	}
	if len(*parts) != 1 || (*parts)[0] != 3 {
		t.Fatalf("particles %v", *parts)
	}

	c = makeCore(t, 1)
	ch = addYunjin(t, c, 1, 40)
	if _, err = ch.Character.(*char).Skill(map[string]int{"perfect": 1}); err != nil {
		t.Fatal(err)
	}
	advance(c, skillCDStarts[0])
	// C1 is an 18% shorter 9s cooldown, rounded up from 442.8 frames.
	if ch.Cooldown(action.ActionSkill) != 443 {
		t.Fatalf("c1 perfect cd %d", ch.Cooldown(action.ActionSkill))
	}
}

func TestHoldLevel1ParticlesFollowThreeToTwo(t *testing.T) {
	var twos, threes int
	for seed := int64(1); seed <= 1000; seed++ {
		c := makeCore(t, seed)
		ch := addYunjin(t, c, 0, 90)
		var n float64
		c.Events.Subscribe(event.OnParticleReceived, func(args ...any) {
			n = args[0].(character.Particle).Num
		}, "particles")
		if _, err := ch.Character.(*char).Skill(map[string]int{"hold": 1}); err != nil {
			t.Fatal(err)
		}
		advance(c, skillHitmarks[1]+ch.ParticleDelay)
		switch n {
		case 2:
			twos++
		case 3:
			threes++
		default:
			t.Fatalf("seed %d particles %v", seed, n)
		}
	}
	if twos < 560 || twos > 640 || twos+threes != 1000 {
		t.Fatalf("2-particle count %d, 3-particle count %d", twos, threes)
	}
}
