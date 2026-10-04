package vodyanitsa

import (
	"math"
	"testing"

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

func addVody(t *testing.T, c *core.Core, cons int, talent int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Vodyanitsa, testhelper.TestWeaponKey)
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

func TestBaseStats(t *testing.T) {
	c, _ := makeCore(1)
	ch := addVody(t, c, 0, 1)
	if ch.Base.Element != attributes.Hydro {
		t.Fatalf("element %v", ch.Base.Element)
	}
	if ch.EnergyMax != 60 {
		t.Fatalf("energy %v", ch.EnergyMax)
	}
	if math.Abs(ch.BaseStats[attributes.HPP]-0.288) > 1e-9 {
		t.Fatalf("hp%% ascension %v", ch.BaseStats[attributes.HPP])
	}
	if ch.BaseStats[attributes.BaseHP] < 10000 {
		t.Fatalf("base hp %v", ch.BaseStats[attributes.BaseHP])
	}
	if ch.Base.Ascension < 4 {
		t.Fatalf("ascension %v", ch.Base.Ascension)
	}
}

func TestNormalAttackScalingAndGauge(t *testing.T) {
	c, trg := makeCore(1)
	ch := addVody(t, c, 0, 1)
	var hits []info.AttackInfo
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hits = append(hits, args[1].(*info.AttackEvent).Info)
	}, "na")
	if err := c.Player.Exec(action.ActionAttack, keys.Vodyanitsa, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, attackHitmarks[0]+2)
	if len(hits) != 1 {
		t.Fatalf("hits %d", len(hits))
	}
	if hits[0].Abil != "Normal 0" || hits[0].Element != attributes.Hydro || hits[0].Durability != 25 {
		t.Fatalf("hit %+v", hits[0])
	}
	if hits[0].ICDTag != attacks.ICDTagNormalAttack || hits[0].ICDGroup != attacks.ICDGroupDefault {
		t.Fatalf("icd %+v %+v", hits[0].ICDTag, hits[0].ICDGroup)
	}
	if math.Abs(hits[0].Mult-attack[0][ch.TalentLvlAttack()]) > 1e-9 {
		t.Fatalf("mult %v", hits[0].Mult)
	}
	if !trg[0].AuraContains(attributes.Hydro) {
		t.Fatal("expected hydro aura")
	}
}

func TestSkillHornHealParticlesAndRes(t *testing.T) {
	c, trg := makeCore(1)
	ch := addVody(t, c, 0, 10)
	var horns int
	var initial int
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		abil := args[1].(*info.AttackEvent).Info.Abil
		switch abil {
		case "Horn of Spring's Call":
			horns++
		case "Rechitativ: Sonorous Dawn":
			initial++
		}
	}, "hits")
	var heals int
	var healAmt float64
	c.Events.Subscribe(event.OnHeal, func(args ...any) {
		heals++
		healAmt = args[4].(float64)
	}, "heal")
	var particles float64
	c.Events.Subscribe(event.OnParticleReceived, func(args ...any) {
		p := args[0].(character.Particle)
		if p.Ele != attributes.Hydro {
			t.Fatalf("particle ele %v", p.Ele)
		}
		particles += p.Num
	}, "particles")

	if err := c.Player.Exec(action.ActionSkill, keys.Vodyanitsa, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, songDur+ch.ParticleDelay+hornDamageDelay+5)

	if initial != 1 {
		t.Fatalf("initial hits %d", initial)
	}
	if horns != 5 {
		t.Fatalf("horns %d", horns)
	}
	if heals != 10 {
		t.Fatalf("heals %d", heals)
	}
	wantHeal := healFlat[ch.TalentLvlSkill()] + healPct[ch.TalentLvlSkill()]*ch.MaxHP()
	if math.Abs(healAmt-wantHeal) > 1 {
		t.Fatalf("heal %v want %v", healAmt, wantHeal)
	}
	if particles != 5*particleCount {
		t.Fatalf("particles %v", particles)
	}
	if !trg[0].ResistModIsActive(hydroShredKey) && !trg[0].ResistModIsActive(cryoShredKey) {
		// shred lasts 6s from the last horn, which is inside the song; after songDur+5 it may have expired
	}
	lead, _ := ch.Condition([]string{"lead"})
	// initial + 5 horns, on field, consume lead
	if lead.(int) != leadMax-6 {
		t.Fatalf("lead %v", lead)
	}
}

func TestSkillResActiveAfterHit(t *testing.T) {
	c, trg := makeCore(1)
	addVody(t, c, 0, 10)
	if err := c.Player.Exec(action.ActionSkill, keys.Vodyanitsa, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillHitmark+2)
	if !trg[0].ResistModIsActive(hydroShredKey) || !trg[0].ResistModIsActive(cryoShredKey) {
		t.Fatal("expected hydro and cryo res shred")
	}
}

func TestBurstSongRatioAndEnergy(t *testing.T) {
	c, _ := makeCore(1)
	ch := addVody(t, c, 0, 10)
	var flat []float64
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.Abil == "Koda: Sink With Thee" {
			flat = append(flat, ai.FlatDmg)
		}
	}, "burst")

	ch.Energy = 0
	if _, fail := ch.ActionReady(action.ActionBurst, nil); fail != action.InsufficientEnergy {
		t.Fatalf("expected energy gate, got %v energy %v", fail, ch.Energy)
	}
	ch.AddEnergy("test", 60)
	if err := c.Player.Exec(action.ActionBurst, keys.Vodyanitsa, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, burstHitmark+2)
	if len(flat) != 1 {
		t.Fatalf("bursts %d", len(flat))
	}
	want := burst[ch.TalentLvlBurst()] * ch.MaxHP()
	if math.Abs(flat[0]-want) > 1 {
		t.Fatalf("no-song burst %v want %v", flat[0], want)
	}
	if ch.Energy != 0 {
		t.Fatalf("energy after burst %v", ch.Energy)
	}

	ch.AddEnergy("test", 60)
	ch.ResetActionCooldown(action.ActionBurst)
	if err := c.Player.Exec(action.ActionSkill, keys.Vodyanitsa, nil); err != nil {
		t.Fatal(err)
	}
	if err := c.Player.Exec(action.ActionBurst, keys.Vodyanitsa, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, burstHitmark+2)
	if len(flat) != 2 {
		t.Fatalf("bursts %d", len(flat))
	}
	wantSong := burst[ch.TalentLvlBurst()] * (1 + burstBonus[ch.TalentLvlBurst()]) * ch.MaxHP()
	if math.Abs(flat[1]-wantSong)/wantSong > 0.02 {
		t.Fatalf("song burst %v want about %v", flat[1], wantSong)
	}
}

func TestC2ExtendsSong(t *testing.T) {
	c, _ := makeCore(1)
	addVody(t, c, 2, 6)
	var horns int
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		if args[1].(*info.AttackEvent).Info.Abil == "Horn of Spring's Call" {
			horns++
		}
	}, "horns")
	if err := c.Player.Exec(action.ActionSkill, keys.Vodyanitsa, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, songDur+c2SongExtend+5)
	if horns != 8 {
		t.Fatalf("c2 horns %d", horns)
	}
}

func TestC1AndC4(t *testing.T) {
	c, _ := makeCore(1)
	ch := addVody(t, c, 4, 6)
	atkBefore := ch.TotalAtk()
	if err := c.Player.Exec(action.ActionSkill, keys.Vodyanitsa, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, healPreDelay+healInterval+2)
	n, _ := ch.Condition([]string{"c4"})
	if n.(int) != 1 {
		t.Fatalf("c4 stacks %v", n)
	}
	if ch.TotalAtk()-atkBefore < c1ATKRatio*ch.MaxHP()*0.9 {
		t.Fatalf("c1 atk delta %v", ch.TotalAtk()-atkBefore)
	}
	// drop below 40% before the next heal; that heal is boosted and adds no stack
	ch.SetHPByRatio(0.1)
	advance(c, healInterval)
	n, _ = ch.Condition([]string{"c4"})
	if n.(int) != 1 {
		t.Fatalf("c4 stacks after low hp heal %v", n)
	}
}

func TestC3C5TalentLevels(t *testing.T) {
	c, _ := makeCore(1)
	ch := addVody(t, c, 6, 10)
	if ch.TalentLvlSkill() != 12 {
		t.Fatalf("c3 skill index %d", ch.TalentLvlSkill())
	}
	if ch.TalentLvlBurst() != 12 {
		t.Fatalf("c5 burst index %d", ch.TalentLvlBurst())
	}
	if math.Abs(skill[ch.TalentLvlSkill()]-0.06953) > 1e-6 {
		t.Fatalf("c3 skill ratio %v", skill[ch.TalentLvlSkill()])
	}
	if math.Abs(burst[ch.TalentLvlBurst()]-0.970632) > 1e-6 {
		t.Fatalf("c5 burst ratio %v", burst[ch.TalentLvlBurst()])
	}
}

func TestC6HydroDmg(t *testing.T) {
	c, _ := makeCore(1)
	addVody(t, c, 6, 10)
	var dmgP float64
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ae := args[1].(*info.AttackEvent)
		if ae.Info.Abil == "Normal 0" {
			dmgP = ae.Snapshot.Stats[attributes.DmgP]
		}
	}, "dmg")
	if err := c.Player.Exec(action.ActionSkill, keys.Vodyanitsa, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillFrames[action.ActionAttack])
	if err := c.Player.Exec(action.ActionAttack, keys.Vodyanitsa, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, attackHitmarks[0]+5)
	if math.Abs(dmgP-c6DmgP) > 1e-6 {
		t.Fatalf("dmg%% %v", dmgP)
	}
}

func TestA4BelowThresholdAndCap(t *testing.T) {
	c, _ := makeCore(1)
	ch := addVody(t, c, 0, 10)
	raw := ch.Character.(*char)
	// A naked level 90 is below the 40000 Max HP threshold.
	if ch.MaxHP() >= a4MinHP {
		t.Fatalf("expected naked level 90 below threshold, hp %v", ch.MaxHP())
	}
	if raw.concertoFlat(false) != 0 || raw.concertoFlat(true) != 0 {
		t.Fatal("bonus should be zero below 40000 hp")
	}
	ch.BaseStats[attributes.HP] = 50000 - ch.MaxHP()
	got := raw.concertoFlat(false)
	want := (ch.MaxHP() - a4MinHP) * a4HydroPerKilo * 0.001
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("hydro bonus %v want %v hp %v", got, want, ch.MaxHP())
	}
	ch.BaseStats[attributes.HP] = 80000
	if math.Abs(raw.concertoFlat(false)-a4HydroCap) > 1e-6 {
		t.Fatalf("hydro cap %v", raw.concertoFlat(false))
	}
	if math.Abs(raw.concertoFlat(true)-a4StellarCap) > 1e-6 {
		t.Fatalf("stellar cap %v", raw.concertoFlat(true))
	}
}

func TestA1DetonateShred(t *testing.T) {
	c, trg := makeCore(1)
	ch := addVody(t, c, 0, 1)
	raw := ch.Character.(*char)
	raw.wandering = true
	c.Events.Emit(event.OnStellarVortexDetonate, 0, nil, nil)
	if trg[0].ResistModIsActive(anemoShredKey) == false {
		t.Fatal("expected anemo shred")
	}
	if raw.wandering {
		t.Fatal("wandering should clear on detonate")
	}
	if raw.alterUntil != c.F+a1AlterDur {
		t.Fatalf("alter until %d", raw.alterUntil)
	}
}

func TestRadianceExtendOnlyOnEnter(t *testing.T) {
	c, _ := makeCore(1)
	if reactable.RadianceSwirlDuration(c, false) != 8*60 {
		t.Fatal("base duration")
	}
	c.Flags.Custom[reactable.VodyanitsaSongUntilKey] = float64(c.F + 100)
	if reactable.RadianceSwirlDuration(c, false) != 12*60 {
		t.Fatal("enter extension")
	}
	if reactable.RadianceSwirlDuration(c, true) != 8*60 {
		t.Fatal("refresh should stay 8s")
	}
	c.Flags.Custom[reactable.VodyanitsaSongUntilKey] = float64(c.F - 1)
	if reactable.RadianceSwirlDuration(c, false) != 8*60 {
		t.Fatal("expired song flag must not extend radiance")
	}
}

func TestChargeAttackHasNoICD(t *testing.T) {
	c, _ := makeCore(1)
	addVody(t, c, 0, 1)
	var hit info.AttackInfo
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hit = args[1].(*info.AttackEvent).Info
	}, "ca")
	if err := c.Player.Exec(action.ActionCharge, keys.Vodyanitsa, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, chargeHitmark+2)
	if hit.ICDTag != attacks.ICDTagNone || hit.Durability != 25 || hit.Element != attributes.Hydro {
		t.Fatalf("charge %+v", hit)
	}
}

func TestC4HealIgnoresStackItGrants(t *testing.T) {
	c, _ := makeCore(1)
	ch := addVody(t, c, 4, 10)
	if err := c.Player.Exec(action.ActionSkill, keys.Vodyanitsa, nil); err != nil {
		t.Fatal(err)
	}
	hp := ch.MaxHP()
	var got float64
	c.Events.Subscribe(event.OnHeal, func(args ...any) {
		got = args[4].(float64)
	}, "heal")
	advance(c, healPreDelay+healInterval+2)
	want := healFlat[ch.TalentLvlSkill()] + healPct[ch.TalentLvlSkill()]*hp
	if math.Abs(got-want) > 1 {
		t.Fatalf("triggering heal %v want %v", got, want)
	}
	n, _ := ch.Condition([]string{"c4"})
	if n.(int) != 1 {
		t.Fatalf("c4 stacks %v", n)
	}
	if ch.MaxHP() <= hp*1.1 {
		t.Fatalf("max hp after stack %v before %v", ch.MaxHP(), hp)
	}
}

func TestC4StacksExpireIndependently(t *testing.T) {
	c, _ := makeCore(1)
	ch := addVody(t, c, 4, 10)
	raw := ch.Character.(*char)
	raw.addC4Stack()
	advance(c, 30)
	raw.addC4Stack()
	advance(c, 30)
	raw.addC4Stack()
	raw.addC4Stack()
	if n, _ := ch.Condition([]string{"c4"}); n.(int) != 3 {
		t.Fatalf("cap %v", n)
	}
	// First stack was added at frame 0 and lasts c4Dur. The later two are still inside their windows.
	advance(c, c4Dur-c.F+1)
	n, _ := ch.Condition([]string{"c4"})
	if n.(int) != 2 {
		t.Fatalf("after first expiry %v frame %d", n, c.F)
	}
}

func TestA4HydroFlatAndStellarStage(t *testing.T) {
	c, _ := makeCore(1)
	ch := addVody(t, c, 0, 10)
	raw := ch.Character.(*char)
	ch.BaseStats[attributes.HP] = 50000
	raw.lead = 5
	raw.chorus = 5
	ch.AddStatus(concertoKey, 300, false)

	ae := &info.AttackEvent{Info: info.AttackInfo{
		ActorIndex: ch.Index(),
		AttackTag:  attacks.AttackTagElementalArt,
		Element:    attributes.Hydro,
		FlatDmg:    100,
	}}
	c.Events.Emit(event.OnEnemyHit, nil, ae)
	want := 100 + raw.concertoFlat(false)
	if math.Abs(ae.Info.FlatDmg-want) > 1e-6 || raw.lead != 4 {
		t.Fatalf("hydro flat %v want %v lead %d", ae.Info.FlatDmg, want, raw.lead)
	}

	raw.wandering = true
	stellar := &info.AttackEvent{Info: info.AttackInfo{
		ActorIndex: ch.Index(),
		AttackTag:  attacks.AttackTagReactionStellarSwirl,
		Element:    attributes.Anemo,
	}}
	c.Events.Emit(event.OnSpecialReactionAttack, nil, stellar)
	if math.Abs(stellar.Info.FlatDmg-raw.concertoFlat(true)) > 1e-6 || raw.lead != 3 {
		t.Fatalf("stellar flat %v lead %d", stellar.Info.FlatDmg, raw.lead)
	}
	// The baked swirl attack must not consume a second stack.
	c.Events.Emit(event.OnEnemyHit, nil, stellar)
	if raw.lead != 3 {
		t.Fatalf("stellar carrier consumed again, lead %d", raw.lead)
	}
	// Hydro talent hits do not consume during the alter window.
	before := ae.Info.FlatDmg
	c.Events.Emit(event.OnEnemyHit, nil, ae)
	if raw.lead != 3 || ae.Info.FlatDmg != before {
		t.Fatalf("hydro consumed during alter lead %d flat %v", raw.lead, ae.Info.FlatDmg)
	}
}

func TestPlungeWhileAirborne(t *testing.T) {
	c, _ := makeCore(1)
	ch := addVody(t, c, 0, 1)
	var hits []info.AttackInfo
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hits = append(hits, args[1].(*info.AttackEvent).Info)
	}, "plunge")
	if err := c.Player.SetAirborne(player.AirborneXianyun); err != nil {
		t.Fatal(err)
	}
	if err := c.Player.Exec(action.ActionLowPlunge, keys.Vodyanitsa, map[string]int{"collision": 1}); err != nil {
		t.Fatal(err)
	}
	advance(c, lowPlungeHitmark+2)
	if len(hits) != 2 {
		t.Fatalf("plunge hits %d %+v", len(hits), hits)
	}
	if hits[0].Durability != 0 || hits[1].Durability != 25 || hits[1].Element != attributes.Hydro {
		t.Fatalf("plunge gauges %+v %+v", hits[0], hits[1])
	}
	if math.Abs(hits[1].Mult-lowPlunge[ch.TalentLvlAttack()]) > 1e-9 {
		t.Fatalf("low plunge mult %v", hits[1].Mult)
	}
}

func TestNoA4BelowAscension(t *testing.T) {
	c, _ := makeCore(1)
	prof := testhelper.DefaultProfile(keys.Vodyanitsa, testhelper.TestWeaponKey)
	prof.Base.Level = 20
	prof.Base.MaxLevel = 20
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(idx)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	ch := c.Player.ByIndex(idx)
	if err := c.Player.Exec(action.ActionSkill, keys.Vodyanitsa, nil); err != nil {
		t.Fatal(err)
	}
	lead, _ := ch.Condition([]string{"lead"})
	if lead.(int) != 0 {
		t.Fatalf("lead without a4 %v", lead)
	}
}
