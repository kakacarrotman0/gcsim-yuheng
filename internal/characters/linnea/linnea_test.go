package linnea

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/noelle"
	_ "github.com/genshinsim/gcsim/internal/characters/sucrose"
	_ "github.com/genshinsim/gcsim/internal/characters/xingqiu"
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

func TestIdentity(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Linnea, 0, 1)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	if ch.Base.Element != attributes.Geo || ch.Weapon.Class != info.WeaponClassBow || ch.EnergyMax != 60 {
		t.Fatalf("identity ele %v weapon %v energy %v", ch.Base.Element, ch.Weapon.Class, ch.EnergyMax)
	}
	if ch.Moonsign != 1 {
		t.Fatalf("moonsign %v", ch.Moonsign)
	}
	if math.Abs(ch.BaseStats[attributes.CR]-(0.05+0.192)) > 1e-4 {
		t.Fatalf("cr %v", ch.BaseStats[attributes.CR])
	}
	if ch.TalentLvlSkill() != 0 || ch.TalentLvlAttack() != 0 {
		t.Fatalf("talent skill %d attack %d", ch.TalentLvlSkill(), ch.TalentLvlAttack())
	}
}

func TestTalentCons(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Linnea, 6, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	// C3 is the skill con and C5 is the burst con. Normal stays at talent 10.
	if ch.TalentLvlAttack() != 9 || ch.TalentLvlSkill() != 12 || ch.TalentLvlBurst() != 12 {
		t.Fatalf("talents na %d skill %d burst %d", ch.TalentLvlAttack(), ch.TalentLvlSkill(), ch.TalentLvlBurst())
	}
}

func TestBurstHeal(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Linnea, 0, 10)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	ch.SetHPByRatio(0.5)
	var amounts []float64
	c.Events.Subscribe(event.OnHeal, func(args ...any) {
		amounts = append(amounts, args[4].(float64))
	}, "heal")
	ch.Energy = ch.EnergyMax
	if err := c.Player.Exec(action.ActionBurst, keys.Linnea, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 12*60+2)
	if len(amounts) != 13 {
		t.Fatalf("heals %d, want initial + 12 ticks", len(amounts))
	}
	lvl := ch.TalentLvlBurst()
	def := ch.TotalDef(false)
	wantInitial := burstInitialDef[lvl] + burstInitialFlat[lvl]*def
	wantTick := burstTickDef[lvl] + burstTickFlat[lvl]*def
	if math.Abs(amounts[0]-wantInitial) > 1e-3 {
		t.Fatalf("initial %v want %v", amounts[0], wantInitial)
	}
	for i := 1; i < len(amounts); i++ {
		if math.Abs(amounts[i]-wantTick) > 1e-3 {
			t.Fatalf("tick %d %v want %v", i, amounts[i], wantTick)
		}
	}
}

func TestSkillParticles(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Linnea, 0, 1)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	var particles []character.Particle
	c.Events.Subscribe(event.OnParticleReceived, func(args ...any) {
		particles = append(particles, args[0].(character.Particle))
	}, "p")
	if err := c.Player.Exec(action.ActionSkill, keys.Linnea, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 36+122+ch.ParticleDelay+5)
	if len(particles) != 1 || particles[0].Num != particleCount || particles[0].Ele != attributes.Geo {
		t.Fatalf("particles %+v", particles)
	}
}

func TestC2DoesNotStopAtOtherElements(t *testing.T) {
	c, trg := makeCore(1)
	linnea := addProfile(t, c, keys.Linnea, 2, 1)
	noelle := addProfile(t, c, keys.Noelle, 0, 1)
	sucrose := addProfile(t, c, keys.Sucrose, 0, 1)
	xingqiu := addProfile(t, c, keys.Xingqiu, 0, 1)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(linnea.Index())
	c.Events.Emit(event.OnMoondriftHarmony, trg[0])
	if math.Abs(noelle.Stat(attributes.CD)-0.9) > 1e-9 {
		t.Fatalf("noelle cd %v", noelle.Stat(attributes.CD))
	}
	if math.Abs(sucrose.Stat(attributes.CD)-0.5) > 1e-9 {
		t.Fatalf("sucrose cd %v", sucrose.Stat(attributes.CD))
	}
	if math.Abs(xingqiu.Stat(attributes.CD)-0.9) > 1e-9 {
		t.Fatalf("xingqiu cd %v", xingqiu.Stat(attributes.CD))
	}
}
