package disenchantmentindeepshadow

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/kaeya"
	_ "github.com/genshinsim/gcsim/internal/characters/noelle"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/reactable"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestWeapon()
}

func TestCritRateBonusIsOnlyTheWearersSpecialReaction(t *testing.T) {
	cases := []struct {
		name string
		key  string
	}{
		{name: "superconduct", key: reactable.SuperConductShredKey},
		{name: "stellar conduct", key: reactable.StellarConductShredKey},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, trg := makeCore(t)
			trg.AddStatus(tc.key, 60, false)

			wearer := specialCR(c, trg, 0)
			other := specialCR(c, trg, 1)
			if math.Abs(wearer-0.16) > 1e-9 {
				t.Fatalf("wearer crit rate %v", wearer)
			}
			if math.Abs(other) > 1e-9 {
				t.Fatalf("other character crit rate %v", other)
			}
		})
	}

	t.Run("no shred", func(t *testing.T) {
		c, trg := makeCore(t)
		if got := specialCR(c, trg, 0); math.Abs(got) > 1e-9 {
			t.Fatalf("wearer crit rate without shred %v", got)
		}
	})
}

func specialCR(c *core.Core, trg *enemy.Enemy, actor int) float64 {
	ae := &info.AttackEvent{Info: info.AttackInfo{ActorIndex: actor}}
	c.Events.Emit(event.OnSpecialReactionAttack, trg, ae)
	return ae.Snapshot.Stats[attributes.CR]
}

func makeCore(t *testing.T) (*core.Core, *enemy.Enemy) {
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

	wearer := testhelper.DefaultProfile(keys.Noelle, testhelper.TestWeaponKey)
	wearer.Sets[keys.DisenchantmentInDeepShadow] = 4
	other := testhelper.DefaultProfile(keys.Kaeya, testhelper.TestWeaponKey)
	if _, err := c.AddChar(wearer); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AddChar(other); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	return c, e
}
