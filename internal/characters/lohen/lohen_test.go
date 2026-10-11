package lohen

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/bennett"
	_ "github.com/genshinsim/gcsim/internal/characters/fischl"
	_ "github.com/genshinsim/gcsim/internal/characters/kaeya"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
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
	return addProfileWithParams(t, c, key, cons, talent, nil)
}

func addProfileWithParams(t *testing.T, c *core.Core, key keys.Char, cons, talent int, params map[string]int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(key, testhelper.TestWeaponKey)
	prof.Base.Cons = cons
	prof.Talents = info.TalentProfile{Attack: talent, Skill: talent, Burst: talent}
	if params != nil {
		prof.Params = params
	}
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
	// Same frame, same teammate. The success gate is already up. The A1 extra is a separate gate.
	c.Events.Emit(event.OnEnemyDamage, trg[0], &info.AttackEvent{
		Info: info.AttackInfo{ActorIndex: ben.Index()},
	}, base*a1ExtraRatio, false)
	will, _ = ch.Condition([]string{"will"})
	want := willSuccess[asLohen(ch).skillLvl()] + a1ExtraWill
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
	var etchedCD float64
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ae := args[1].(*info.AttackEvent)
		if etchedCD == 0 && len(ae.Info.Abil) >= 6 && ae.Info.Abil[:6] == "Etched" {
			etchedCD = ae.Snapshot.Stats[attributes.CD]
		}
	}, "c6cd")
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
	if math.Abs(etchedCD-(ch.Stat(attributes.CD)+c6Crit)) > 1e-6 {
		t.Fatalf("c6 crit dmg %v stat %v", etchedCD, ch.Stat(attributes.CD))
	}
	if !lc.c6Mark {
		t.Fatal("c6 mark was not set on the etched final hit")
	}
	if lc.etchedMax() != etchedBase+2 {
		t.Fatalf("c6 etched uses %d", lc.etchedMax())
	}
	until := lc.masterUntil
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, map[string]int{"etched": 1}); err != nil {
		t.Fatal(err)
	}
	if lc.masterUntil-until != c6Extend {
		t.Fatalf("c6 extension %d", lc.masterUntil-until)
	}
	will, _ = ch.Condition([]string{"will"})
	if will.(float64) != 0 {
		t.Fatalf("c6 will inside the 7s flag %v", will)
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
	// 30× is also a success. That gate is still up, and ascension 0 has no extra.
	emit(base * a1ExtraRatio)
	will, _ = ch.Condition([]string{"will"})
	want := willFailGain + willSuccess[lc.skillLvl()]
	if math.Abs(will.(float64)-want) > 1e-6 {
		t.Fatalf("ascension 0 extra will %v want %v", will, want)
	}
}

func waitReady(t *testing.T, c *core.Core, act action.Action, key keys.Char, limit int) bool {
	t.Helper()
	for c.F < limit {
		err := c.Player.ReadyCheck(act, key, nil)
		if err == nil {
			return true
		}
		if err != player.ErrPlayerNotReady && err != player.ErrActionNotReady {
			t.Fatalf("ready %v: %v", act, err)
		}
		advance(c, 1)
	}
	return false
}

func TestLohenSkillICD(t *testing.T) {
	c, trg := makeCore(1)
	ch := addProfile(t, c, keys.Lohen, 0, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	var applied []bool
	// OnEnemyHit is emitted before ICD. OnEnemyDamage carries the post-ICD durability.
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.Abil != "icd-probe" {
			return
		}
		applied = append(applied, ai.Durability > 0)
	}, "icd")
	probe := func() {
		c.QueueAttack(info.AttackInfo{
			ActorIndex: ch.Index(),
			Abil:       "icd-probe",
			AttackTag:  attacks.AttackTagNormal,
			ICDTag:     attacks.ICDTagNormalAttack,
			ICDGroup:   attacks.ICDGroupLohenSkill,
			Element:    attributes.Cryo,
			Durability: 25,
		}, combat.NewSingleTargetHit(trg[0].Key()), 0, 0)
	}
	for range 14 {
		probe()
		advance(c, 1)
	}
	if len(applied) != 14 {
		t.Fatalf("probes %d", len(applied))
	}
	for i, got := range applied {
		want := i < 12 && i%2 == 0
		if got != want {
			t.Fatalf("probe %d applied %v want %v", i, got, want)
		}
	}
	// The 5s group arms its reset on the first hit. gcsim schedules that reset
	// at timer-1, the same convention as every other ICD group.
	first := 0
	resetAt := first + attacks.ICDGroupResetTimer[attacks.ICDGroupLohenSkill] - 1
	for c.F < resetAt {
		probe()
		advance(c, 1)
		if applied[len(applied)-1] {
			t.Fatalf("application returned at frame %d, reset frame %d", c.F-1, resetAt)
		}
	}
	probe()
	advance(c, 1)
	if !applied[len(applied)-1] {
		t.Fatalf("no application after reset, frame %d applied %v", c.F, applied[len(applied)-1])
	}
}

func TestNormalAndChargedShareSkillICD(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Lohen, 0, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	var durs []float64
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.ActorIndex != ch.Index() {
			return
		}
		if ai.AttackTag != attacks.AttackTagNormal && ai.AttackTag != attacks.AttackTagExtra {
			return
		}
		durs = append(durs, float64(ai.Durability))
	}, "hits")
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	limit := int(masterDuration[0] * 60)
	if !waitReady(t, c, action.ActionAttack, keys.Lohen, limit) {
		t.Fatal("attack never ready")
	}
	if err := c.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	if !waitReady(t, c, action.ActionCharge, keys.Lohen, limit) {
		t.Fatal("charge never ready")
	}
	if err := c.Player.Exec(action.ActionCharge, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, chargeHitmarks[len(chargeHitmarks)-1]+2)
	if len(durs) != 3 {
		t.Fatalf("hits %d", len(durs))
	}
	// N1 applies, the first charged hit does not, the second charged hit does.
	want := []float64{25, 0, 25}
	for i, got := range durs {
		if got != want[i] {
			t.Fatalf("hit %d durability %v want %v", i, got, want[i])
		}
	}
}

func TestPlaceholderN1CCount(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Lohen, 0, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	limit := int(masterDuration[0] * 60)
	n1c := 0
	for c.F < limit {
		if !waitReady(t, c, action.ActionAttack, keys.Lohen, limit) || !asLohen(ch).masterActive() {
			break
		}
		if err := c.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
			t.Fatal(err)
		}
		if !waitReady(t, c, action.ActionCharge, keys.Lohen, limit) || !asLohen(ch).masterActive() {
			break
		}
		if err := c.Player.Exec(action.ActionCharge, keys.Lohen, nil); err != nil {
			t.Fatal(err)
		}
		n1c++
	}
	// Placeholder N1 cancel into CA is the 22-frame N1 recovery, and CA cancel
	// into N1 is 36. That loop does not reach KQM's 15–16 average or 17 ideal
	// counts. Those counts are not a frame sheet, so the windows stay as they are.
	if n1c != 13 {
		t.Fatalf("placeholder N1C count %d at frame %d", n1c, c.F)
	}
}

// BowTae's 60 FPS recording, measured from the controller overlay, repeats an
// attack press every 46–47 frames (median 46, first transition 51). That is
// wall-clock time between inputs. It includes hitlag and input delay, and it
// does not separate the N1 cancel from the CA cancel. Parser sims enable
// hitlag; Lohen sets no halt, so the action clock is not extended and the
// engine spacing stays the placeholder sum.
func TestN1CSpacingStaysPlaceholderWithHitlag(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1, EnableHitlag: true, DefHalt: true})
	if err != nil {
		t.Fatal(err)
	}
	a := avatar.New(c, info.Point{X: 0, Y: 0}, 1)
	c.Combat.SetPlayer(a)
	e := enemy.New(c, info.EnemyProfile{
		Level:  100,
		Resist: make(map[attributes.Element]float64),
		Pos:    info.Coord{X: 0, Y: 0, R: 1},
	})
	c.Combat.AddEnemy(e)
	c.Combat.DefaultTarget = e.Key()
	c.Player.SetActive(0)
	ch := addProfile(t, c, keys.Lohen, 0, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	limit := int(masterDuration[0] * 60)
	var starts []int
	for c.F < limit {
		if !waitReady(t, c, action.ActionAttack, keys.Lohen, limit) || !asLohen(ch).masterActive() {
			break
		}
		starts = append(starts, c.F)
		if err := c.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
			t.Fatal(err)
		}
		if !waitReady(t, c, action.ActionCharge, keys.Lohen, limit) || !asLohen(ch).masterActive() {
			break
		}
		if err := c.Player.Exec(action.ActionCharge, keys.Lohen, nil); err != nil {
			t.Fatal(err)
		}
	}
	if len(starts) < 2 {
		t.Fatalf("attack starts %d", len(starts))
	}
	for i := 1; i < len(starts); i++ {
		delta := starts[i] - starts[i-1]
		if delta != 58 {
			t.Fatalf("N1C spacing %d at loop %d, want 58 (placeholder 22+36, not the 46–47 frame input cadence)", delta, i)
		}
	}
}

type n1cHit struct {
	frame int
	abil  string
	ele   attributes.Element
}

type n1cRun struct {
	attackStarts []int
	chargeStarts []int
	hits         []n1cHit
}

// runEarliestN1C queues N1 then CA as soon as the engine allows, with the
// hitlag clock and defense halt on. params selects the timing profile.
func runEarliestN1C(t *testing.T, params map[string]int) n1cRun {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1, EnableHitlag: true, DefHalt: true})
	if err != nil {
		t.Fatal(err)
	}
	a := avatar.New(c, info.Point{X: 0, Y: 0}, 1)
	c.Combat.SetPlayer(a)
	e := enemy.New(c, info.EnemyProfile{
		Level:  100,
		Resist: make(map[attributes.Element]float64),
		Pos:    info.Coord{X: 0, Y: 0, R: 1},
	})
	c.Combat.AddEnemy(e)
	c.Combat.DefaultTarget = e.Key()
	c.Player.SetActive(0)
	ch := addProfileWithParams(t, c, keys.Lohen, 0, 10, params)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	var run n1cRun
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.ActorIndex != ch.Index() {
			return
		}
		if ai.AttackTag != attacks.AttackTagNormal && ai.AttackTag != attacks.AttackTagExtra {
			return
		}
		run.hits = append(run.hits, n1cHit{frame: c.F, abil: ai.Abil, ele: ai.Element})
	}, "n1c-hits")
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	limit := int(masterDuration[0] * 60)
	for c.F < limit {
		if !waitReady(t, c, action.ActionAttack, keys.Lohen, limit) || !asLohen(ch).masterActive() {
			break
		}
		run.attackStarts = append(run.attackStarts, c.F)
		if err := c.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
			t.Fatal(err)
		}
		if !waitReady(t, c, action.ActionCharge, keys.Lohen, limit) || !asLohen(ch).masterActive() {
			break
		}
		run.chargeStarts = append(run.chargeStarts, c.F)
		if err := c.Player.Exec(action.ActionCharge, keys.Lohen, nil); err != nil {
			t.Fatal(err)
		}
	}
	return run
}

func TestN1CProfilesAgainstVideoCadence(t *testing.T) {
	if experimentalN1ToCA+experimentalCAToN1 != experimentalN1C {
		t.Fatalf("cadence split %d+%d, want sum %d", experimentalN1ToCA, experimentalCAToN1, experimentalN1C)
	}
	// The split is the placeholder hitmarks plus an even share of whatever
	// is left. 12 and 28 are the old hitmarks, not new measurements.
	if experimentalN1ToCA != attackHitmarks[0][0]+3 || experimentalCAToN1 != chargeHitmarks[1]+3 {
		t.Fatalf("split %d/%d drifted from hitmarks %d and %d", experimentalN1ToCA, experimentalCAToN1, attackHitmarks[0][0], chargeHitmarks[1])
	}

	base := runEarliestN1C(t, nil)
	cadence := runEarliestN1C(t, map[string]int{"n1c_cadence": 1})

	if len(base.attackStarts) < 2 || base.attackStarts[0] != 24 {
		t.Fatalf("baseline first attack %v", base.attackStarts)
	}
	for i := 1; i < len(base.attackStarts); i++ {
		if got := base.attackStarts[i] - base.attackStarts[i-1]; got != 58 {
			t.Fatalf("baseline spacing %d", got)
		}
	}
	if len(base.chargeStarts) != 13 {
		t.Fatalf("baseline completed N1C %d, want 13", len(base.chargeStarts))
	}

	if len(cadence.attackStarts) < 2 || cadence.attackStarts[0] != 24 {
		t.Fatalf("cadence first attack %v", cadence.attackStarts)
	}
	for i := 1; i < len(cadence.attackStarts); i++ {
		if got := cadence.attackStarts[i] - cadence.attackStarts[i-1]; got != experimentalN1C {
			t.Fatalf("cadence spacing %d, want %d", got, experimentalN1C)
		}
	}
	// Skill→N1 is still the unverified 24-frame placeholder. With that
	// window and a 46-frame loop, the engine starts 17 N1C strings. The
	// 17th charged attack's hits fall after the unextended 13s, so 16
	// strings land both charged hits inside Masterstroke.
	if len(cadence.chargeStarts) != 17 {
		t.Fatalf("cadence started N1C %d, want 17", len(cadence.chargeStarts))
	}
	var n1, ca int
	for _, h := range cadence.hits {
		if h.ele != attributes.Cryo || len(h.abil) < 12 || h.abil[:12] != "Masterstroke" {
			continue
		}
		switch h.abil {
		case "Masterstroke Normal 0":
			n1++
		case "Masterstroke Charged Attack":
			ca++
		}
	}
	if n1 != 17 || ca != 32 {
		t.Fatalf("in-mode hits N1 %d CA %d, want 17 and 32", n1, ca)
	}

	// One cadence cycle keeps the placeholder hitmarks: N1 at +12, CA at +18 and +28.
	a0 := cadence.attackStarts[0]
	c0 := cadence.chargeStarts[0]
	var rel []int
	for _, h := range cadence.hits {
		switch h.abil {
		case "Masterstroke Normal 0":
			if h.frame >= a0 && h.frame < c0 {
				rel = append(rel, h.frame-a0)
			}
		case "Masterstroke Charged Attack":
			if h.frame >= c0 && h.frame < cadence.attackStarts[1] {
				rel = append(rel, h.frame-c0)
			}
		}
	}
	want := []int{attackHitmarks[0][0], chargeHitmarks[0], chargeHitmarks[1]}
	if len(rel) != len(want) {
		t.Fatalf("first-cycle hit offsets %v", rel)
	}
	for i := range want {
		if rel[i] != want[i] {
			t.Fatalf("hit %d at +%d, placeholder hitmark %d", i, rel[i], want[i])
		}
	}
}

func enemyCB(t info.Target) info.AttackCB {
	return info.AttackCB{Target: t, AttackEvent: &info.AttackEvent{Info: info.AttackInfo{Element: attributes.Physical}}}
}

func TestJoyAndWillGates(t *testing.T) {
	c, trg := makeCore(1)
	ch := addProfile(t, c, keys.Lohen, 0, 10)
	ben := addProfile(t, c, keys.Bennett, 0, 10)
	kaeya := addProfile(t, c, keys.Kaeya, 0, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	lc := asLohen(ch)
	joyAmt := joyOnNormal[lc.skillLvl()]
	cb := enemyCB(trg[0])
	lc.hitCB(joyAmt, joyNAKey, false)(cb)
	lc.hitCB(joyAmt, joyNAKey, false)(cb)
	joy, _ := ch.Condition([]string{"joy"})
	if math.Abs(joy.(float64)-joyAmt) > 1e-9 {
		t.Fatalf("second normal inside 0.1s joy %v", joy)
	}
	// Charged A and charged B do not share the normal gate, or each other.
	chargeAmt := joyOnCharge[lc.skillLvl()]
	lc.hitCB(chargeAmt, joyCAAKey, false)(cb)
	lc.hitCB(chargeAmt, joyCABKey, false)(cb)
	joy, _ = ch.Condition([]string{"joy"})
	if math.Abs(joy.(float64)-(joyAmt+2*chargeAmt)) > 1e-9 {
		t.Fatalf("charged joy %v", joy)
	}
	advance(c, joyWindow-1)
	lc.hitCB(joyAmt, joyNAKey, false)(cb)
	joy, _ = ch.Condition([]string{"joy"})
	if math.Abs(joy.(float64)-(joyAmt+2*chargeAmt)) > 1e-9 {
		t.Fatalf("joy before window end %v", joy)
	}
	advance(c, 1)
	lc.hitCB(joyAmt, joyNAKey, false)(cb)
	joy, _ = ch.Condition([]string{"joy"})
	if math.Abs(joy.(float64)-(2*joyAmt+2*chargeAmt)) > 1e-9 {
		t.Fatalf("joy after 0.1s %v", joy)
	}

	base := ch.Stat(attributes.BaseATK)
	lvl := lc.skillLvl()
	emit := func(src int, dmg float64) {
		c.Events.Emit(event.OnEnemyDamage, trg[0], &info.AttackEvent{
			Info: info.AttackInfo{ActorIndex: src},
		}, dmg, false)
	}
	emit(ben.Index(), base*willAtkRatio[lvl])
	emit(ben.Index(), base*willAtkRatio[lvl])
	will, _ := ch.Condition([]string{"will"})
	one := willSuccess[lvl]
	if math.Abs(will.(float64)-one) > 1e-6 {
		t.Fatalf("same-character will %v", will)
	}
	// Another teammate is a different unique modifier.
	emit(kaeya.Index(), base*willAtkRatio[lvl])
	will, _ = ch.Condition([]string{"will"})
	if math.Abs(will.(float64)-2*one) > 1e-6 {
		t.Fatalf("second character will %v", will)
	}
	advance(c, willWindowF)
	emit(ben.Index(), 1)
	will, _ = ch.Condition([]string{"will"})
	if math.Abs(will.(float64)-(2*one+willFailGain)) > 1e-6 {
		t.Fatalf("fail after window %v", will)
	}
}

func TestC1C2C4C5(t *testing.T) {
	c, trg := makeCore(1)
	ch := addProfile(t, c, keys.Lohen, 1, 8)
	ben := addProfile(t, c, keys.Bennett, 0, 8)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	lc := asLohen(ch)
	base := ch.Stat(attributes.BaseATK)
	c.Events.Emit(event.OnEnemyDamage, trg[0], &info.AttackEvent{
		Info: info.AttackInfo{ActorIndex: ben.Index()},
	}, base*willAtkRatio[lc.skillLvl()], false)
	will, _ := ch.Condition([]string{"will"})
	if math.Abs(will.(float64)-willSuccess[lc.skillLvl()]*c1Efficiency) > 1e-6 {
		t.Fatalf("c1 will %v", will)
	}
	if lc.willMax() != willCapBase*c1MaxRatio {
		t.Fatalf("c1 cap %v", lc.willMax())
	}

	c2, trg2 := makeCore(1)
	// Evilsbane is a 5.5 box starting 1.3 in front of Lohen. The default target sits on him.
	trg2[0].SetPos(info.Point{X: 0, Y: 3})
	ch2 := addProfile(t, c2, keys.Lohen, 2, 8)
	ben2 := addProfile(t, c2, keys.Bennett, 0, 8)
	if err := c2.Init(); err != nil {
		t.Fatal(err)
	}
	c2.Player.SetActive(ch2.Index())
	if err := c2.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	lc2 := asLohen(ch2)
	lc2.joy = joyCap
	var evils []float64
	c2.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.Abil == "Evilsbane Blade" {
			evils = append(evils, ai.Mult)
		}
	}, "c2")
	if err := c2.Player.Exec(action.ActionSkill, keys.Lohen, map[string]int{"etched": 1}); err != nil {
		t.Fatal(err)
	}
	advance(c2, 40)
	before := ben2.Stat(attributes.EM)
	lohenEM := ch2.Stat(attributes.EM)
	if err := c2.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(c2, attackHitmarks[0][0]+2)
	if len(evils) != 1 || math.Abs(evils[0]-c2Ratio) > 1e-9 {
		t.Fatalf("evilsbane %v", evils)
	}
	if math.Abs(ben2.Stat(attributes.EM)-(before+c2EM)) > 1e-6 {
		t.Fatalf("c2 em %v", ben2.Stat(attributes.EM))
	}
	if math.Abs(ch2.Stat(attributes.EM)-lohenEM) > 1e-6 {
		t.Fatalf("lohen em changed from %v to %v", lohenEM, ch2.Stat(attributes.EM))
	}
	advance(c2, c2Duration)
	if math.Abs(ben2.Stat(attributes.EM)-before) > 1e-6 {
		t.Fatalf("c2 em after 8s %v", ben2.Stat(attributes.EM))
	}

	c4, _ := makeCore(1)
	ch4 := addProfile(t, c4, keys.Lohen, 4, 8)
	if err := c4.Init(); err != nil {
		t.Fatal(err)
	}
	c4.Player.SetActive(ch4.Index())
	ch4.Energy = 0
	if err := c4.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	if ch4.Energy != c4Energy {
		t.Fatalf("c4 immediate energy %v", ch4.Energy)
	}
	if ch4.TalentLvlBurst() != 7 {
		t.Fatalf("c4 burst index %d", ch4.TalentLvlBurst())
	}
	ch4.Energy = ch4.EnergyMax
	asLohen(ch4).will = 1
	asLohen(ch4).enterMasterstroke()
	var burstMult float64
	c4.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.AttackTag == attacks.AttackTagElementalBurst && burstMult == 0 {
			burstMult = ai.Mult
		}
	}, "c4burst")
	if err := c4.Player.Exec(action.ActionBurst, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(c4, burstHitmarks[0]+2)
	lvl := ch4.TalentLvlBurst()
	want := burstDMG[lvl] * (1 + asLohen(ch4).willMax()*burstWillRatio[lvl])
	if math.Abs(burstMult-want) > 1e-4 {
		t.Fatalf("c4 burst mult %v want %v", burstMult, want)
	}
	if ch4.Energy != c4Energy {
		t.Fatalf("c4 refund after burst %v", ch4.Energy)
	}

	c5, _ := makeCore(1)
	ch5 := addProfile(t, c5, keys.Lohen, 5, 8)
	if err := c5.Init(); err != nil {
		t.Fatal(err)
	}
	if ch5.TalentLvlBurst() != 10 {
		t.Fatalf("c5 burst index %d", ch5.TalentLvlBurst())
	}
}

func TestStellarA4AndHexerei(t *testing.T) {
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
	before := ch.Stat(attributes.ATKP)
	atk := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: ben.Index()}}
	c.Events.Emit(event.OnStellarSwirl, trg[0], atk)
	if math.Abs(ch.Stat(attributes.ATKP)-(before+a4ATK)) > 1e-9 {
		t.Fatalf("stellar swirl atk %v", ch.Stat(attributes.ATKP))
	}
	c.Events.Emit(event.OnStellarConduct, trg[0], atk)
	if math.Abs(ben.Stat(attributes.ATKP)-a4ATK) > 1e-9 {
		t.Fatalf("stellar conduct source atk %v", ben.Stat(attributes.ATKP))
	}
	advance(c, a4Duration)
	if math.Abs(ch.Stat(attributes.ATKP)-before) > 1e-9 {
		t.Fatalf("a4 after 8s %v", ch.Stat(attributes.ATKP))
	}

	hx, _ := makeCore(1)
	lo := addProfileWithParams(t, hx, keys.Lohen, 0, 10, map[string]int{"hexerei": 1})
	addProfile(t, hx, keys.Fischl, 0, 10)
	if err := hx.Init(); err != nil {
		t.Fatal(err)
	}
	hx.Player.SetActive(lo.Index())
	if err := hx.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	lc := asLohen(lo)
	lc.joy = joyCap
	lc.will = lc.willMax() * hexThreshold
	var dmg []float64
	var names []string
	// OnEnemyHit is emitted before attack mods are written into the snapshot.
	hx.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ae := args[1].(*info.AttackEvent)
		if ae.Info.ActorIndex != lo.Index() || ae.Info.AttackTag != attacks.AttackTagNormal {
			return
		}
		names = append(names, ae.Info.Abil)
		dmg = append(dmg, ae.Snapshot.Stats[attributes.DmgP])
	}, "hex")
	if err := hx.Player.Exec(action.ActionSkill, keys.Lohen, map[string]int{"etched": 1}); err != nil {
		t.Fatal(err)
	}
	advance(hx, 40)
	if err := hx.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(hx, attackHitmarks[0][0]+2)
	if len(dmg) != 1 || math.Abs(dmg[0]-hexBonus) > 1e-6 {
		t.Fatalf("hexerei normal %v %v", names, dmg)
	}
	lc.NormalCounter = 4
	if err := hx.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(hx, attackHitmarks[4][1]+2)
	if len(dmg) != 3 {
		t.Fatalf("n5 hits %v %v", names, dmg)
	}
	if math.Abs(dmg[1]-hexBonus) > 1e-6 || math.Abs(dmg[2]) > 1e-6 {
		t.Fatalf("n5 hexerei %v %v", names, dmg)
	}
	advance(hx, hexDuration)
	lc.NormalCounter = 0
	if err := hx.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(hx, attackHitmarks[0][0]+2)
	if math.Abs(dmg[len(dmg)-1]) > 1e-6 {
		t.Fatalf("hexerei after 6s %v", dmg)
	}

	off, _ := makeCore(1)
	solo := addProfileWithParams(t, off, keys.Lohen, 0, 10, map[string]int{"hexerei": 1})
	if err := off.Init(); err != nil {
		t.Fatal(err)
	}
	off.Player.SetActive(solo.Index())
	if err := off.Player.Exec(action.ActionSkill, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	asLohen(solo).joy = joyCap
	asLohen(solo).will = 100
	var bonus float64
	off.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ae := args[1].(*info.AttackEvent)
		if ae.Info.AttackTag == attacks.AttackTagNormal {
			bonus = ae.Snapshot.Stats[attributes.DmgP]
		}
	}, "off")
	if err := off.Player.Exec(action.ActionSkill, keys.Lohen, map[string]int{"etched": 1}); err != nil {
		t.Fatal(err)
	}
	advance(off, 40)
	if err := off.Player.Exec(action.ActionAttack, keys.Lohen, nil); err != nil {
		t.Fatal(err)
	}
	advance(off, attackHitmarks[0][0]+2)
	if math.Abs(bonus) > 1e-6 {
		t.Fatalf("solo hexerei bonus %v", bonus)
	}
}

func TestCadenceLeavesOtherCancels(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfileWithParams(t, c, keys.Lohen, 0, 10, map[string]int{"n1c_cadence": 1})
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	lc := asLohen(ch)
	n1, err := lc.Attack(nil)
	if err != nil {
		t.Fatal(err)
	}
	if n1.Frames(action.ActionCharge) != experimentalN1ToCA {
		t.Fatalf("n1 charge cancel %d", n1.Frames(action.ActionCharge))
	}
	n2, err := lc.Attack(nil)
	if err != nil {
		t.Fatal(err)
	}
	if n2.Frames(action.ActionCharge) != attackFrames[1][action.ActionCharge] {
		t.Fatalf("n2 charge cancel %d, baseline %d", n2.Frames(action.ActionCharge), attackFrames[1][action.ActionCharge])
	}
	if n2.Frames(action.ActionDash) != attackFrames[1][action.ActionDash] {
		t.Fatalf("n2 dash %d", n2.Frames(action.ActionDash))
	}
	ca, err := lc.ChargeAttack(nil)
	if err != nil {
		t.Fatal(err)
	}
	if ca.Frames(action.ActionDash) != chargeFrames[action.ActionDash] {
		t.Fatalf("ca dash %d", ca.Frames(action.ActionDash))
	}
	if ca.Frames(action.ActionSkill) != chargeFrames[action.ActionSkill] {
		t.Fatalf("ca skill %d", ca.Frames(action.ActionSkill))
	}
	if ca.Frames(action.ActionAttack) != experimentalCAToN1 {
		t.Fatalf("ca attack cancel %d", ca.Frames(action.ActionAttack))
	}
	sk, err := lc.Skill(nil)
	if err != nil {
		t.Fatal(err)
	}
	if sk.Frames(action.ActionAttack) != skillFrames[action.ActionAttack] {
		t.Fatalf("skill attack cancel %d", sk.Frames(action.ActionAttack))
	}
}
