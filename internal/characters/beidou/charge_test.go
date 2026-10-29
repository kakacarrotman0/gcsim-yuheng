package beidou

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/internal/frames"
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func addBeidouCharge(t *testing.T, c *core.Core, maxLevel, talent int) {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Beidou, testhelper.TestWeaponKey)
	prof.Base.Level = maxLevel
	prof.Base.MaxLevel = maxLevel
	prof.Talents = info.TalentProfile{Attack: talent, Skill: talent, Burst: talent}
	prof.Stats[attributes.CR] = 1
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
	dmg   float64
}

func watch(c *core.Core) *[]hit {
	var hits []hit
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hits = append(hits, hit{frame: c.F, info: args[1].(*info.AttackEvent).Info})
	}, "hits")
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ae := args[1].(*info.AttackEvent)
		for i := range hits {
			if hits[i].frame == c.F && hits[i].info.Abil == ae.Info.Abil && hits[i].dmg == 0 {
				hits[i].dmg = args[2].(float64)
				break
			}
		}
	}, "dmg")
	return &hits
}

func doCharge(t *testing.T, c *core.Core, p map[string]int) {
	t.Helper()
	if err := c.Player.Exec(action.ActionCharge, keys.Beidou, p); err != nil {
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

func spinOpen(atkspd float64) int {
	return frames.AtkSpdAdjust(tmpl.ClaymoreSpinWindup, atkspd) + frames.AtkSpdAdjust(tmpl.ClaymoreSpinHit, atkspd)
}

func TestOceanborneCharge(t *testing.T) {
	c := makeCore(t)
	addBeidouCharge(t, c, 90, 9)
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
	var cyclic, finisher int
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
			if h.info.Mult != chargeFinal[8] || h.info.PoiseDMG != 120 {
				t.Fatalf("finisher %+v", h.info)
			}
		case "Stunning Revenge (C4)":
			t.Fatal("constellation 4 followed a charged attack")
		}
	}
	if cyclic != 2 || finisher != 1 {
		t.Fatalf("cyclic %d finisher %d", cyclic, finisher)
	}
}

func TestTidecallerA4SpeedsChargeWithoutShorteningWindup(t *testing.T) {
	plain := chargeDelay(t, 0)
	buffed := chargeDelay(t, 2)
	if plain != spinOpen(0) {
		t.Fatalf("plain delay %d", plain)
	}
	if buffed != spinOpen(0.15) {
		t.Fatalf("a4 delay %d, attack-speed clock %d", buffed, spinOpen(0.15))
	}
	if buffed >= plain {
		t.Fatalf("a4 did not shorten playback: %d vs %d", buffed, plain)
	}
}

func chargeDelay(t *testing.T, counter int) int {
	t.Helper()
	c := makeCore(t)
	addBeidouCharge(t, c, 90, 9)
	hits := watch(c)
	if err := c.Player.Exec(action.ActionSkill, keys.Beidou, map[string]int{"counter": counter}); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c)
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

func TestTidecallerA4BuffsChargeDamage(t *testing.T) {
	plain := chargeDamage(t, 0)
	buffed := chargeDamage(t, 2)
	if plain <= 0 || buffed <= 0 {
		t.Fatalf("damage plain %v buffed %v", plain, buffed)
	}
	if math.Abs(buffed/plain-1.15) > 1e-6 {
		t.Fatalf("a4 ratio %v (%v / %v)", buffed/plain, buffed, plain)
	}
}

func chargeDamage(t *testing.T, counter int) float64 {
	t.Helper()
	c := makeCore(t)
	addBeidouCharge(t, c, 90, 9)
	hits := watch(c)
	if err := c.Player.Exec(action.ActionSkill, keys.Beidou, map[string]int{"counter": counter}); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c)
	doCharge(t, c, nil)
	untilQueue(t, c)
	for _, h := range *hits {
		if h.info.Abil == "Charge Attack" {
			return h.dmg
		}
	}
	t.Fatal("no charged hit")
	return 0
}

func TestStormbreakerFollowsCharge(t *testing.T) {
	c := makeCore(t)
	addBeidouCharge(t, c, 90, 9)
	hits := watch(c)
	c.Player.ByIndex(0).Energy = c.Player.ByIndex(0).EnergyMax
	if err := c.Player.Exec(action.ActionBurst, keys.Beidou, nil); err != nil {
		t.Fatal(err)
	}
	untilQueue(t, c)
	doCharge(t, c, nil)
	untilQueue(t, c)
	advance(c, 3)
	saw := false
	for _, h := range *hits {
		if h.info.Abil == "Stormbreaker (Bounce)" {
			saw = true
			if h.info.Element != attributes.Electro || h.info.AttackTag != attacks.AttackTagElementalBurst {
				t.Fatalf("bounce %+v", h.info)
			}
		}
	}
	if !saw {
		t.Fatal("charged attack did not trigger Stormbreaker")
	}
}

func TestOceanborneChargeEndsAtFiveSeconds(t *testing.T) {
	c := makeCore(t)
	addBeidouCharge(t, c, 90, 9)
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
