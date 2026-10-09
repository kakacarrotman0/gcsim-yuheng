package lohen

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/bennett"
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

func makeCore(trgCount int) (*core.Core, []*enemy.Enemy) {
	c, err := core.New(core.Opt{Seed: 1, Debug: false})
	if err != nil {
		panic(err)
	}
	a := avatar.New(c, info.Point{X: 0, Y: 0}, 1)
	c.Combat.SetPlayer(a)
	var trgs []*enemy.Enemy
	for range trgCount {
		e := enemy.New(c, info.EnemyProfile{
			Level:  100,
			Resist: make(map[attributes.Element]float64),
			Pos:    info.Coord{X: 0, Y: 0, R: 1},
		})
		trgs = append(trgs, e)
		c.Combat.AddEnemy(e)
	}
	c.Player.SetActive(0)
	if len(trgs) > 0 {
		c.Combat.DefaultTarget = trgs[0].Key()
	}
	return c, trgs
}

func advance(c *core.Core, n int) {
	for range n {
		c.F++
		c.Tick()
	}
}

func addProfile(t *testing.T, c *core.Core, key keys.Char, cons, talent int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(key, testhelper.TestWeaponKey)
	prof.Base.Cons = cons
	prof.Talents = info.TalentProfile{Attack: talent, Skill: talent, Burst: talent}
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("add %v: %v", key, err)
	}
	return c.Player.ByIndex(idx)
}

func asLohen(ch *character.CharWrapper) *char {
	return ch.Character.(*char)
}

func TestIdentity(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Lohen, 0, 1)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	if ch.Base.Element != attributes.Cryo || ch.Weapon.Class != info.WeaponClassSpear || ch.EnergyMax != 60 {
		t.Fatalf("identity ele %v weapon %v energy %v", ch.Base.Element, ch.Weapon.Class, ch.EnergyMax)
	}
	if ch.Moonsign != 0 || ch.IsHexerei {
		t.Fatalf("moonsign %v hexerei %v", ch.Moonsign, ch.IsHexerei)
	}
	if math.Abs(ch.BaseStats[attributes.CD]-(0.5+0.384)) > 1e-4 {
		t.Fatalf("cd %v", ch.BaseStats[attributes.CD])
	}
}

func TestJoyAndWill(t *testing.T) {
	c, trg := makeCore(1)
	ch := addProfile(t, c, keys.Lohen, 0, 10)
	ben := addProfile(t, c, keys.Bennett, 0, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	if err := c.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, attackHitmarks[0][0]+2)
	joy, _ := ch.Condition([]string{"joy"})
	if math.Abs(joy.(float64)-joyOnNormal[ch.TalentLvlSkill()]) > 1e-9 {
		t.Fatalf("joy %v", joy)
	}

	base := ch.Stat(attributes.BaseATK)
	c.Events.Emit(event.OnEnemyDamage, trg[0], &info.AttackEvent{
		Info: info.AttackInfo{ActorIndex: ben.Index()},
	}, base*willAtkRatio[asLohen(ch).skillLvl()], false)
	will, _ := ch.Condition([]string{"will"})
	if math.Abs(will.(float64)-willSuccess[asLohen(ch).skillLvl()]) > 1e-9 {
		t.Fatalf("will %v", will)
	}
	c.Events.Emit(event.OnEnemyDamage, trg[0], &info.AttackEvent{
		Info: info.AttackInfo{ActorIndex: ben.Index()},
	}, base*a1ExtraRatio, false)
	will, _ = ch.Condition([]string{"will"})
	want := willSuccess[asLohen(ch).skillLvl()] + willSuccess[asLohen(ch).skillLvl()] + a1ExtraWill
	if math.Abs(will.(float64)-want) > 1e-6 {
		t.Fatalf("will after a1 %v want %v", will, want)
	}
}

func TestEtchedConsumesWill(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Lohen, 0, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	lc := asLohen(ch)
	lc.joy = joyCap
	lc.will = 50
	var mults []float64
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.AttackTag == attacks.AttackTagElementalArt {
			mults = append(mults, ai.Mult)
		}
	}, "hits")
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, map[string]int{"etched": 1}); err != nil {
		t.Fatal(err)
	}
	will, _ := ch.Condition([]string{"will"})
	if will.(float64) != 0 {
		t.Fatalf("will after etched %v", will)
	}
	advance(c, 40)
	if len(mults) != 4 {
		t.Fatalf("etched hits %d", len(mults))
	}
	want := raid[lc.skillLvl()] * (1 + 50*willRatio[lc.skillLvl()])
	for i, got := range mults {
		if math.Abs(got-want) > 1e-6 {
			t.Fatalf("hit %d mult %v want %v", i, got, want)
		}
	}
}

func TestC6KeepsWill(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Lohen, 6, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	lc := asLohen(ch)
	lc.joy = joyCap
	lc.will = 80
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, map[string]int{"etched": 1}); err != nil {
		t.Fatal(err)
	}
	will, _ := ch.Condition([]string{"will"})
	if will.(float64) != 80 {
		t.Fatalf("c6 will %v", will)
	}
	advance(c, 40)
	joy, _ := ch.Condition([]string{"joy"})
	if joy.(float64) != joyCap {
		t.Fatalf("c6 joy %v", joy)
	}
	if ch.TalentLvlSkill() != 12 {
		t.Fatalf("c3 skill index %d", ch.TalentLvlSkill())
	}
}

func TestA4CryoReaction(t *testing.T) {
	c, trg := makeCore(1)
	ch := addProfile(t, c, keys.Lohen, 0, 10)
	ben := addProfile(t, c, keys.Bennett, 0, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	beforeL := ch.Stat(attributes.ATKP)
	beforeB := ben.Stat(attributes.ATKP)
	c.Events.Emit(event.OnMelt, trg[0], &info.AttackEvent{
		Info: info.AttackInfo{ActorIndex: ben.Index()},
	})
	if math.Abs(ch.Stat(attributes.ATKP)-(beforeL+a4ATK)) > 1e-9 {
		t.Fatalf("lohen atk %v", ch.Stat(attributes.ATKP))
	}
	if math.Abs(ben.Stat(attributes.ATKP)-(beforeB+a4ATK)) > 1e-9 {
		t.Fatalf("bennett atk %v", ben.Stat(attributes.ATKP))
	}
}

func TestBurstHitCount(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Lohen, 0, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	ch.Energy = ch.EnergyMax
	var n int
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.AttackTag == attacks.AttackTagElementalBurst && ai.Element == attributes.Cryo {
			n++
		}
	}, "hits")
	if err := c.Player.Exec(action.ActionBurst, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, burstHitmarks[len(burstHitmarks)-1]+2)
	if n != 6 {
		t.Fatalf("burst hits %d", n)
	}
}
