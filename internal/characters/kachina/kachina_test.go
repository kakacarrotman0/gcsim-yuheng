package kachina

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/bennett"
	"github.com/genshinsim/gcsim/internal/template/crystallize"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/construct"
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

func addKachina(t *testing.T, c *core.Core, cons, talent int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Kachina, testhelper.TestWeaponKey)
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
	ch := addKachina(t, c, 0, 1)
	if ch.Base.Element != attributes.Geo {
		t.Fatalf("element %v", ch.Base.Element)
	}
	if ch.Weapon.Class != info.WeaponClassSpear {
		t.Fatalf("weapon %v", ch.Weapon.Class)
	}
	if ch.EnergyMax != 70 {
		t.Fatalf("energy %v", ch.EnergyMax)
	}
	if math.Abs(ch.BaseStats[attributes.GeoP]-0.24) > 1e-9 {
		t.Fatalf("geo%% %v", ch.BaseStats[attributes.GeoP])
	}
	if ch.BaseStats[attributes.BaseDEF] < 500 {
		t.Fatalf("base def %v", ch.BaseStats[attributes.BaseDEF])
	}
}

func TestNormalAttack(t *testing.T) {
	c, trg := makeCore(1)
	ch := addKachina(t, c, 0, 1)
	var hits []info.AttackInfo
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hits = append(hits, args[1].(*info.AttackEvent).Info)
	}, "na")
	if err := c.Player.Exec(action.ActionAttack, keys.Kachina, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, attackHitmark[0][0]+2)
	if len(hits) != 1 {
		t.Fatalf("hits %d", len(hits))
	}
	if hits[0].Abil != "Normal 0" || hits[0].Element != attributes.Physical || hits[0].Durability != 25 {
		t.Fatalf("hit %+v", hits[0])
	}
	if math.Abs(hits[0].Mult-attack[0][0][ch.TalentLvlAttack()]) > 1e-9 {
		t.Fatalf("mult %v", hits[0].Mult)
	}
	if trg[0].AuraContains(attributes.Geo) {
		t.Fatal("normal should not apply geo")
	}
}

func TestSkillOffFieldSlams(t *testing.T) {
	c, trg := makeCore(1)
	ch := addKachina(t, c, 0, 1)
	var hits []info.AttackInfo
	var crystallized int
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hits = append(hits, args[1].(*info.AttackEvent).Info)
	}, "e")
	c.Events.Subscribe(event.OnCrystallizeHydro, func(_ ...any) {
		crystallized++
	}, "cr")
	seed := &info.AttackEvent{
		Info: info.AttackInfo{
			ActorIndex: ch.Index(),
			Abil:       "Hydro Seed",
			Element:    attributes.Hydro,
			Durability: 25,
			ICDTag:     attacks.ICDTagNone,
		},
	}
	trg[0].HandleAttack(seed)
	advance(c, 1)
	if !trg[0].AuraContains(attributes.Hydro) {
		t.Fatal("seed hydro missing")
	}
	hits = nil
	if err := c.Player.Exec(action.ActionSkill, keys.Kachina, nil); err != nil {
		t.Fatal(err)
	}
	if !ch.StatusIsActive(twirlyKey) {
		t.Fatal("twirly missing")
	}
	state, err := ch.Condition([]string{"nightsoul", "state"})
	if err != nil || state != true {
		t.Fatalf("blessing %v %v", state, err)
	}
	points, _ := ch.Condition([]string{"nightsoul", "points"})
	if points.(float64) != 60 {
		t.Fatalf("points %v", points)
	}
	mounted, _ := ch.Condition([]string{"mounted"})
	if mounted.(bool) {
		t.Fatal("tap should not mount")
	}
	if c.Constructs.CountByType(construct.GeoConstructKachinaSkill) != 1 {
		t.Fatalf("constructs %d", c.Constructs.CountByType(construct.GeoConstructKachinaSkill))
	}

	advance(c, firstSlam-1)
	if len(hits) != 0 {
		t.Fatalf("early hits %d", len(hits))
	}
	advance(c, 2)
	if len(hits) != 1 {
		t.Fatalf("first slam hits %d", len(hits))
	}
	hit := hits[0]
	if hit.Abil != "Turbo Twirly" || hit.Element != attributes.Geo || hit.Durability != 25 || !hit.UseDef {
		t.Fatalf("slam %+v", hit)
	}
	if hit.ICDTag != attacks.ICDTagElementalArt || hit.ICDGroup != attacks.ICDGroupDefault {
		t.Fatalf("icd %v %v", hit.ICDTag, hit.ICDGroup)
	}
	if len(hit.AdditionalTags) != 1 || hit.AdditionalTags[0] != attacks.AttackTagNightsoul {
		t.Fatalf("tags %v", hit.AdditionalTags)
	}
	if math.Abs(hit.Mult-independent[ch.TalentLvlSkill()]) > 1e-9 {
		t.Fatalf("mult %v", hit.Mult)
	}
	if math.Abs(hit.FlatDmg-ch.TotalDef(false)*a4Ratio) > 1e-6 {
		t.Fatalf("a4 flat %v def %v", hit.FlatDmg, ch.TotalDef(false))
	}
	points, _ = ch.Condition([]string{"nightsoul", "points"})
	if points.(float64) != 50 {
		t.Fatalf("points after slam %v", points)
	}
	if crystallized != 1 {
		t.Fatalf("crystallize %d", crystallized)
	}

	advance(c, slamInterval)
	if len(hits) != 2 {
		t.Fatalf("second slam hits %d", len(hits))
	}
}

func TestSkillEndsWhenPointsDeplete(t *testing.T) {
	c, _ := makeCore(1)
	ch := addKachina(t, c, 0, 1)
	var slams int
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		if args[1].(*info.AttackEvent).Info.Abil == "Turbo Twirly" {
			slams++
		}
	}, "e")
	if err := c.Player.Exec(action.ActionSkill, keys.Kachina, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, firstSlam+slamInterval*5+5)
	if slams != 6 {
		t.Fatalf("slams %d", slams)
	}
	if ch.StatusIsActive(twirlyKey) {
		t.Fatal("twirly should end")
	}
	if c.Constructs.CountByType(construct.GeoConstructKachinaSkill) != 0 {
		t.Fatal("construct should be removed when nightsoul ends")
	}
	state, _ := ch.Condition([]string{"nightsoul", "state"})
	if state.(bool) {
		t.Fatal("blessing should end")
	}
	advance(c, slamInterval)
	if slams != 6 {
		t.Fatalf("extra slam %d", slams)
	}
}

func TestParticles(t *testing.T) {
	c, _ := makeCore(1)
	ch := addKachina(t, c, 0, 1)
	var particles []character.Particle
	c.Events.Subscribe(event.OnParticleReceived, func(args ...any) {
		particles = append(particles, args[0].(character.Particle))
	}, "p")
	if err := c.Player.Exec(action.ActionSkill, keys.Kachina, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, firstSlam+ch.ParticleDelay+2)
	if len(particles) != 1 || particles[0].Num != 1 || particles[0].Ele != attributes.Geo {
		t.Fatalf("particles %+v", particles)
	}
}

func TestBurstAndC2(t *testing.T) {
	c, _ := makeCore(1)
	ch := addKachina(t, c, 0, 1)
	var hits []info.AttackInfo
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hits = append(hits, args[1].(*info.AttackEvent).Info)
	}, "q")
	if err := c.Player.Exec(action.ActionBurst, keys.Kachina, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, burstHitmark+2)
	if len(hits) != 1 || hits[0].Abil != "Time to Get Serious!" || !hits[0].UseDef || hits[0].ICDTag != attacks.ICDTagNone {
		t.Fatalf("burst %+v", hits)
	}
	if math.Abs(hits[0].Mult-burst[ch.TalentLvlBurst()]) > 1e-9 {
		t.Fatalf("burst mult %v", hits[0].Mult)
	}
	if !ch.StatusIsActive(fieldKey) {
		t.Fatal("field missing")
	}
	if ch.StatusIsActive(twirlyKey) {
		t.Fatal("c0 burst should not summon twirly")
	}
	if c.Constructs.CountByType(construct.GeoConstructKachinaSkill) != 0 {
		t.Fatal("c0 burst should not leave a construct")
	}

	c2, _ := makeCore(1)
	ch2 := addKachina(t, c2, 2, 1)
	if err := c2.Player.Exec(action.ActionBurst, keys.Kachina, nil); err != nil {
		t.Fatal(err)
	}
	advance(c2, burstHitmark+2)
	points, _ := ch2.Condition([]string{"nightsoul", "points"})
	if points.(float64) != 20 || !ch2.StatusIsActive(twirlyKey) {
		t.Fatalf("c2 points %v status %v", points, ch2.StatusIsActive(twirlyKey))
	}
	if c2.Constructs.CountByType(construct.GeoConstructKachinaSkill) != 1 {
		t.Fatal("c2 should summon one construct")
	}
	var slams int
	c2.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		if args[1].(*info.AttackEvent).Info.Abil == "Turbo Twirly" {
			slams++
		}
	}, "c2")
	advance(c2, firstSlam)
	if slams != 1 {
		t.Fatalf("c2 slam %d", slams)
	}
	points, _ = ch2.Condition([]string{"nightsoul", "points"})
	if points.(float64) != 10 {
		t.Fatalf("c2 points after slam %v", points)
	}
}

func TestC1CrystallizePickup(t *testing.T) {
	c, _ := makeCore(1)
	ch := addKachina(t, c, 1, 1)
	ch.Energy = 0
	c.Player.Shields.Add(crystallize.NewShield(ch.Index(), attributes.Cryo, 1, 90, 0, c.F+600))
	if math.Abs(ch.Energy-c1Energy) > 1e-9 {
		t.Fatalf("energy %v", ch.Energy)
	}
	c.Player.Shields.Add(crystallize.NewShield(ch.Index(), attributes.Hydro, 2, 90, 0, c.F+600))
	if math.Abs(ch.Energy-c1Energy) > 1e-9 {
		t.Fatalf("icd energy %v", ch.Energy)
	}
	advance(c, 5*60+1)
	c.Player.Shields.Add(crystallize.NewShield(ch.Index(), attributes.Pyro, 3, 90, 0, c.F+600))
	if math.Abs(ch.Energy-c1Energy*2) > 1e-9 {
		t.Fatalf("second pickup %v", ch.Energy)
	}
}

func TestC4ActiveDEF(t *testing.T) {
	c, _ := makeCore(1)
	ch := addKachina(t, c, 4, 1)
	if err := c.Player.Exec(action.ActionBurst, keys.Kachina, nil); err != nil {
		t.Fatal(err)
	}
	if ch.Stat(attributes.DEFP) != 0 {
		t.Fatalf("def before hit %v", ch.Stat(attributes.DEFP))
	}
	advance(c, burstHitmark+1)
	if math.Abs(ch.Stat(attributes.DEFP)-c4Ratio[0]) > 1e-9 {
		t.Fatalf("def %v", ch.Stat(attributes.DEFP))
	}
}

func TestC6ShieldBreak(t *testing.T) {
	c, _ := makeCore(1)
	ch := addKachina(t, c, 6, 1)
	var hits int
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		if args[1].(*info.AttackEvent).Info.Abil == "This Time, I've Gotta Win" {
			hits++
		}
	}, "c6")
	sh := crystallize.NewShield(ch.Index(), attributes.Geo, 1, 90, 0, c.F+30)
	c.Player.Shields.Add(sh)
	advance(c, 2)
	if hits != 0 {
		t.Fatalf("first shield should not trigger, hits %d", hits)
	}
	c.Player.Shields.Add(crystallize.NewShield(ch.Index(), attributes.Geo, 2, 90, 0, c.F+30))
	advance(c, 2)
	if hits != 1 {
		t.Fatalf("replace hits %d", hits)
	}
	advance(c, 5*60)
	c.Player.Shields.Add(crystallize.NewShield(ch.Index(), attributes.Electro, 3, 90, 0, c.F+5))
	advance(c, 8)
	if hits != 2 {
		t.Fatalf("break hits %d", hits)
	}
}

func TestA1OwnGeoBonus(t *testing.T) {
	c, _ := makeCore(1)
	prof := testhelper.DefaultProfile(keys.Kachina, testhelper.TestWeaponKey)
	kidx, err := c.AddChar(prof)
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
	ch := c.Player.ByIndex(kidx)
	ben := c.Player.ByIndex(bidx)
	before := ch.Stat(attributes.GeoP)
	c.Events.Emit(event.OnNightsoulBurst, nil, nil)
	if math.Abs(ch.Stat(attributes.GeoP)-before-a1Ratio) > 1e-9 {
		t.Fatalf("kachina geo %v before %v", ch.Stat(attributes.GeoP), before)
	}
	if ben.Stat(attributes.GeoP) != 0 {
		t.Fatalf("bennett geo %v", ben.Stat(attributes.GeoP))
	}
}

func TestTalentScaling(t *testing.T) {
	c, _ := makeCore(1)
	ch := addKachina(t, c, 0, 10)
	if ch.TalentLvlSkill() != 9 {
		t.Fatalf("skill index %d", ch.TalentLvlSkill())
	}
	var mult float64
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		mult = args[1].(*info.AttackEvent).Info.Mult
	}, "scale")
	if err := c.Player.Exec(action.ActionSkill, keys.Kachina, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, firstSlam+2)
	if math.Abs(mult-independent[9]) > 1e-9 {
		t.Fatalf("mult %v", mult)
	}
}

func TestOffFieldContinues(t *testing.T) {
	c, _ := makeCore(1)
	kprof := testhelper.DefaultProfile(keys.Kachina, testhelper.TestWeaponKey)
	kidx, err := c.AddChar(kprof)
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
	c.Player.SetActive(kidx)
	var slams int
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		if args[1].(*info.AttackEvent).Info.Abil == "Turbo Twirly" {
			slams++
		}
	}, "off")
	if err := c.Player.Exec(action.ActionSkill, keys.Kachina, nil); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(bidx)
	advance(c, firstSlam+slamInterval+2)
	if slams != 2 {
		t.Fatalf("off-field slams %d", slams)
	}
}
