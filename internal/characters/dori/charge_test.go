package dori

import (
	"math"
	"testing"

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

func makeCore(t *testing.T) *core.Core {
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
	return c
}

func addDori(t *testing.T, c *core.Core, maxLevel, talent, cons int) {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Dori, testhelper.TestWeaponKey)
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
	frame  int
	info   info.AttackInfo
	radius float64
}

func watch(c *core.Core) *[]hit {
	var hits []hit
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ev := args[1].(*info.AttackEvent)
		r := -1.0
		if circ, ok := ev.Pattern.Shape.(*info.Circle); ok {
			r = circ.Radius()
		}
		hits = append(hits, hit{frame: c.F, info: ev.Info, radius: r})
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
	if err := c.Player.Exec(action.ActionCharge, keys.Dori, p); err != nil {
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

func TestLampCharge(t *testing.T) {
	c := makeCore(t)
	addDori(t, c, 90, 9, 0)
	hits := watch(c)
	doCharge(t, c, nil)
	untilQueue(t, c)
	if math.Abs(c.Player.Stam-(player.MaxStam-12)) > 1e-6 {
		t.Fatalf("stamina %v", c.Player.Stam)
	}
	doCharge(t, c, nil)
	untilQueue(t, c)

	var cyclic []hit
	for _, h := range *hits {
		if h.info.Abil == "Charge Attack" {
			cyclic = append(cyclic, h)
		}
		if h.info.Abil == "Charge Attack (Finisher)" {
			t.Fatal("finisher landed before release")
		}
	}
	if len(cyclic) != 2 {
		t.Fatalf("cyclic hits %d", len(cyclic))
	}
	for _, h := range cyclic {
		if h.info.Mult != charge[8] || h.info.PoiseDMG != 60 || h.info.Durability != 25 || h.radius != 2.8 {
			t.Fatalf("cyclic %+v radius %v", h.info, h.radius)
		}
		if h.info.AttackTag != attacks.AttackTagExtra || h.info.Element != attributes.Physical || h.info.StrikeType != attacks.StrikeTypeBlunt {
			t.Fatalf("class %+v", h.info)
		}
		if h.info.ICDTag != attacks.ICDTagNormalAttack || h.info.ICDGroup != attacks.ICDGroupDefault {
			t.Fatalf("icd %+v", h.info)
		}
	}
	if cyclic[1].frame-cyclic[0].frame >= tmpl.ClaymoreSpinWindup {
		t.Fatalf("chained slash kept the windup: %d then %d", cyclic[0].frame, cyclic[1].frame)
	}

	doCharge(t, c, map[string]int{"final": 1})
	for range 80 {
		advance(c, 1)
		if len(*hits) > 2 && (*hits)[len(*hits)-1].info.Abil == "Charge Attack (Finisher)" {
			break
		}
	}
	last := (*hits)[len(*hits)-1]
	if last.info.Abil != "Charge Attack (Finisher)" || last.info.Mult != chargeFinal[8] || last.info.PoiseDMG != 120 || last.radius != 3.3 {
		t.Fatalf("finisher %+v radius %v", last.info, last.radius)
	}
}

func TestLampChargeEndsAtFiveSeconds(t *testing.T) {
	c := makeCore(t)
	addDori(t, c, 90, 9, 0)
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

func TestTroubleshooterInfusionSkipsTheHeal(t *testing.T) {
	c := makeCore(t)
	addDori(t, c, 90, 9, 6)
	hits := watch(c)
	heals := 0
	c.Events.Subscribe(event.OnHeal, func(args ...any) {
		hi := args[0].(*info.HealInfo)
		if hi.Message == "dori-c6-heal" {
			heals++
		}
	}, "c6-heal")
	if err := c.Player.Exec(action.ActionSkill, keys.Dori, nil); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c)
	if err := c.Player.Exec(action.ActionAttack, keys.Dori, nil); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c)
	advance(c, 5)
	if heals != 1 {
		t.Fatalf("normal heal count %d", heals)
	}
	start := c.F
	doCharge(t, c, nil)
	untilQueue(t, c)
	var charged *hit
	for i := range *hits {
		h := &(*hits)[i]
		if h.info.Abil == "Charge Attack" && h.frame >= start {
			charged = h
		}
	}
	if charged == nil || charged.info.Element != attributes.Electro {
		t.Fatalf("infused charge %+v", charged)
	}
	if heals != 1 {
		t.Fatalf("charged attack healed, count %d", heals)
	}
}
