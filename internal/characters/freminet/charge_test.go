package freminet

import (
	"math"
	"strings"
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
	freminetChargeTalent9      = 1.149091
	freminetChargeFinalTalent9 = 2.0777
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

func addFreminet(t *testing.T, c *core.Core) *char {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Freminet, testhelper.TestWeaponKey)
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
	if err := c.Player.Exec(action.ActionCharge, keys.Freminet, p); err != nil {
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

func assertClaymoreCharge(t *testing.T, h info.AttackInfo, mult, poise float64) {
	t.Helper()
	if h.Mult != mult || h.PoiseDMG != poise || h.Durability != 25 {
		t.Fatalf("numbers %+v", h)
	}
	if h.AttackTag != attacks.AttackTagExtra || h.Element != attributes.Physical || h.StrikeType != attacks.StrikeTypeBlunt {
		t.Fatalf("class %+v", h)
	}
	if h.ICDTag != attacks.ICDTagNormalAttack || h.ICDGroup != attacks.ICDGroupDefault {
		t.Fatalf("icd %+v", h)
	}
}

func TestFlowingEddiesCharge(t *testing.T) {
	c := makeCore(t)
	addFreminet(t, c)
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
		if strings.Contains(h.info.Abil, "Pers Time Frost") || strings.Contains(h.info.Abil, "Shattering Pressure") {
			t.Fatalf("normal-only skill hit during a charged attack: %s", h.info.Abil)
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
		assertClaymoreCharge(t, h.info, freminetChargeTalent9, 60)
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
	last := (*hits)[len(*hits)-1].info
	if last.Abil != "Charge Attack (Finisher)" {
		t.Fatalf("finisher %+v", last)
	}
	assertClaymoreCharge(t, last, freminetChargeFinalTalent9, 120)
}

func TestFlowingEddiesChargeEndsAtFiveSeconds(t *testing.T) {
	c := makeCore(t)
	addFreminet(t, c)
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

func TestPersTimerStaysOnNormalAttacks(t *testing.T) {
	c := makeCore(t)
	fr := addFreminet(t, c)
	hits := watch(c)
	if err := c.Player.Exec(action.ActionSkill, keys.Freminet, nil); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c)
	if !fr.StatusIsActive(persTimeKey) || fr.skillStacks != 0 {
		t.Fatalf("pers %v stacks %d", fr.StatusIsActive(persTimeKey), fr.skillStacks)
	}

	doCharge(t, c, nil)
	untilQueue(t, c)
	if fr.skillStacks != 0 || !fr.StatusIsActive(persTimeKey) {
		t.Fatalf("charge changed pers stacks %d active %v", fr.skillStacks, fr.StatusIsActive(persTimeKey))
	}
	for _, h := range *hits {
		if strings.Contains(h.info.Abil, "Pers Time Frost") || strings.Contains(h.info.Abil, "Shattering Pressure") {
			t.Fatalf("charged attack triggered %s", h.info.Abil)
		}
	}

	if err := c.Player.Exec(action.ActionAttack, keys.Freminet, nil); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c)
	advance(c, 20)
	if fr.skillStacks != 1 {
		t.Fatalf("normal stacks %d", fr.skillStacks)
	}
	sawFrost := false
	for _, h := range *hits {
		if h.info.Abil == "Pressurized Floe: Pers Time Frost" && h.info.Element == attributes.Cryo {
			sawFrost = true
		}
	}
	if !sawFrost {
		t.Fatal("normal attack did not release pers frost")
	}

	fr.skillStacks = 4
	before := len(*hits)
	doCharge(t, c, map[string]int{"final": 1})
	for range 80 {
		advance(c, 1)
	}
	if fr.skillStacks != 4 || !fr.StatusIsActive(persTimeKey) {
		t.Fatalf("finisher detonated pers stacks %d active %v", fr.skillStacks, fr.StatusIsActive(persTimeKey))
	}
	for _, h := range (*hits)[before:] {
		if strings.Contains(h.info.Abil, "Shattering Pressure") || strings.Contains(h.info.Abil, "Pers Time Frost") {
			t.Fatalf("finisher triggered %s", h.info.Abil)
		}
	}

	if err := c.Player.Exec(action.ActionAttack, keys.Freminet, nil); err != nil {
		t.Fatal(err)
	}
	if fr.skillStacks != 0 || fr.StatusIsActive(persTimeKey) {
		t.Fatalf("normal did not detonate stacks %d active %v", fr.skillStacks, fr.StatusIsActive(persTimeKey))
	}
	untilQueue(t, c)
	advance(c, 50)
	sawPressure := false
	for _, h := range *hits {
		if strings.Contains(h.info.Abil, "Shattering Pressure") {
			sawPressure = true
			if h.info.Element != attributes.Cryo && h.info.Element != attributes.Physical {
				t.Fatalf("pressure element %+v", h.info)
			}
		}
	}
	if !sawPressure {
		t.Fatal("level 4 normal did not release shattering pressure")
	}
}
