package alyosha

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
	"github.com/genshinsim/gcsim/pkg/reactable"
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

func addAlyosha(t *testing.T, c *core.Core, cons, talent int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Alyosha, testhelper.TestWeaponKey)
	prof.Base.Cons = cons
	prof.Talents = info.TalentProfile{Attack: talent, Skill: talent, Burst: talent}
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("add char: %v", err)
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	return c.Player.ByIndex(idx)
}

func hitsOf(c *core.Core) *[]info.AttackInfo {
	var hits []info.AttackInfo
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hits = append(hits, args[1].(*info.AttackEvent).Info)
	}, "hits")
	return &hits
}

func TestBaseStats(t *testing.T) {
	c, _ := makeCore(1)
	ch := addAlyosha(t, c, 0, 1)
	if ch.Base.Element != attributes.Electro {
		t.Fatalf("element %v", ch.Base.Element)
	}
	if ch.Weapon.Class != info.WeaponClassSpear {
		t.Fatalf("weapon %v", ch.Weapon.Class)
	}
	if ch.EnergyMax != 70 {
		t.Fatalf("energy %v", ch.EnergyMax)
	}
	if math.Abs(ch.BaseStats[attributes.ER]-1.2667) > 1e-4 {
		t.Fatalf("er %v", ch.BaseStats[attributes.ER])
	}
}

func TestSkillMarkAndParticles(t *testing.T) {
	c, trg := makeCore(1)
	ch := addAlyosha(t, c, 0, 1)
	hits := hitsOf(c)
	var particles []character.Particle
	c.Events.Subscribe(event.OnParticleReceived, func(args ...any) {
		particles = append(particles, args[0].(character.Particle))
	}, "p")
	if err := c.Player.Exec(action.ActionSkill, keys.Alyosha, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillTapHitmark+2)
	if len(*hits) != 1 {
		t.Fatalf("hits %d", len(*hits))
	}
	hit := (*hits)[0]
	if hit.Element != attributes.Electro || hit.Durability != 25 || hit.ICDTag != attacks.ICDTagNone || !trg[0].AuraContains(attributes.Electro) {
		t.Fatalf("skill %+v aura %v", hit, trg[0].AuraContains(attributes.Electro))
	}
	if math.Abs(hit.Mult-skillTap[ch.TalentLvlSkill()]) > 1e-9 {
		t.Fatalf("mult %v", hit.Mult)
	}
	marks, _ := ch.Condition([]string{"marks"})
	if marks.(int) != 1 {
		t.Fatalf("marks %v", marks)
	}
	if ch.Stat(attributes.ATKP) != 0 {
		t.Fatalf("atk before activation %v", ch.Stat(attributes.ATKP))
	}
	advance(c, ch.ParticleDelay)
	if len(particles) != 1 || particles[0].Num != particleCount || particles[0].Ele != attributes.Electro {
		t.Fatalf("particles %+v", particles)
	}

	ch.ResetActionCooldown(action.ActionSkill)
	if err := c.Player.Exec(action.ActionSkill, keys.Alyosha, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillTapHitmark+2)
	marks, _ = ch.Condition([]string{"marks"})
	prec, _ := ch.Condition([]string{"precision"})
	if marks.(int) != 0 || prec.(int) != 1 {
		t.Fatalf("after activate marks %v precision %v", marks, prec)
	}
	want := precisionATK[ch.TalentLvlSkill()]
	if math.Abs(ch.Stat(attributes.ATKP)-want) > 1e-9 {
		t.Fatalf("atk %v want %v", ch.Stat(attributes.ATKP), want)
	}
}

func TestPrecisionIsActiveOnly(t *testing.T) {
	c, _ := makeCore(1)
	kprof := testhelper.DefaultProfile(keys.Alyosha, testhelper.TestWeaponKey)
	aidx, err := c.AddChar(kprof)
	if err != nil {
		t.Fatal(err)
	}
	bprof := testhelper.DefaultProfile(keys.Bennett, testhelper.TestWeaponKey)
	bidx, err := c.AddChar(bprof)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	ch := c.Player.ByIndex(aidx)
	ben := c.Player.ByIndex(bidx)
	c.Player.SetActive(aidx)
	if err := c.Player.Exec(action.ActionSkill, keys.Alyosha, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillTapHitmark+2)
	ch.ResetActionCooldown(action.ActionSkill)
	if err := c.Player.Exec(action.ActionSkill, keys.Alyosha, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillTapHitmark+2)
	if ch.Stat(attributes.ATKP) == 0 {
		t.Fatal("expected precision on alyosha")
	}
	c.Player.SetActive(bidx)
	if ch.Stat(attributes.ATKP) != 0 {
		t.Fatalf("off-field alyosha atk %v", ch.Stat(attributes.ATKP))
	}
	if math.Abs(ben.Stat(attributes.ATKP)-precisionATK[0]) > 1e-9 {
		t.Fatalf("bennett atk %v", ben.Stat(attributes.ATKP))
	}
}

func TestBurstFieldDogAndHeal(t *testing.T) {
	c, _ := makeCore(1)
	ch := addAlyosha(t, c, 0, 1)
	ch.SetHPByRatio(0.5)
	var heals []float64
	c.Events.Subscribe(event.OnHeal, func(args ...any) {
		heals = append(heals, args[4].(float64))
	}, "heal")
	var names []string
	var durs []float64
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		names = append(names, atk.Info.Abil)
		durs = append(durs, float64(atk.Info.Durability))
	}, "dmg")
	if err := c.Player.Exec(action.ActionBurst, keys.Alyosha, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, burstHitmark+2)
	if !ch.StatusIsActive(fieldKey) || len(names) != 1 || names[0] != "Fulgurite Hunting Field" {
		t.Fatalf("field %v %v", ch.StatusIsActive(fieldKey), names)
	}
	if durs[0] <= 0 {
		t.Fatal("field should apply electro")
	}
	advance(c, dogDelay)
	if len(names) != 2 || names[1] != "Tugarin" {
		t.Fatalf("dog %v", names)
	}
	if durs[1] != 0 {
		t.Fatalf("dog should share the burst icd, dur %v", durs[1])
	}
	if len(heals) != 1 || math.Abs(heals[0]-ch.TotalAtk()*a1HealRatio) > 1 {
		t.Fatalf("heals %v atk %v", heals, ch.TotalAtk())
	}
	if ch.StatusDuration(fieldKey) < 13*60 {
		t.Fatalf("duration %d", ch.StatusDuration(fieldKey))
	}
}

func TestC2DurationAndMark(t *testing.T) {
	c, _ := makeCore(1)
	ch := addAlyosha(t, c, 2, 1)
	if err := c.Player.Exec(action.ActionSkill, keys.Alyosha, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillTapHitmark+2)
	if err := c.Player.Exec(action.ActionBurst, keys.Alyosha, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, burstHitmark+dogDelay+2)
	marks, _ := ch.Condition([]string{"marks"})
	prec, _ := ch.Condition([]string{"precision"})
	if marks.(int) != 1 || prec.(int) != 0 {
		t.Fatalf("c2 should apply without activating, marks %v precision %v", marks, prec)
	}
	if ch.StatusDuration(fieldKey) < 19*60 {
		t.Fatalf("c2 duration %d", ch.StatusDuration(fieldKey))
	}
}

func TestC1ElectroReactionOnly(t *testing.T) {
	c, _ := makeCore(1)
	ch := addAlyosha(t, c, 1, 1)
	ch.Energy = 0
	c.Events.Emit(event.OnQuicken, nil, nil)
	c.Events.Emit(event.OnAggravate, nil, nil)
	if ch.Energy != 0 {
		t.Fatalf("quicken granted energy %v", ch.Energy)
	}
	c.Events.Emit(event.OnOverload, nil, nil)
	if math.Abs(ch.Energy-c1Energy) > 1e-9 {
		t.Fatalf("overload energy %v", ch.Energy)
	}
	c.Events.Emit(event.OnElectroCharged, nil, nil)
	if math.Abs(ch.Energy-c1Energy) > 1e-9 {
		t.Fatalf("icd failed %v", ch.Energy)
	}
}

func TestC6SecondStack(t *testing.T) {
	c, _ := makeCore(1)
	ch := addAlyosha(t, c, 6, 1)
	// apply, activate, apply, activate: the second activation is the C6 stack
	for range 4 {
		ch.ResetActionCooldown(action.ActionSkill)
		if err := c.Player.Exec(action.ActionSkill, keys.Alyosha, nil); err != nil {
			t.Fatal(err)
		}
		advance(c, skillTapHitmark+2)
	}
	prec, _ := ch.Condition([]string{"precision"})
	if prec.(int) != 2 {
		t.Fatalf("stacks %v", prec)
	}
	// DefaultProfile contributes 100 EM on top of the C6 bonus.
	if math.Abs(ch.Stat(attributes.EM)-(100+c6EM)) > 1e-9 {
		t.Fatalf("em %v", ch.Stat(attributes.EM))
	}
	want := precisionATK[ch.TalentLvlSkill()] * 2
	if math.Abs(ch.Stat(attributes.ATKP)-want) > 1e-6 {
		t.Fatalf("atk %v want %v", ch.Stat(attributes.ATKP), want)
	}
}

func TestStellarBonusNotGrantedWithoutField(t *testing.T) {
	c, _ := makeCore(1)
	ch := addAlyosha(t, c, 0, 1)
	ac := ch.Character.(*char)
	for range 2 {
		ch.ResetActionCooldown(action.ActionSkill)
		if err := c.Player.Exec(action.ActionSkill, keys.Alyosha, nil); err != nil {
			t.Fatal(err)
		}
		advance(c, skillTapHitmark+2)
	}
	if ac.stellarBonusFor(ch.Index(), attacks.AttackTagDirectStellarConduct) != 0 {
		t.Fatal("stellar bonus granted without a polestar field")
	}
	ch.AddStatus(reactable.PolestarFieldKey, 60, true)
	ch.ResetActionCooldown(action.ActionSkill)
	if err := c.Player.Exec(action.ActionSkill, keys.Alyosha, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillTapHitmark+2)
	ch.ResetActionCooldown(action.ActionSkill)
	if err := c.Player.Exec(action.ActionSkill, keys.Alyosha, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillTapHitmark+2)
	got := ac.stellarBonusFor(ch.Index(), attacks.AttackTagDirectStellarConduct)
	if math.Abs(got-stellarBonusValue) > 1e-9 {
		t.Fatalf("stellar %v", got)
	}
	if ac.stellarBonusFor(ch.Index(), attacks.AttackTagElementalBurst) != 0 {
		t.Fatal("stellar bonus leaked onto a non-reaction tag")
	}
}

func TestA4Caps(t *testing.T) {
	c, _ := makeCore(1)
	ch := addAlyosha(t, c, 0, 1)
	ac := ch.Character.(*char)
	bonus := ac.Stat(attributes.ER) * a4PerER
	if bonus > a4Cap {
		bonus = a4Cap
	}
	if bonus <= 0 || bonus >= a4Cap {
		t.Fatalf("uncapped naked bonus %v er %v", bonus, ac.Stat(attributes.ER))
	}
}
