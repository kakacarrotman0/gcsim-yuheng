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
	"github.com/genshinsim/gcsim/pkg/core/player"
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

func addProfileAt(t *testing.T, c *core.Core, key keys.Char, cons, talent, level, maxLevel int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(key, testhelper.TestWeaponKey)
	prof.Base.Cons = cons
	prof.Base.Level = level
	prof.Base.MaxLevel = maxLevel
	prof.Talents = info.TalentProfile{Attack: talent, Skill: talent, Burst: talent}
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("add %v: %v", key, err)
	}
	return c.Player.ByIndex(idx)
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

func TestPhysicalOutsideMasterstroke(t *testing.T) {
	c, trg := makeCore(1)
	ch := addProfile(t, c, keys.Lohen, 0, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	ch.ParticleDelay = 0
	c.Player.SetActive(ch.Index())
	var particles int
	c.Events.Subscribe(event.OnParticleReceived, func(args ...any) {
		particles++
	}, "particles")
	var hits []info.AttackInfo
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.ActorIndex == ch.Index() {
			hits = append(hits, ai)
		}
	}, "hits")

	if err := c.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, attackHitmarks[0][0]+2)
	if len(hits) != 1 || hits[0].Element != attributes.Physical || hits[0].IgnoreInfusion || math.Abs(hits[0].Mult-attack1[ch.TalentLvlAttack()]) > 1e-9 {
		t.Fatalf("normal outside %+v", hits)
	}
	if particles != 0 || trg[0].AuraContains(attributes.Cryo) {
		t.Fatalf("physical normal particles %d aura cryo %v", particles, trg[0].AuraContains(attributes.Cryo))
	}

	hits = nil
	if err := c.Player.Exec(action.ActionCharge, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, chargeHitmarks[len(chargeHitmarks)-1]+2)
	if len(hits) != 2 {
		t.Fatalf("charged hits %d", len(hits))
	}
	for i, ai := range hits {
		if ai.Element != attributes.Physical || ai.IgnoreInfusion || math.Abs(ai.Mult-attack7[ch.TalentLvlAttack()]) > 1e-9 {
			t.Fatalf("charged %d %+v", i, ai)
		}
	}
	if particles != 0 {
		t.Fatalf("physical charged particles %d", particles)
	}

	c.Player.AddWeaponInfuse(ch.Index(), "test-pyro", attributes.Pyro, 300, true, attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge)
	asLohen(ch).ResetNormalCounter()
	hits = nil
	if err := c.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, attackHitmarks[0][0]+2)
	if len(hits) != 1 || hits[0].Element != attributes.Pyro || hits[0].IgnoreInfusion {
		t.Fatalf("infused normal outside %+v", hits)
	}
}

func TestMasterstrokeCryoConversion(t *testing.T) {
	c, trg := makeCore(2)
	ch := addProfile(t, c, keys.Lohen, 0, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	ch.ParticleDelay = 0
	c.Player.SetActive(ch.Index())
	c.Player.AddWeaponInfuse(ch.Index(), "test-pyro", attributes.Pyro, 600, true, attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge)
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	trg[0].AttachOrRefill(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Pyro, Durability: 25},
	})
	advance(c, 1)

	lc := asLohen(ch)
	var melts int
	c.Events.Subscribe(event.OnMelt, func(args ...any) {
		if args[0].(info.Target).Key() == trg[0].Key() {
			melts++
		}
	}, "melt")
	var particles []character.Particle
	c.Events.Subscribe(event.OnParticleReceived, func(args ...any) {
		particles = append(particles, args[0].(character.Particle))
	}, "particles")
	var hits []info.AttackInfo
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.ActorIndex == ch.Index() && ai.AttackTag != attacks.AttackTagElementalArt {
			hits = append(hits, ai)
		}
	}, "hits")

	lc.ResetNormalCounter()
	if err := c.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, attackHitmarks[0][0]+2)
	if len(hits) != 2 {
		t.Fatalf("normal hits %d", len(hits))
	}
	for i, ai := range hits {
		if ai.Element != attributes.Cryo || !ai.IgnoreInfusion || math.Abs(ai.Mult-enhanced1[lc.skillLvl()]) > 1e-9 {
			t.Fatalf("masterstroke normal %d %+v want mult %v", i, ai, enhanced1[lc.skillLvl()])
		}
	}
	if melts != 1 || !trg[1].AuraContains(attributes.Cryo) {
		t.Fatalf("melt %d clean aura cryo %v", melts, trg[1].AuraContains(attributes.Cryo))
	}
	if len(particles) != 1 || particles[0].Ele != attributes.Cryo || particles[0].Num != 1 {
		t.Fatalf("particles %+v", particles)
	}

	hits = nil
	if err := c.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, attackHitmarks[1][0]+2)
	if len(hits) != 2 || len(particles) != 1 {
		t.Fatalf("second normal hits %d particles %d", len(hits), len(particles))
	}

	hits = nil
	if err := c.Player.Exec(action.ActionCharge, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, chargeHitmarks[len(chargeHitmarks)-1]+2)
	if len(hits) != 4 {
		t.Fatalf("charged hits %d", len(hits))
	}
	for i, ai := range hits {
		if ai.Element != attributes.Cryo || !ai.IgnoreInfusion || math.Abs(ai.Mult-enhanced7[lc.skillLvl()]) > 1e-9 {
			t.Fatalf("masterstroke charged %d %+v", i, ai)
		}
	}

	hits = nil
	if err := c.Player.SetAirborne(player.AirborneXianyun); err != nil {
		t.Fatal(err)
	}
	if err := c.Player.Exec(action.ActionLowPlunge, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, lowPlungeHitmark+2)
	if len(hits) != 2 {
		t.Fatalf("plunge hits %d", len(hits))
	}
	for i, ai := range hits {
		if ai.Element != attributes.Cryo || !ai.IgnoreInfusion || ai.AttackTag != attacks.AttackTagPlunge || math.Abs(ai.Mult-lowPlunge[lc.skillLvl()]) > 1e-9 {
			t.Fatalf("masterstroke plunge %d %+v", i, ai)
		}
	}
}

func TestMasterstrokeEndsOnSwap(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Lohen, 0, 10)
	addProfile(t, c, keys.Bennett, 0, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	lc := asLohen(ch)
	lc.joy = joyCap
	lc.will = 40
	ready, fail := ch.ActionReady(action.ActionSkill, map[string]int{"etched": 1})
	if !ready || fail != action.NoFailure {
		t.Fatalf("etched before swap ready %v fail %v", ready, fail)
	}
	started := c.F
	if err := c.Player.Exec(action.ActionSwap, keys.Bennett, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 2)
	if c.F-started >= int(masterDuration[0]*60) {
		t.Fatalf("swap test ran past the masterstroke timer: %d", c.F-started)
	}
	master, _ := ch.Condition([]string{"masterstroke"})
	joy, _ := ch.Condition([]string{"joy"})
	will, _ := ch.Condition([]string{"will"})
	if master != 0 || joy.(float64) != 0 || will.(float64) != 0 || lc.c6Mark {
		t.Fatalf("after swap master %v joy %v will %v mark %v", master, joy, will, lc.c6Mark)
	}
	ready, _ = ch.ActionReady(action.ActionSkill, map[string]int{"etched": 1})
	if ready {
		t.Fatal("etched stayed available after leaving the field")
	}

	if err := c.Player.Exec(action.ActionSwap, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 2)
	master, _ = ch.Condition([]string{"masterstroke"})
	joy, _ = ch.Condition([]string{"joy"})
	will, _ = ch.Condition([]string{"will"})
	ready, _ = ch.ActionReady(action.ActionSkill, map[string]int{"etched": 1})
	if master != 0 || joy.(float64) != 0 || will.(float64) != 0 || ready {
		t.Fatalf("swap back resumed master %v joy %v will %v etched %v", master, joy, will, ready)
	}

	var ele attributes.Element
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.ActorIndex == ch.Index() && ai.AttackTag == attacks.AttackTagNormal {
			ele = ai.Element
		}
	}, "hits")
	lc.ResetNormalCounter()
	if err := c.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, attackHitmarks[0][0]+2)
	if ele != attributes.Physical {
		t.Fatalf("normal after swap-back %v", ele)
	}
}

func TestWillAndHighSpiritsAtAscensionZero(t *testing.T) {
	c, trg := makeCore(1)
	ch := addProfileAt(t, c, keys.Lohen, 0, 1, 1, 20)
	ben := addProfile(t, c, keys.Bennett, 0, 1)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	if ch.Base.Ascension != 0 {
		t.Fatalf("ascension %d", ch.Base.Ascension)
	}
	c.Player.SetActive(ch.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	if !ch.StatusIsActive(spiritsKey) {
		t.Fatal("high spirits did not apply at ascension 0")
	}
	lc := asLohen(ch)
	base := ch.Stat(attributes.BaseATK)
	emit := func(dmg float64) {
		c.Events.Emit(event.OnEnemyDamage, trg[0], &info.AttackEvent{
			Info: info.AttackInfo{ActorIndex: ben.Index()},
		}, dmg, false)
	}
	emit(1)
	will, _ := ch.Condition([]string{"will"})
	if math.Abs(will.(float64)-willFailGain) > 1e-9 {
		t.Fatalf("fail will %v", will)
	}
	emit(base * willAtkRatio[lc.skillLvl()])
	will, _ = ch.Condition([]string{"will"})
	if math.Abs(will.(float64)-(willFailGain+willSuccess[lc.skillLvl()])) > 1e-6 {
		t.Fatalf("base will %v", will)
	}
	emit(base * a1ExtraRatio)
	will, _ = ch.Condition([]string{"will"})
	want := willFailGain + willSuccess[lc.skillLvl()]*2
	if math.Abs(will.(float64)-want) > 1e-6 {
		t.Fatalf("ascension 0 extra will %v want %v", will, want)
	}
}
