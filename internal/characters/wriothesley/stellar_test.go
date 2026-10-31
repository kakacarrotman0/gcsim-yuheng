package wriothesley

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/kaeya"
	_ "github.com/genshinsim/gcsim/internal/characters/odette"
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
	"github.com/genshinsim/gcsim/pkg/reactable"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

const (
	n3L1    = 0.672228
	n5L1    = 0.90742
	caL1    = 1.5296
	skillL1 = 1.431695
	n3C3    = 0.859826
	n5C3    = 1.160654
	caC3    = 1.912
)

func init() {
	testhelper.RegisterTestWeapon()
}

type party struct {
	c    *core.Core
	trg  *enemy.Enemy
	wrio *character.CharWrapper
	mate *character.CharWrapper
}

func setup(t *testing.T, cons int, params map[string]int, enabler bool) *party {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
	trg := enemy.New(c, info.EnemyProfile{
		Level:  100,
		Resist: make(map[attributes.Element]float64),
		Pos:    info.Coord{X: 0, Y: 0, R: 1},
	})
	c.Combat.AddEnemy(trg)
	c.Combat.DefaultTarget = trg.Key()

	if enabler {
		add(t, c, keys.Odette, 0, nil)
	}
	wrio := add(t, c, keys.Wriothesley, cons, params)
	mate := add(t, c, keys.Kaeya, 0, nil)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(wrio.Index())
	return &party{c: c, trg: trg, wrio: wrio, mate: mate}
}

func add(t *testing.T, c *core.Core, key keys.Char, cons int, params map[string]int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(key, testhelper.TestWeaponKey)
	prof.Base.Cons = cons
	prof.Base.Ascension = 6
	prof.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	for k, v := range params {
		prof.Params[k] = v
	}
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatal(err)
	}
	return c.Player.ByIndex(idx)
}

func advance(c *core.Core, n int) {
	for range n {
		c.F++
		c.Tick()
	}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func (p *party) hits() *[]info.AttackInfo {
	var hits []info.AttackInfo
	p.c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex == p.wrio.Index() {
			hits = append(hits, atk.Info)
		}
	}, "wrio-hits")
	return &hits
}

func (p *party) conduct(t *testing.T) {
	t.Helper()
	var conduct, plain int
	p.c.Events.Subscribe(event.OnStellarConduct, func(...any) { conduct++ }, "conduct")
	p.c.Events.Subscribe(event.OnSuperconduct, func(...any) { plain++ }, "super")
	p.trg.AttachOrRefill(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25, ActorIndex: p.wrio.Index()},
	})
	p.trg.React(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Electro, Durability: 25, ActorIndex: p.wrio.Index()},
	})
	if conduct != 1 || plain != 0 {
		t.Fatalf("conduct %d superconduct %d", conduct, plain)
	}
	if !p.wrio.StatusIsActive(reactable.PolestarFieldKey) {
		t.Fatal("polestar status missing")
	}
}

func (p *party) skill(t *testing.T) {
	t.Helper()
	if err := p.c.Player.Exec(action.ActionSkill, keys.Wriothesley, nil); err != nil {
		t.Fatal(err)
	}
}

func (p *party) punch(t *testing.T, n int) {
	t.Helper()
	for range n {
		if err := p.c.Player.Exec(action.ActionAttack, keys.Wriothesley, nil); err != nil {
			t.Fatal(err)
		}
	}
}

func (p *party) charge(t *testing.T) {
	t.Helper()
	if p.c.Player.LastAction.Type != action.ActionAttack {
		p.punch(t, 1)
	}
	if err := p.c.Player.Exec(action.ActionCharge, keys.Wriothesley, nil); err != nil {
		t.Fatal(err)
	}
}

func named(hits []info.AttackInfo, name string) []info.AttackInfo {
	var out []info.AttackInfo
	for i := range hits {
		if hits[i].Abil == name {
			out = append(out, hits[i])
		}
	}
	return out
}

func TestRevelationDefaultsOn(t *testing.T) {
	on, _ := setup(t, 0, nil, false).wrio.Condition([]string{"revelation"})
	if on != true {
		t.Fatal("omitted revelation param")
	}
	off, _ := setup(t, 0, map[string]int{"revelation": 0}, false).wrio.Condition([]string{"revelation"})
	if off != false {
		t.Fatal("revelation=0")
	}
}

func TestNormalsAndRebukeOutsideRadiance(t *testing.T) {
	p := setup(t, 0, nil, false)
	hits := p.hits()
	p.skill(t)
	p.punch(t, 5)
	advance(p.c, 45)
	n3 := named(*hits, "Normal 3 (Enhanced)")
	n1 := named(*hits, "Normal 1 (Enhanced)")
	if len(n3) != 1 || len(n1) != 1 {
		t.Fatalf("n1 %d n3 %d", len(n1), len(n3))
	}
	if n3[0].AttackTag != attacks.AttackTagNormal || n3[0].Durability != 25 || !near(n3[0].Mult, n3L1*skillL1) {
		t.Fatalf("n3 %+v", n3[0])
	}
	if !p.trg.AuraContains(attributes.Cryo) {
		t.Fatal("enhanced normal did not apply cryo")
	}
	if p.wrio.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagDirectStellarConduct}) != 0.3 {
		t.Fatal("revelation react bonus")
	}

	before := p.c.Player.Stam
	p.wrio.SetHPByRatio(1)
	p.charge(t)
	advance(p.c, chargeHitmark+2)
	ca := named(*hits, "Charge Attack")
	if len(ca) != 1 || ca[0].AttackTag != attacks.AttackTagExtra || !near(ca[0].Mult, caL1) {
		t.Fatalf("ca %+v", ca)
	}
	if !near(p.c.Player.Stam, before-50) {
		t.Fatalf("stam %v from %v", p.c.Player.Stam, before)
	}

	*hits = nil
	p.wrio.SetHPByRatio(0.55)
	p.c.Player.Stam = before
	var healed float64
	p.c.Events.Subscribe(event.OnHeal, func(args ...any) {
		healed = args[4].(float64)
	}, "rebuke-heal")
	p.charge(t)
	advance(p.c, chargeHitmark+2)
	rebuke := named(*hits, "Rebuke: Vaulting Fist")
	if len(rebuke) != 1 || rebuke[0].AttackTag != attacks.AttackTagExtra || len(named(*hits, "Luster: Vaulting Fist"+stellarConductText)) != 0 {
		t.Fatalf("rebuke %+v", *hits)
	}
	if !near(p.c.Player.Stam, before) || !near(healed, 0.3*p.wrio.MaxHP()) {
		t.Fatalf("stam %v heal %v max %v", p.c.Player.Stam, healed, p.wrio.MaxHP())
	}
}

func TestPolestarConvertsTheThirdAndFifthStrikes(t *testing.T) {
	p := setup(t, 0, nil, true)
	p.conduct(t)
	hits := p.hits()
	var extra int
	p.c.Events.Subscribe(event.OnStellarConduct, func(...any) { extra++ }, "extra-conduct")
	p.skill(t)
	p.punch(t, 5)
	advance(p.c, 45)
	if extra != 0 {
		t.Fatalf("normals emitted stellar conduct %d", extra)
	}
	n1 := named(*hits, "Normal 1 (Enhanced)")
	n3 := named(*hits, "Normal 3 (Enhanced)"+stellarConductText)
	n5 := named(*hits, "Normal 5 (Enhanced)"+stellarConductText)
	if len(n1) != 1 || len(n3) != 1 || len(n5) != 1 || len(named(*hits, "Normal 3 (Enhanced)")) != 0 {
		t.Fatalf("hits %+v", names(*hits))
	}
	if n1[0].AttackTag != attacks.AttackTagNormal || n1[0].Durability != 25 {
		t.Fatalf("n1 %+v", n1[0])
	}
	if n3[0].AttackTag != attacks.AttackTagDirectStellarConduct || n3[0].Durability != 0 || n3[0].IgnoreDefPercent != 1 || !near(n3[0].Mult, n3L1*skillL1*0.6) {
		t.Fatalf("n3 %+v", n3[0])
	}
	if n5[0].AttackTag != attacks.AttackTagDirectStellarConduct || n5[0].Durability != 0 || !near(n5[0].Mult, n5L1*skillL1*0.8) {
		t.Fatalf("n5 %+v", n5[0])
	}
	if len(n5[0].AdditionalTags) != 1 || n5[0].AdditionalTags[0] != attacks.AttackTagWrioSSCN5 {
		t.Fatalf("n5 tags %v", n5[0].AdditionalTags)
	}
}

func TestLeavingTheFieldRestoresCryoStrikes(t *testing.T) {
	p := setup(t, 0, nil, true)
	p.conduct(t)
	p.c.Combat.SetPlayerPos(info.Point{X: 100, Y: 100})
	left := p.wrio.StatusDuration(reactable.PolestarFieldKey)
	advance(p.c, left+5)
	if p.wrio.StatusIsActive(reactable.PolestarFieldKey) {
		t.Fatal("polestar refreshed outside the field")
	}
	p.c.Combat.SetPlayerPos(info.Point{})
	hits := p.hits()
	p.skill(t)
	p.punch(t, 3)
	advance(p.c, 25)
	n3 := named(*hits, "Normal 3 (Enhanced)")
	if len(n3) != 1 || n3[0].AttackTag != attacks.AttackTagNormal || !near(n3[0].Mult, n3L1*skillL1) {
		t.Fatalf("n3 %+v", n3)
	}
}

func TestRevelationOffKeepsTheBaseKitInsideTheField(t *testing.T) {
	p := setup(t, 0, map[string]int{"revelation": 0}, true)
	p.conduct(t)
	if p.wrio.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagDirectStellarConduct}) != 0 {
		t.Fatal("react bonus without revelation")
	}
	hits := p.hits()
	p.skill(t)
	p.punch(t, 3)
	advance(p.c, 25)
	n3 := named(*hits, "Normal 3 (Enhanced)")
	if len(n3) != 1 || n3[0].AttackTag != attacks.AttackTagNormal || n3[0].Durability != 25 {
		t.Fatalf("n3 %+v", n3)
	}
	p.wrio.SetHPByRatio(0.55)
	ready, _ := p.wrio.Condition([]string{"gracious-rebuke-unveiling"})
	if ready != false {
		t.Fatal("luster without revelation")
	}
}

func TestLusterHealICDAndStamina(t *testing.T) {
	p := setup(t, 0, nil, true)
	p.conduct(t)
	p.skill(t)
	hits := p.hits()
	var heals int
	p.c.Events.Subscribe(event.OnHeal, func(args ...any) {
		hi := args[0].(*info.HealInfo)
		if hi.Message == a1HealName && near(args[4].(float64), 0.3*p.wrio.MaxHP()) {
			heals++
		}
	}, "luster-heal")
	p.wrio.SetHPByRatio(0.55)
	stam := p.c.Player.Stam
	p.charge(t)
	advance(p.c, chargeHitmark+2)
	luster := named(*hits, "Luster: Vaulting Fist"+stellarConductText)
	if len(luster) != 1 || luster[0].AttackTag != attacks.AttackTagDirectStellarConduct || luster[0].Durability != 0 || luster[0].IgnoreDefPercent != 1 || !near(luster[0].Mult, caL1) {
		t.Fatalf("luster %+v", luster)
	}
	if len(named(*hits, "Rebuke: Vaulting Fist")) != 0 || len(named(*hits, "Charge Attack")) != 0 {
		t.Fatalf("duplicate charge %+v", names(*hits))
	}
	if heals != 1 || !near(p.c.Player.Stam, stam) {
		t.Fatalf("heals %d stam %v", heals, p.c.Player.Stam)
	}
	p.wrio.SetHPByRatio(0.55)
	p.charge(t)
	advance(p.c, chargeHitmark+2)
	if heals != 1 || len(named(*hits, "Luster: Vaulting Fist"+stellarConductText)) != 2 {
		t.Fatalf("second luster heals %d hits %d", heals, len(named(*hits, "Luster: Vaulting Fist"+stellarConductText)))
	}
	advance(p.c, 2*60)
	p.wrio.SetHPByRatio(0.55)
	p.charge(t)
	advance(p.c, chargeHitmark+2)
	if heals != 2 {
		t.Fatalf("heal after icd %d", heals)
	}
}

func TestConductNeedsAnEnabler(t *testing.T) {
	p := setup(t, 0, nil, false)
	var conduct, plain int
	p.c.Events.Subscribe(event.OnStellarConduct, func(...any) { conduct++ }, "conduct")
	p.c.Events.Subscribe(event.OnSuperconduct, func(...any) { plain++ }, "super")
	p.trg.AttachOrRefill(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25, ActorIndex: p.wrio.Index()},
	})
	p.trg.React(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Electro, Durability: 25, ActorIndex: p.wrio.Index()},
	})
	if conduct != 0 || plain != 1 || p.wrio.StatusIsActive(reactable.PolestarFieldKey) {
		t.Fatalf("conduct %d superconduct %d field %v", conduct, plain, p.wrio.StatusIsActive(reactable.PolestarFieldKey))
	}
}

func TestC1ChainsFiveSecondBonusesAndExtendsTheSkillOnce(t *testing.T) {
	p := setup(t, 1, nil, true)
	p.conduct(t)
	p.skill(t)
	p.punch(t, 5)
	advance(p.c, 45)
	probe := info.AttackInfo{AttackTag: attacks.AttackTagDirectStellarConduct, AdditionalTags: []attacks.AttackTag{attacks.AttackTagWrioSSCCA}}
	if !p.wrio.ReactBonusModIsActive(c1LusterBonus) || !near(p.wrio.ReactBonus(probe), 0.8) {
		t.Fatalf("luster bonus %v active %v", p.wrio.ReactBonus(probe), p.wrio.ReactBonusModIsActive(c1LusterBonus))
	}
	ready, _ := p.wrio.Condition([]string{"gracious-rebuke"})
	if ready != false {
		t.Fatal("stellar n5 granted gracious rebuke")
	}
	start := p.c.F
	before := p.wrio.StatusDuration(skillKey)
	p.wrio.SetHPByRatio(0.55)
	p.charge(t)
	advance(p.c, chargeHitmark+2)
	elapsed := p.c.F - start
	if p.wrio.StatusDuration(skillKey) != before-elapsed+4*60 {
		t.Fatalf("skill dur %d want %d", p.wrio.StatusDuration(skillKey), before-elapsed+4*60)
	}
	if p.wrio.ReactBonusModIsActive(c1LusterBonus) || !p.wrio.ReactBonusModIsActive(c1N5Bonus) {
		t.Fatal("bonus did not move from luster to n5")
	}
	n5 := info.AttackInfo{AttackTag: attacks.AttackTagDirectStellarConduct, AdditionalTags: []attacks.AttackTag{attacks.AttackTagWrioSSCN5}}
	if !near(p.wrio.ReactBonus(n5), 0.8) {
		t.Fatalf("n5 bonus %v", p.wrio.ReactBonus(n5))
	}
	again := p.c.F
	dur := p.wrio.StatusDuration(skillKey)
	p.wrio.SetHPByRatio(0.55)
	p.charge(t)
	advance(p.c, chargeHitmark+2)
	if p.wrio.StatusDuration(skillKey) != dur-(p.c.F-again) {
		t.Fatal("skill extended twice")
	}
}

func TestC2ChangesFractionsOnlyAtFiveStacks(t *testing.T) {
	p := setup(t, 2, nil, true)
	p.skill(t)
	for range 4 {
		p.c.Player.Drain(info.DrainInfo{ActorIndex: p.wrio.Index(), Abil: "edict", Amount: 1})
	}
	stacks, _ := p.wrio.Condition([]string{"prosecution-edict"})
	if stacks != 4 {
		t.Fatalf("stacks %v", stacks)
	}
	hits := p.hits()
	p.wrio.ResetNormalCounter()
	p.punch(t, 1)
	advance(p.c, 20)
	n1 := named(*hits, "Normal 1 (Enhanced)")
	if len(n1) != 1 || !near(n1[0].Mult, 0.533596*skillL1) {
		t.Fatalf("four stacks %+v", n1)
	}
	p.c.Player.Drain(info.DrainInfo{ActorIndex: p.wrio.Index(), Abil: "edict", Amount: 1})
	*hits = nil
	p.wrio.ResetNormalCounter()
	p.punch(t, 1)
	p.charge(t)
	advance(p.c, chargeHitmark+2)
	n1 = named(*hits, "Normal 1 (Enhanced)")
	ca := named(*hits, "Charge Attack")
	if len(n1) != 1 || !near(n1[0].Mult, 0.533596*skillL1*1.25) || len(ca) != 1 || !near(ca[0].Mult, caL1*1.3) {
		t.Fatalf("n1 %v ca %v", n1, ca)
	}

	p.conduct(t)
	*hits = nil
	p.wrio.ResetNormalCounter()
	p.punch(t, 5)
	advance(p.c, 45)
	n3 := named(*hits, "Normal 3 (Enhanced)"+stellarConductText)
	n5 := named(*hits, "Normal 5 (Enhanced)"+stellarConductText)
	if len(n3) != 1 || !near(n3[0].Mult, n3L1*skillL1*0.9) || len(n5) != 1 || !near(n5[0].Mult, n5L1*skillL1*1.2) {
		t.Fatalf("n3 %v n5 %v", n3, n5)
	}
	*hits = nil
	p.wrio.SetHPByRatio(0.55)
	p.charge(t)
	advance(p.c, chargeHitmark+2)
	luster := named(*hits, "Luster: Vaulting Fist"+stellarConductText)
	if len(luster) != 1 || !near(luster[0].Mult, caL1*1.5) {
		t.Fatalf("luster %v", luster)
	}
}

func TestC4HealsMoreOutsideAndBuffsSpeedInside(t *testing.T) {
	p := setup(t, 4, nil, true)
	var healed float64
	p.c.Events.Subscribe(event.OnHeal, func(args ...any) {
		healed = args[4].(float64)
	}, "c4-heal")
	p.conduct(t)
	if !near(p.wrio.Stat(attributes.AtkSpd), 0.2) || !near(p.mate.Stat(attributes.AtkSpd), 0.1) {
		t.Fatalf("wrio spd %v mate %v", p.wrio.Stat(attributes.AtkSpd), p.mate.Stat(attributes.AtkSpd))
	}
	if p.wrio.DamageReduction(p.wrio.Index()) != 0 {
		t.Fatal("incoming damage reduction is not implemented")
	}
	p.skill(t)
	p.wrio.SetHPByRatio(0.55)
	p.charge(t)
	advance(p.c, chargeHitmark+2)
	if !near(healed, 0.3*p.wrio.MaxHP()) {
		t.Fatalf("luster heal %v", healed)
	}
	p.c.Combat.SetPlayerPos(info.Point{X: 100, Y: 100})
	left := p.wrio.StatusDuration(reactable.PolestarFieldKey)
	advance(p.c, left+5)
	if p.wrio.Stat(attributes.AtkSpd) != 0 || p.mate.Stat(attributes.AtkSpd) != 0 {
		t.Fatalf("spd after exit wrio %v mate %v", p.wrio.Stat(attributes.AtkSpd), p.mate.Stat(attributes.AtkSpd))
	}
	p.c.Combat.SetPlayerPos(info.Point{})
	healed = 0
	p.wrio.SetHPByRatio(0.55)
	p.charge(t)
	advance(p.c, chargeHitmark+2)
	if !near(healed, 0.5*p.wrio.MaxHP()) {
		t.Fatalf("rebuke heal %v", healed)
	}
}

func TestC6IcicleRatioAndClassification(t *testing.T) {
	p := setup(t, 6, nil, false)
	hits := p.hits()
	p.skill(t)
	p.wrio.SetHPByRatio(0.55)
	p.charge(t)
	advance(p.c, chargeHitmark+2)
	rebuke := named(*hits, "Rebuke: Vaulting Fist")
	icicle := named(*hits, "Rebuke: Vaulting Fist (C6)")
	if len(rebuke) != 1 || len(icicle) != 1 || !near(icicle[0].Mult, rebuke[0].Mult) || !near(rebuke[0].Mult, caC3) || icicle[0].AttackTag != attacks.AttackTagExtra {
		t.Fatalf("rebuke %+v icicle %+v", rebuke, icicle)
	}

	q := setup(t, 6, nil, true)
	q.conduct(t)
	qHits := q.hits()
	q.skill(t)
	q.punch(t, 5)
	advance(q.c, 45)
	n5 := named(*qHits, "Normal 5 (Enhanced)"+stellarConductText)
	extra := named(*qHits, "Normal 5 (Enhanced) (C6)"+stellarConductText)
	if len(n5) != 1 || len(extra) != 1 || !near(extra[0].Mult, n5[0].Mult*c6StellarIcicleRatio) {
		t.Fatalf("n5 %v extra %v", n5, extra)
	}
	if extra[0].AttackTag != attacks.AttackTagDirectStellarConduct || len(extra[0].AdditionalTags) != 0 || extra[0].Durability != 0 {
		t.Fatalf("icicle %+v", extra[0])
	}
	if !near(n5[0].Mult, n5C3*skillL1*0.8) {
		t.Fatalf("n5 mult %v", n5[0].Mult)
	}
	var baseCR, baseCD, buffCR, buffCD float64
	plain := setup(t, 0, nil, true)
	plain.conduct(t)
	plain.c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.Abil == "Normal 1 (Enhanced)" && atk.Info.ActorIndex == plain.wrio.Index() {
			baseCR = atk.Snapshot.Stats[attributes.CR]
			baseCD = atk.Snapshot.Stats[attributes.CD]
		}
	}, "c0-crit")
	plain.skill(t)
	plain.punch(t, 1)
	advance(plain.c, 20)
	q.c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.Abil == "Normal 1 (Enhanced)" && atk.Info.ActorIndex == q.wrio.Index() {
			buffCR = atk.Snapshot.Stats[attributes.CR]
			buffCD = atk.Snapshot.Stats[attributes.CD]
		}
	}, "c6-crit")
	q.wrio.ResetNormalCounter()
	q.punch(t, 1)
	advance(q.c, 20)
	if !near(buffCR-baseCR, 0.1) || !near(buffCD-baseCD, 0.8) {
		t.Fatalf("crit delta cr %v cd %v", buffCR-baseCR, buffCD-baseCD)
	}
}

func names(hits []info.AttackInfo) []string {
	out := make([]string, len(hits))
	for i := range hits {
		out[i] = hits[i].Abil
	}
	return out
}
