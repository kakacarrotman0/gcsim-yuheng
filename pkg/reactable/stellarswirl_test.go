package reactable

import (
	"testing"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/target"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func stellarCore(t *testing.T) (*core.Core, *testTarget) {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1, Debug: false})
	if err != nil {
		t.Fatal(err)
	}
	player := &testTarget{}
	player.Target = target.New(c, info.Point{X: 0, Y: 0}, 1)
	player.Reactable = &Reactable{}
	player.typ = info.TargettablePlayer
	player.Init(player, c)
	c.Combat.SetPlayer(player)

	p := testhelper.DefaultProfile(testhelper.TestCharKey, testhelper.TestWeaponKey)
	idx, err := c.AddChar(p)
	if err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(idx)
	return c, addTargetToCore(c)
}

func countReactions(c *core.Core) (*int, *int) {
	var s, p int
	c.Events.Subscribe(event.OnStellarSwirl, func(...any) { s++ }, "stellar-swirl-count")
	c.Events.Subscribe(event.OnSwirlCryo, func(...any) { p++ }, "cryo-swirl-count")
	return &s, &p
}

func freezeAura(trg *testTarget) {
	trg.AttachOrRefill(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25},
	})
	trg.React(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Hydro, Durability: 25},
	})
}

func anemoHit(dur info.Durability) *info.AttackEvent {
	return &info.AttackEvent{
		Info: info.AttackInfo{
			Abil:       "anemo-gauge",
			Element:    attributes.Anemo,
			Durability: dur,
		},
	}
}

func TestStellarSwirlCryoConsumesCryo(t *testing.T) {
	c, trg := stellarCore(t)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Flags.Custom[StellarSwirlEnableKey] = 1
	stellar, plain := countReactions(c)

	trg.AttachOrRefill(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25},
	})
	before := trg.GetAuraDurability(info.ReactionModKeyCryo)
	if !durApproxEqual(20, before, 0.0001) {
		t.Fatalf("cryo before swirl %v", before)
	}
	trg.React(anemoHit(25))
	if *stellar != 1 || *plain != 0 {
		t.Fatalf("stellar %d plain %d", *stellar, *plain)
	}
	got := trg.GetAuraDurability(info.ReactionModKeyCryo)
	if !durApproxEqual(7.5, got, 0.0001) {
		t.Fatalf("cryo after 1U anemo %v", got)
	}
}

func TestStellarSwirlFrozenConsumesFrozen(t *testing.T) {
	c, trg := stellarCore(t)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Flags.Custom[StellarSwirlEnableKey] = 1
	stellar, plain := countReactions(c)

	freezeAura(trg)
	frozen := trg.GetAuraDurability(info.ReactionModKeyFrozen)
	cryo := trg.GetAuraDurability(info.ReactionModKeyCryo)
	if !durApproxEqual(40, frozen, 0.0001) || cryo > info.ZeroDur {
		t.Fatalf("frozen %v cryo %v", frozen, cryo)
	}
	trg.React(anemoHit(25))
	if *stellar != 1 || *plain != 0 {
		t.Fatalf("stellar %d plain %d", *stellar, *plain)
	}
	if got := trg.GetAuraDurability(info.ReactionModKeyFrozen); !durApproxEqual(27.5, got, 0.0001) {
		t.Fatalf("frozen after 1U anemo %v, want 27.5", got)
	}
	if got := trg.GetAuraDurability(info.ReactionModKeyCryo); got > info.ZeroDur {
		t.Fatalf("cryo appeared %v", got)
	}
}

func TestStellarSwirlFrozenDoesNotSpendCoexistingCryo(t *testing.T) {
	c, trg := stellarCore(t)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Flags.Custom[StellarSwirlEnableKey] = 1
	stellar, plain := countReactions(c)

	// 1U cryo coexists with frozen. A 2U anemo swirl spends the cryo aura and
	// still has gauge left, so the frozen aura is a second Stellar Swirl.
	trg.SetAuraDurability(info.ReactionModKeyCryo, 20, 0)
	trg.SetAuraDurability(info.ReactionModKeyFrozen, 40, 0)
	trg.React(anemoHit(50))
	if *stellar != 2 || *plain != 0 {
		t.Fatalf("stellar %d plain %d", *stellar, *plain)
	}
	if got := trg.GetAuraDurability(info.ReactionModKeyCryo); got > info.ZeroDur {
		t.Fatalf("cryo left %v", got)
	}
	if got := trg.GetAuraDurability(info.ReactionModKeyFrozen); !durApproxEqual(35, got, 0.0001) {
		t.Fatalf("frozen left %v, want 35", got)
	}
}

func TestOrdinaryFrozenSwirlStillConsumesFrozen(t *testing.T) {
	c, trg := stellarCore(t)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	stellar, plain := countReactions(c)
	freezeAura(trg)
	trg.React(anemoHit(25))
	if *stellar != 0 || *plain != 1 {
		t.Fatalf("stellar %d plain %d", *stellar, *plain)
	}
	if got := trg.GetAuraDurability(info.ReactionModKeyFrozen); !durApproxEqual(27.5, got, 0.0001) {
		t.Fatalf("frozen after ordinary swirl %v", got)
	}
}
