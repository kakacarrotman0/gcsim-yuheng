package gaming

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
	gamingChargeTalent9      = 1.149091
	gamingChargeFinalTalent9 = 2.0777
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

func addGaming(t *testing.T, c *core.Core) {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Gaming, testhelper.TestWeaponKey)
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
	if err := c.Player.Exec(action.ActionCharge, keys.Gaming, p); err != nil {
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
	if h.Abil == specialPlungeKey {
		t.Fatalf("charged attack used Charmed Cloudstrider: %+v", h)
	}
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

func TestStellarRendCharge(t *testing.T) {
	c := makeCore(t)
	addGaming(t, c)
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
		if h.info.Abil == specialPlungeKey {
			t.Fatal("charged attack became Charmed Cloudstrider")
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
		assertClaymoreCharge(t, h.info, gamingChargeTalent9, 60)
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
	assertClaymoreCharge(t, last, gamingChargeFinalTalent9, 120)
}

func TestStellarRendChargeEndsAtFiveSeconds(t *testing.T) {
	c := makeCore(t)
	addGaming(t, c)
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

func TestCharmedCloudstriderStaysTheSkillPlunge(t *testing.T) {
	c := makeCore(t)
	addGaming(t, c)
	hits := watch(c)
	if err := c.Player.Exec(action.ActionSkill, keys.Gaming, nil); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c)
	doCharge(t, c, nil)
	untilQueue(t, c)
	for _, h := range *hits {
		if h.info.Abil == specialPlungeKey || h.info.AttackTag == attacks.AttackTagPlunge {
			t.Fatalf("skill then charge plunged: %+v", h.info)
		}
	}

	c = makeCore(t)
	addGaming(t, c)
	hits = watch(c)
	if err := c.Player.Exec(action.ActionSkill, keys.Gaming, nil); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c)
	if err := c.Player.Exec(action.ActionLowPlunge, keys.Gaming, nil); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c)
	advance(c, 40)
	saw := false
	for _, h := range *hits {
		if h.info.Abil == specialPlungeKey && h.info.AttackTag == attacks.AttackTagPlunge && h.info.Element == attributes.Pyro {
			saw = true
		}
	}
	if !saw {
		t.Fatal("skill plunge no longer releases Charmed Cloudstrider")
	}
}
