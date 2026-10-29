package gestofthemightywolf

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/noelle"
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
