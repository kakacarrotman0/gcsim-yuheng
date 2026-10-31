package gestofthemightywolf

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/dehya"
	_ "github.com/genshinsim/gcsim/internal/characters/lisa"
	_ "github.com/genshinsim/gcsim/internal/characters/noelle"
	"github.com/genshinsim/gcsim/internal/frames"
	_ "github.com/genshinsim/gcsim/internal/weapons/catalyst/magicguide"
	_ "github.com/genshinsim/gcsim/internal/weapons/catalyst/tulaytullahsremembrance"
	_ "github.com/genshinsim/gcsim/internal/weapons/claymore/waster"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func TestAttackSpeedAppliesOnlyDuringNormalAttacks(t *testing.T) {
	c := makeCore(t)
	ch := addNoelle(t, c)

	if got := ch.Stat(attributes.AtkSpd); math.Abs(got) > 1e-9 {
		t.Fatalf("idle attack speed %v", got)
	}

	if err := c.Player.Exec(action.ActionAttack, keys.Noelle, nil); err != nil {
		t.Fatal(err)
	}
	if c.Player.CurrentState() != action.NormalAttackState {
		t.Fatalf("state %v", c.Player.CurrentState())
	}
	if got := ch.Stat(attributes.AtkSpd); math.Abs(got-0.1) > 1e-9 {
		t.Fatalf("normal attack speed %v", got)
	}

	if err := c.Player.Exec(action.ActionSkill, keys.Noelle, nil); err != nil {
		t.Fatal(err)
	}
	if c.Player.CurrentState() != action.SkillState {
		t.Fatalf("state %v", c.Player.CurrentState())
	}
	if got := ch.Stat(attributes.AtkSpd); math.Abs(got) > 1e-9 {
		t.Fatalf("skill attack speed %v", got)
	}

	if err := c.Player.Exec(action.ActionCharge, keys.Noelle, nil); err != nil {
		t.Fatal(err)
	}
	if c.Player.CurrentState() != action.ChargeAttackState {
		t.Fatalf("state %v", c.Player.CurrentState())
	}
	if got := ch.Stat(attributes.AtkSpd); math.Abs(got) > 1e-9 {
		t.Fatalf("charged attack speed %v", got)
	}
}

func makeCore(t *testing.T) *core.Core {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1})
	if err != nil {
		t.Fatal(err)
	}
	a := avatar.New(c, info.Point{}, 1)
	c.Combat.SetPlayer(a)
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

// Sheet cancels measured with no attack speed. Noelle N1 and N2 are the
// attack-to-attack frames. A neutral claymore charge dash cancel is 56+23.
// Dehya's charge uses that same helper clock. Dehya N1 and N2 are 31 and 34.
const (
	noelleN1Attack    = 38
	noelleN2Attack    = 46
	neutralChargeDash = 79
	dehyaN1Attack     = 31
	dehyaN2Attack     = 34
	lisaN1Attack      = 30
	lisaN2Attack      = 20
	gestAtkSpd        = 0.1
)

type swingFrames struct {
	n1, n2, chargeAfterNormal, chargeFromIdle int
}

func TestGestFramesFollowNormalAttacksOnly(t *testing.T) {
	base := measureSwings(t, keys.Noelle, keys.WasterGreatsword)
	gest := measureSwings(t, keys.Noelle, keys.GestOfTheMightyWolf)
	if base != (swingFrames{noelleN1Attack, noelleN2Attack, neutralChargeDash, neutralChargeDash}) {
		t.Fatalf("waster noelle %+v", base)
	}
	want := swingFrames{
		n1:                frames.AtkSpdAdjust(noelleN1Attack, gestAtkSpd),
		n2:                frames.AtkSpdAdjust(noelleN2Attack, gestAtkSpd),
		chargeAfterNormal: neutralChargeDash,
		chargeFromIdle:    neutralChargeDash,
	}
	if gest != want {
		t.Fatalf("gest noelle %+v want %+v", gest, want)
	}

	base = measureSwings(t, keys.Dehya, keys.WasterGreatsword)
	gest = measureSwings(t, keys.Dehya, keys.GestOfTheMightyWolf)
	if base != (swingFrames{dehyaN1Attack, dehyaN2Attack, neutralChargeDash, neutralChargeDash}) {
		t.Fatalf("waster dehya %+v", base)
	}
	want = swingFrames{
		n1:                frames.AtkSpdAdjust(dehyaN1Attack, gestAtkSpd),
		n2:                frames.AtkSpdAdjust(dehyaN2Attack, gestAtkSpd),
		chargeAfterNormal: neutralChargeDash,
		chargeFromIdle:    neutralChargeDash,
	}
	if gest != want {
		t.Fatalf("gest dehya %+v want %+v", gest, want)
	}
}

// Tulaytullah's Remembrance uses the same NormalAttackState gate and does not
// record the action being built. Its first Normal Attack therefore stays on
// the unbuffed sheet. This test keeps that shared sample visible.
func TestTulaytullahStillSamplesBeforeTheStateUpdates(t *testing.T) {
	plain := measureSwings(t, keys.Lisa, keys.MagicGuide)
	tula := measureSwings(t, keys.Lisa, keys.TulaytullahsRemembrance)
	if plain.n1 != lisaN1Attack || plain.n2 != lisaN2Attack {
		t.Fatalf("magic guide %+v", plain)
	}
	if tula.n1 != lisaN1Attack {
		t.Fatalf("tulaytullah first normal %d", tula.n1)
	}
	if tula.n2 != frames.AtkSpdAdjust(lisaN2Attack, gestAtkSpd) {
		t.Fatalf("tulaytullah second normal %d", tula.n2)
	}
}

func measureSwings(t *testing.T, char keys.Char, weapon keys.Weapon) swingFrames {
	t.Helper()
	var got swingFrames

	c := makeCore(t)
	addCharWeapon(t, c, char, weapon)
	if err := c.Player.Exec(action.ActionAttack, char, nil); err != nil {
		t.Fatal(err)
	}
	got.n1 = ticksUntil(t, c, action.ActionAttack)

	c = makeCore(t)
	addCharWeapon(t, c, char, weapon)
	if err := c.Player.Exec(action.ActionAttack, char, nil); err != nil {
		t.Fatal(err)
	}
	_ = ticksUntil(t, c, action.ActionAttack)
	if err := c.Player.Exec(action.ActionAttack, char, nil); err != nil {
		t.Fatal(err)
	}
	got.n2 = ticksUntil(t, c, action.ActionAttack)

	c = makeCore(t)
	addCharWeapon(t, c, char, weapon)
	if err := c.Player.Exec(action.ActionAttack, char, nil); err != nil {
		t.Fatal(err)
	}
	_ = ticksUntil(t, c, action.ActionCharge)
	if err := c.Player.Exec(action.ActionCharge, char, nil); err != nil {
		t.Fatal(err)
	}
	got.chargeAfterNormal = ticksUntil(t, c, action.ActionDash)

	c = makeCore(t)
	addCharWeapon(t, c, char, weapon)
	if err := c.Player.Exec(action.ActionCharge, char, nil); err != nil {
		t.Fatal(err)
	}
	got.chargeFromIdle = ticksUntil(t, c, action.ActionDash)
	return got
}

func ticksUntil(t *testing.T, c *core.Core, next action.Action) int {
	t.Helper()
	for i := range 500 {
		if !c.Player.IsAnimationLocked(next) {
			return i
		}
		c.F++
		if err := c.Tick(); err != nil {
			t.Fatal(err)
		}
	}
	t.Fatal("action stayed locked")
	return -1
}

func addCharWeapon(t *testing.T, c *core.Core, char keys.Char, weapon keys.Weapon) {
	t.Helper()
	prof := testhelper.DefaultProfile(char, weapon)
	prof.Weapon.Refine = 1
	prof.Weapon.Level = 90
	prof.Weapon.MaxLevel = 90
	prof.Base.Cons = 0
	if _, err := c.AddChar(prof); err != nil {
		t.Fatal(err)
	}
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
}

func addNoelle(t *testing.T, c *core.Core) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Noelle, keys.GestOfTheMightyWolf)
	prof.Weapon.Refine = 1
	prof.Weapon.Level = 1
	prof.Weapon.MaxLevel = 20
	prof.Base.Cons = 0
	if _, err := c.AddChar(prof); err != nil {
		t.Fatal(err)
	}
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	return c.Player.ByIndex(0)
}
