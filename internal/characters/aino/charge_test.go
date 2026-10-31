package aino

import (
	"math"
	"os"
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

// Talent 9. KQM rounds the cyclic hit to 114.9% and the frame column to 114.87%.
// ProudSkill group 12131 param 3 is 1.14866. The finisher, param 4, is 2.0777.
// Aino's normal has three hits, so these sit one index earlier than a four-hit claymore.
const (
	ainoChargeTalent9      = 1.14866
	ainoChargeFinalTalent9 = 2.0777
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

func addAino(t *testing.T, c *core.Core) {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Aino, testhelper.TestWeaponKey)
	prof.Talents.Attack = 9
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
	if err := c.Player.Exec(action.ActionCharge, keys.Aino, p); err != nil {
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

func TestBishBashBoshCharge(t *testing.T) {
	c := makeCore(t)
	addAino(t, c)
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
		if h.info.Mult != ainoChargeTalent9 || h.info.PoiseDMG != 60 || h.info.Durability != 25 {
			t.Fatalf("cyclic %+v", h.info)
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
	last := (*hits)[len(*hits)-1].info
	if last.Abil != "Charge Attack (Finisher)" || last.Mult != ainoChargeFinalTalent9 || last.PoiseDMG != 120 {
		t.Fatalf("finisher %+v", last)
	}
	if last.Element != attributes.Physical || last.StrikeType != attacks.StrikeTypeBlunt || last.Durability != 25 {
		t.Fatalf("finisher class %+v", last)
	}
}

func TestBishBashBoshChargeEndsAtFiveSeconds(t *testing.T) {
	c := makeCore(t)
	addAino(t, c)
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

func TestCyclicProudSkillSeries(t *testing.T) {
	want := []float64{
		0.62522, 0.67611, 0.727, 0.7997, 0.85059, 0.90875, 0.98872,
		1.06869, 1.14866, 1.2359, 1.32314, 1.41038, 1.49762, 1.58486, 1.6721,
	}
	if len(charge) != len(want) {
		t.Fatalf("len %d", len(charge))
	}
	for i := range want {
		if charge[i] != want[i] {
			t.Fatalf("talent %d got %v want %v", i+1, charge[i], want[i])
		}
	}
}

func TestHoldSkillStaysDocumented(t *testing.T) {
	raw, err := os.ReadFile("config.yml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "Hold skill is not implemented.") {
		t.Fatal("config.yml no longer records the unimplemented hold skill")
	}
}
