package navia

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

// Talent 9 rows KQM prints as 114.9% and 207.77%.
const (
	naviaChargeTalent9      = 1.149091
	naviaChargeFinalTalent9 = 2.0777
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

func addNavia(t *testing.T, c *core.Core) *char {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Navia, testhelper.TestWeaponKey)
	prof.Talents.Attack = 9
	if _, err := c.AddChar(prof); err != nil {
		t.Fatal(err)
	}
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	return c.Player.ByIndex(0).Character.(*char)
}

type hit struct {
	frame int
	info  info.AttackInfo
	dmgP  float64
}

func watch(c *core.Core) *[]hit {
	var hits []hit
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ev := args[1].(*info.AttackEvent)
		hits = append(hits, hit{frame: c.F, info: ev.Info})
	}, "hits")
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ev := args[1].(*info.AttackEvent)
		for i := range hits {
			if hits[i].info.Abil == ev.Info.Abil && hits[i].frame == ev.SourceFrame {
				hits[i].dmgP = ev.Snapshot.Stats[attributes.DmgP]
			}
		}
	}, "damage")
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
	if err := c.Player.Exec(action.ActionCharge, keys.Navia, p); err != nil {
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

func assertClaymoreCharge(t *testing.T, h info.AttackInfo, mult, poise float64, ele attributes.Element) {
	t.Helper()
	if h.Abil == "Rosula Shardshot" || h.Abil == "Surging Blade" {
		t.Fatalf("charged attack fired the skill: %+v", h)
	}
	if h.Mult != mult || h.PoiseDMG != poise || h.Durability != 25 {
		t.Fatalf("numbers %+v", h)
	}
	if h.AttackTag != attacks.AttackTagExtra || h.Element != ele || h.StrikeType != attacks.StrikeTypeBlunt {
		t.Fatalf("class %+v", h)
	}
	if h.ICDTag != attacks.ICDTagNormalAttack || h.ICDGroup != attacks.ICDGroupDefault {
		t.Fatalf("icd %+v", h)
	}
}

func TestBluntRefusalCharge(t *testing.T) {
	c := makeCore(t)
	nav := addNavia(t, c)
	nav.shrapnel = 4
	hits := watch(c)
	doCharge(t, c, nil)
	untilQueue(t, c)
	if math.Abs(c.Player.Stam-(player.MaxStam-12)) > 1e-6 {
		t.Fatalf("stamina %v", c.Player.Stam)
	}
	if nav.shrapnel != 4 {
		t.Fatalf("shrapnel %d", nav.shrapnel)
	}
	doCharge(t, c, nil)
	untilQueue(t, c)

	var cyclic []hit
	for _, h := range *hits {
		if h.info.Abil == "Rosula Shardshot" || h.info.Abil == "Surging Blade" {
			t.Fatalf("charged attack fired the skill: %s", h.info.Abil)
		}
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
		assertClaymoreCharge(t, h.info, naviaChargeTalent9, 60, attributes.Physical)
	}
	if cyclic[1].frame-cyclic[0].frame >= tmpl.ClaymoreSpinWindup {
		t.Fatalf("chained slash kept the windup: %d then %d", cyclic[0].frame, cyclic[1].frame)
	}
	if nav.shrapnel != 4 {
		t.Fatalf("shrapnel %d", nav.shrapnel)
	}

	doCharge(t, c, map[string]int{"final": 1})
	for range 80 {
		advance(c, 1)
		if len(*hits) > 2 && (*hits)[len(*hits)-1].info.Abil == "Charge Attack (Finisher)" {
			break
		}
	}
	last := (*hits)[len(*hits)-1].info
	if last.Abil != "Charge Attack (Finisher)" {
		t.Fatalf("finisher %+v", last)
	}
	assertClaymoreCharge(t, last, naviaChargeFinalTalent9, 120, attributes.Physical)
	if nav.shrapnel != 4 {
		t.Fatalf("finisher spent shrapnel %d", nav.shrapnel)
	}
}

func TestBluntRefusalChargeEndsAtFiveSeconds(t *testing.T) {
	c := makeCore(t)
	addNavia(t, c)
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

func TestCrystalshotInfusesTheFollowingCharge(t *testing.T) {
	c := makeCore(t)
	nav := addNavia(t, c)
	nav.shrapnel = 2
	hits := watch(c)
	if err := c.Player.Exec(action.ActionSkill, keys.Navia, nil); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c)
	advance(c, 2)
	if nav.shrapnel != 0 {
		t.Fatalf("skill left shrapnel %d", nav.shrapnel)
	}
	if !nav.StatusIsActive(a1Key) {
		t.Fatal("ceremonial crystalshot did not arm ascension 1")
	}
	sawGun := false
	for _, h := range *hits {
		if h.info.Abil == "Rosula Shardshot" {
			sawGun = true
		}
	}
	if !sawGun {
		t.Fatal("skill did not fire Rosula Shardshot")
	}
	for range 100 {
		advance(c, 1)
		seen := false
		for _, h := range *hits {
			if h.info.Abil == "Surging Blade" {
				seen = true
			}
		}
		if seen {
			break
		}
	}

	nav.shrapnel = 3
	before := len(*hits)
	doCharge(t, c, nil)
	untilQueue(t, c)
	advance(c, 2)
	if nav.shrapnel != 3 {
		t.Fatalf("infused charge spent shrapnel %d", nav.shrapnel)
	}
	var charged []hit
	for _, h := range (*hits)[before:] {
		if h.info.Abil == "Rosula Shardshot" || h.info.Abil == "Surging Blade" {
			t.Fatalf("charge fired the skill: %s", h.info.Abil)
		}
		if h.info.Abil == "Charge Attack" {
			charged = append(charged, h)
		}
	}
	if len(charged) != 1 {
		t.Fatalf("infused cyclic hits %d", len(charged))
	}
	assertClaymoreCharge(t, charged[0].info, naviaChargeTalent9, 60, attributes.Geo)
	if math.Abs(charged[0].dmgP-0.4) > 1e-9 {
		t.Fatalf("ascension 1 damage bonus %v", charged[0].dmgP)
	}
}
