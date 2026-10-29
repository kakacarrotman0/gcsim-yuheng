package candace

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

func addCandace(t *testing.T, c *core.Core, cons, maxLevel int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Candace, testhelper.TestWeaponKey)
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

func TestTapAndHold(t *testing.T) {
	c := makeCore(t)
	ch := addCandace(t, c, 0, 90)
	hits := watch(c)
	parts := particlesOf(c)
	act, err := ch.Character.(*char).Skill(nil)
	if err != nil {
		t.Fatal(err)
	}
	if act.AnimationLength != skillFrames[0][action.InvalidAction] || act.CanQueueAfter != skillFrames[0][action.ActionSwap] {
		t.Fatalf("tap len %d queue %d", act.AnimationLength, act.CanQueueAfter)
	}
	advance(c, skillHitmarks[0]+ch.ParticleDelay)
	if len(*hits) != 1 {
		t.Fatalf("hits %+v", *hits)
	}
	got := (*hits)[0]
	wantFlat := skillDmg[0][ch.TalentLvlSkill()] * ch.MaxHP()
	if got.frame != 16 || got.info.Abil != "Sacred Rite: Heron's Sanctum" || got.info.Element != attributes.Hydro {
		t.Fatalf("tap %+v", got)
	}
	if math.Abs(got.info.FlatDmg-wantFlat) > 1 || math.Abs(wantFlat-0.204*ch.MaxHP()) > 1 {
		t.Fatalf("flat %v want %v", got.info.FlatDmg, wantFlat)
	}
	if got.info.Durability != 25 || got.info.PoiseDMG != 150 || len(*parts) != 1 || (*parts)[0] != 2 {
		t.Fatalf("gauge %+v particles %v", got.info, *parts)
	}
	if ch.Cooldown(action.ActionSkill) != skillCD[0]-(c.F-skillCDStarts[0]) {
		t.Fatalf("tap cd %d at frame %d", ch.Cooldown(action.ActionSkill), c.F)
	}

	c = makeCore(t)
	ch = addCandace(t, c, 0, 90)
	hits = watch(c)
	parts = particlesOf(c)
	act, err = ch.Character.(*char).Skill(map[string]int{"hold": 1})
	if err != nil {
		t.Fatal(err)
	}
	if act.Frames(action.ActionSwap) != 110 || act.AnimationLength != 113 || act.CanQueueAfter != 110 {
		t.Fatalf("hold frames %d len %d queue %d", act.Frames(action.ActionSwap), act.AnimationLength, act.CanQueueAfter)
	}
	advance(c, 88)
	if ch.Cooldown(action.ActionSkill) != 0 {
		t.Fatalf("cd started early: %d", ch.Cooldown(action.ActionSkill))
	}
	advance(c, 1)
	if ch.Cooldown(action.ActionSkill) != skillCD[1] {
		t.Fatalf("hold cd %d", ch.Cooldown(action.ActionSkill))
	}
	advance(c, skillHitmarks[1]+ch.ParticleDelay-c.F)
	if len(*hits) != 1 || (*hits)[0].frame != 91 || (*hits)[0].info.Abil != "Sacred Rite: Heron's Sanctum (Charged)" {
		t.Fatalf("hold %+v", *hits)
	}
	got = (*hits)[0]
	wantFlat = skillDmg[1][ch.TalentLvlSkill()] * ch.MaxHP()
	if math.Abs(got.info.FlatDmg-wantFlat) > 1 || math.Abs(skillDmg[1][8]-0.32368) > 1e-9 {
		t.Fatalf("charged flat %v ratio %v", got.info.FlatDmg, skillDmg[1][8])
	}
	if got.info.Durability != 25 || got.info.PoiseDMG != 300 || len(*parts) != 1 || (*parts)[0] != 3 {
		t.Fatalf("charged %+v particles %v", got.info, *parts)
	}
}

func TestPerfectCounterRequiresAscensionAndSkips55Frames(t *testing.T) {
	c := makeCore(t)
	ch := addCandace(t, c, 0, 20)
	if ch.Base.Ascension != 0 {
		t.Fatalf("ascension %d", ch.Base.Ascension)
	}
	hits := watch(c)
	act, err := ch.Character.(*char).Skill(map[string]int{"perfect": 1})
	if err != nil {
		t.Fatal(err)
	}
	if act.AnimationLength != skillFrames[0][action.InvalidAction] {
		t.Fatalf("a0 perfect len %d", act.AnimationLength)
	}
	advance(c, 16)
	if len(*hits) != 1 || (*hits)[0].info.Abil != "Sacred Rite: Heron's Sanctum" {
		t.Fatalf("a0 perfect %+v", *hits)
	}

	c = makeCore(t)
	ch = addCandace(t, c, 0, 20)
	hits = watch(c)
	act, err = ch.Character.(*char).Skill(map[string]int{"perfect": 1, "hold": 1})
	if err != nil {
		t.Fatal(err)
	}
	if act.Frames(action.ActionSwap) != 110 || act.CanQueueAfter != 110 {
		t.Fatalf("a0 hold shortcut %d %d", act.Frames(action.ActionSwap), act.CanQueueAfter)
	}
	advance(c, 91)
	if len(*hits) != 1 || (*hits)[0].frame != 91 {
		t.Fatalf("a0 hold %+v", *hits)
	}

	c = makeCore(t)
	ch = addCandace(t, c, 0, 40)
	if ch.Base.Ascension != 1 {
		t.Fatalf("ascension %d", ch.Base.Ascension)
	}
	hits = watch(c)
	parts := particlesOf(c)
	act, err = ch.Character.(*char).Skill(map[string]int{"perfect": 1})
	if err != nil {
		t.Fatal(err)
	}
	if act.Frames(action.ActionSwap) != 110-55 || act.AnimationLength != 113-55 || act.CanQueueAfter != 110-55 {
		t.Fatalf("a1 frames %d len %d queue %d", act.Frames(action.ActionSwap), act.AnimationLength, act.CanQueueAfter)
	}
	advance(c, 33)
	if ch.Cooldown(action.ActionSkill) != 0 {
		t.Fatalf("cd early %d", ch.Cooldown(action.ActionSkill))
	}
	advance(c, 1)
	if ch.Cooldown(action.ActionSkill) != skillCD[1] {
		t.Fatalf("a1 cd %d", ch.Cooldown(action.ActionSkill))
	}
	advance(c, 36+ch.ParticleDelay-c.F)
	if len(*hits) != 1 {
		t.Fatalf("hits %+v", *hits)
	}
	got := (*hits)[0]
	if got.frame != 36 || got.info.Abil != "Sacred Rite: Heron's Sanctum (Charged)" {
		t.Fatalf("a1 %+v", got)
	}
	if got.info.Durability != 25 || got.info.PoiseDMG != 300 || got.info.Element != attributes.Hydro {
		t.Fatalf("a1 attack %+v", got.info)
	}
	if math.Abs(got.info.FlatDmg-skillDmg[1][ch.TalentLvlSkill()]*ch.MaxHP()) > 1 {
		t.Fatalf("flat %v", got.info.FlatDmg)
	}
	if len(*parts) != 1 || (*parts)[0] != 3 {
		t.Fatalf("particles %v", *parts)
	}
}

func TestC4ShortensHoldCooldownIncludingTheCounter(t *testing.T) {
	c := makeCore(t)
	ch := addCandace(t, c, 4, 90)
	if _, err := ch.Character.(*char).Skill(map[string]int{"hold": 1}); err != nil {
		t.Fatal(err)
	}
	advance(c, skillCDStarts[1])
	if ch.Cooldown(action.ActionSkill) != skillCD[0] {
		t.Fatalf("c4 hold cd %d", ch.Cooldown(action.ActionSkill))
	}

	c = makeCore(t)
	ch = addCandace(t, c, 4, 90)
	if _, err := ch.Character.(*char).Skill(map[string]int{"perfect": 1}); err != nil {
		t.Fatal(err)
	}
	advance(c, skillCDStarts[1]-55)
	if ch.Cooldown(action.ActionSkill) != skillCD[0] {
		t.Fatalf("c4 perfect cd %d", ch.Cooldown(action.ActionSkill))
	}
}

func TestC2TriggersOnChargedAndPerfectHits(t *testing.T) {
	for _, param := range []map[string]int{{"hold": 1}, {"perfect": 1}} {
		c := makeCore(t)
		ch := addCandace(t, c, 2, 90)
		if _, err := ch.Character.(*char).Skill(param); err != nil {
			t.Fatal(err)
		}
		hit := skillHitmarks[1]
		if param["perfect"] == 1 {
			hit -= 55
		}
		before := ch.Stat(attributes.HPP)
		advance(c, hit-1)
		if math.Abs(ch.Stat(attributes.HPP)-before) > 1e-9 {
			t.Fatalf("c2 early %v from %v", ch.Stat(attributes.HPP), before)
		}
		advance(c, 1)
		if math.Abs(ch.Stat(attributes.HPP)-before-0.2) > 1e-9 {
			t.Fatalf("c2 %v from %v param %v", ch.Stat(attributes.HPP), before, param)
		}
	}
}
