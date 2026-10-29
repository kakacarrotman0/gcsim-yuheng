package chongyun

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/internal/frames"
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestWeapon()
}

func makeCore(t *testing.T) (*core.Core, *enemy.Enemy) {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
	e := enemy.New(c, info.EnemyProfile{
		Level:  100,
		Resist: make(map[attributes.Element]float64),
		Pos:    info.Coord{X: 0, Y: 0, R: 1},
	})
	c.Combat.AddEnemy(e)
	c.Combat.DefaultTarget = e.Key()
	c.Player.SetActive(0)
	return c, e
}

func addChongyun(t *testing.T, c *core.Core, maxLevel, talent, cons int) {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Chongyun, testhelper.TestWeaponKey)
	prof.Base.Level = maxLevel
	prof.Base.MaxLevel = maxLevel
	prof.Base.Cons = cons
	prof.Talents = info.TalentProfile{Attack: talent, Skill: talent, Burst: talent}
	if _, err := c.AddChar(prof); err != nil {
		t.Fatal(err)
	}
	if err := c.Init(); err != nil {
		t.Fatal(err)
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

func advance(c *core.Core, n int) {
	for range n {
		c.F++
		_ = c.Tick()
	}
}

func doCharge(t *testing.T, c *core.Core, p map[string]int) {
	t.Helper()
	if err := c.Player.Exec(action.ActionCharge, keys.Chongyun, p); err != nil {
		t.Fatal(err)
	}
}

func untilQueue(t *testing.T, c *core.Core) {
	t.Helper()
	for range 800 {
		if c.Player.CanQueueNextAction() {
			return
		}
		advance(c, 1)
	}
	t.Fatal("action never opened the queue")
}

func TestDemonbaneCharge(t *testing.T) {
	c, _ := makeCore(t)
	addChongyun(t, c, 90, 9, 1)
	hits := watch(c)
	doCharge(t, c, nil)
	untilQueue(t, c)
	if math.Abs(c.Player.Stam-(player.MaxStam-12)) > 1e-6 {
		t.Fatalf("stamina %v", c.Player.Stam)
	}
	doCharge(t, c, nil)
	untilQueue(t, c)
	doCharge(t, c, map[string]int{"final": 1})
	for range 80 {
		advance(c, 1)
		if len(*hits) > 0 && (*hits)[len(*hits)-1].info.Abil == "Charge Attack (Finisher)" {
			break
		}
	}
	var cyclic, finisher, blades int
	for _, h := range *hits {
		switch h.info.Abil {
		case "Charge Attack":
			cyclic++
			if h.info.Mult != charge[8] || h.info.PoiseDMG != 60 || h.info.Element != attributes.Physical {
				t.Fatalf("cyclic %+v", h.info)
			}
			if h.info.AttackTag != attacks.AttackTagExtra || h.info.Durability != 25 {
				t.Fatalf("class %+v", h.info)
			}
		case "Charge Attack (Finisher)":
			finisher++
			if h.info.Mult != chargeFinal[8] || h.info.PoiseDMG != 120 || h.info.Element != attributes.Physical {
				t.Fatalf("finisher %+v", h.info)
			}
		case "Chongyun C1":
			blades++
		}
	}
	if cyclic != 2 || finisher != 1 || blades != 0 {
		t.Fatalf("cyclic %d finisher %d c1 %d", cyclic, finisher, blades)
	}
}

func TestChongyunC1StaysOnTheFourthNormal(t *testing.T) {
	c, _ := makeCore(t)
	addChongyun(t, c, 90, 9, 1)
	hits := watch(c)
	for range 4 {
		if err := c.Player.Exec(action.ActionAttack, keys.Chongyun, nil); err != nil {
			t.Fatal(err)
		}
		untilQueue(t, c)
	}
	advance(c, 15)
	blades := 0
	for _, h := range *hits {
		if h.info.Abil == "Chongyun C1" {
			blades++
			if h.info.Element != attributes.Cryo || h.info.Mult != 0.5 {
				t.Fatalf("blade %+v", h.info)
			}
		}
	}
	if blades != 3 {
		t.Fatalf("c1 blades %d", blades)
	}
}

func TestChonghuaInfusesChargeAndMelts(t *testing.T) {
	c, e := makeCore(t)
	addChongyun(t, c, 90, 9, 0)
	hits := watch(c)
	var reacted []info.AttackInfo
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		reacted = append(reacted, args[1].(*info.AttackEvent).Info)
	}, "react")
	if err := c.Player.Exec(action.ActionSkill, keys.Chongyun, nil); err != nil {
		t.Fatal(err)
	}
	for c.F <= skillHitmark {
		advance(c, 1)
	}
	pyro := &info.AttackEvent{Info: info.AttackInfo{Element: attributes.Pyro, Durability: 200}}
	e.React(pyro)
	if !pyro.Info.Amped {
		e.AttachOrRefill(pyro)
	} else {
		e.AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Pyro, Durability: pyro.Info.Durability}})
	}
	doCharge(t, c, nil)
	untilQueue(t, c)
	for _, h := range *hits {
		if h.info.Abil == "Charge Attack" && h.info.Element != attributes.Cryo {
			t.Fatalf("element %+v", h.info)
		}
	}
	var saw bool
	for _, ai := range reacted {
		if ai.Abil != "Charge Attack" {
			continue
		}
		saw = true
		if ai.Element != attributes.Cryo {
			t.Fatalf("element %+v", ai)
		}
		if !ai.Amped || ai.AmpType != info.ReactionTypeMelt || ai.AmpMult != 1.5 {
			t.Fatalf("melt %+v aura %v", ai, e.ActiveAuraString())
		}
	}
	if !saw {
		t.Fatal("no infused charged hit")
	}
}

func TestSteadyBreathingSpeedsChargeInsideTheField(t *testing.T) {
	plain := fieldChargeDelay(t, 20)
	buffed := fieldChargeDelay(t, 40)
	want := tmpl.ClaymoreSpinWindup + tmpl.ClaymoreSpinHit
	if plain != want || buffed != want {
		t.Fatalf("a0 delay %d a1 delay %d", plain, buffed)
	}
	plainNA := normalDashLock(t, 20)
	buffedNA := normalDashLock(t, 40)
	if plainNA != attackHitmarks[0] {
		t.Fatalf("a0 normal dash lock %d", plainNA)
	}
	wantNA := frames.AtkSpdAdjust(attackHitmarks[0], 0.08)
	if buffedNA != wantNA {
		t.Fatalf("a1 normal dash lock %d, want %d", buffedNA, wantNA)
	}
}

func normalDashLock(t *testing.T, maxLevel int) int {
	t.Helper()
	c, _ := makeCore(t)
	addChongyun(t, c, maxLevel, 9, 0)
	if err := c.Player.Exec(action.ActionSkill, keys.Chongyun, nil); err != nil {
		t.Fatal(err)
	}
	for c.F <= skillHitmark {
		advance(c, 1)
	}
	if err := c.Player.Exec(action.ActionAttack, keys.Chongyun, nil); err != nil {
		t.Fatal(err)
	}
	n := 0
	for c.Player.IsAnimationLocked(action.ActionDash) {
		advance(c, 1)
		n++
		if n > 80 {
			t.Fatal("normal never opened a dash cancel")
		}
	}
	return n
}

func fieldChargeDelay(t *testing.T, maxLevel int) int {
	t.Helper()
	c, _ := makeCore(t)
	addChongyun(t, c, maxLevel, 9, 0)
	asc := c.Player.ByIndex(0).Base.Ascension
	if maxLevel == 20 && asc != 0 {
		t.Fatalf("ascension %d", asc)
	}
	if maxLevel == 40 && asc != 1 {
		t.Fatalf("ascension %d", asc)
	}
	hits := watch(c)
	if err := c.Player.Exec(action.ActionSkill, keys.Chongyun, nil); err != nil {
		t.Fatal(err)
	}
	for c.F <= skillHitmark {
		advance(c, 1)
	}
	spd := c.Player.ByIndex(0).Stat(attributes.AtkSpd)
	if maxLevel == 20 && spd != 0 {
		t.Fatalf("a0 attack speed %v", spd)
	}
	if maxLevel == 40 && spd != 0.08 {
		t.Fatalf("a1 attack speed %v", spd)
	}
	start := c.F
	doCharge(t, c, nil)
	untilQueue(t, c)
	for _, h := range *hits {
		if h.info.Abil == "Charge Attack" {
			return h.frame - start
		}
	}
	t.Fatal("no charged hit")
	return 0
}

func TestDemonbaneChargeEndsAtFiveSeconds(t *testing.T) {
	c, _ := makeCore(t)
	addChongyun(t, c, 90, 9, 0)
	hits := watch(c)
	start := c.F
	doCharge(t, c, nil)
	var finisher int
	for c.F < start+tmpl.ClaymoreSpinDuration+180 {
		untilQueue(t, c)
		for _, h := range *hits {
			if h.info.Abil == "Charge Attack (Finisher)" {
				finisher = h.frame
			}
		}
		if finisher != 0 {
			break
		}
		doCharge(t, c, nil)
	}
	if finisher < start+tmpl.ClaymoreSpinDuration || finisher > start+tmpl.ClaymoreSpinDuration+90 {
		t.Fatalf("finisher frame %d start %d", finisher, start)
	}
}
