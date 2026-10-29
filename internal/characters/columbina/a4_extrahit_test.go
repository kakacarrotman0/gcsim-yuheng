package columbina

import (
	"testing"

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

func newCore(t *testing.T) (*core.Core, *enemy.Enemy, *char) {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1, Debug: false})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{}, 1))
	trg := enemy.New(c, info.EnemyProfile{
		Level:  100,
		Resist: make(map[attributes.Element]float64),
		Pos:    info.Coord{X: 0, Y: 0, R: 1},
	})
	c.Combat.AddEnemy(trg)
	c.Combat.DefaultTarget = trg.Key()
	prof := testhelper.DefaultProfile(keys.Columbina, testhelper.TestWeaponKey)
	prof.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	if _, err := c.AddChar(prof); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	col := c.Player.ByIndex(0).Character.(*char)
	if col.Base.Ascension < 4 {
		t.Fatalf("ascension %d", col.Base.Ascension)
	}
	return c, trg, col
}

func step(c *core.Core, n int) {
	for range n {
		c.F++
		c.Tick()
	}
}

func watch(c *core.Core) *[]string {
	hits := []string{}
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hits = append(hits, args[1].(*info.AttackEvent).Info.Abil)
	}, "a4-hits")
	return &hits
}

func count(hits []string, abil string) int {
	n := 0
	for _, name := range hits {
		if name == abil {
			n++
		}
	}
	return n
}

func armAuras(trg *enemy.Enemy, src int) {
	trg.SetAuraDurability(info.ReactionModKeyHydro, 100000, src)
	trg.SetAuraDurability(info.ReactionModKeyElectro, 100000, src)
}

func lcStrikes(c *core.Core, trg *enemy.Enemy, hits *[]string, n int) int {
	before := len(*hits)
	armAuras(trg, 0)
	for range n {
		trg.DoLCAttack()
	}
	step(c, 40)
	return count((*hits)[before:], string(info.ReactionTypeLunarCharged))
}

func lcrStrikes(c *core.Core, trg *enemy.Enemy, hits *[]string, n int) int {
	before := len(*hits)
	var contrib [info.MaxChars]bool
	contrib[0] = true
	for range n {
		reactable.DoLCrAttackWithContrib(contrib, trg, c, 0)
	}
	step(c, 60)
	return count((*hits)[before:], string(info.ReactionTypeLunarCrystallize))
}

func emitLC(c *core.Core, trg *enemy.Enemy) {
	c.Events.Emit(event.OnLunarCharged, trg, &info.AttackEvent{})
}

func emitLCr(c *core.Core, trg *enemy.Enemy) {
	c.Events.Emit(event.OnLunarCrystallize, trg, &info.AttackEvent{})
}

func TestA4ArmsAndClearsEachExtraHitSeparately(t *testing.T) {
	c, trg, col := newCore(t)
	col.applyBurstBuff()
	emitLC(c, trg)
	if c.Flags.Custom[reactable.LcExtraHitOverride] != 0.33 {
		t.Fatalf("lunar-charged chance %v", c.Flags.Custom[reactable.LcExtraHitOverride])
	}
	if _, ok := c.Flags.Custom[reactable.LcrExtraHitOverride]; ok {
		t.Fatal("lunar-charged activation armed lunar-crystallize")
	}
	emitLCr(c, trg)
	if c.Flags.Custom[reactable.LcrExtraHitOverride] != 0.33 || c.Flags.Custom[reactable.LcExtraHitOverride] != 0.33 {
		t.Fatalf("flags lc %v lcr %v", c.Flags.Custom[reactable.LcExtraHitOverride], c.Flags.Custom[reactable.LcrExtraHitOverride])
	}

	col.DeleteReactBonusMod(burstBuffKey)
	emitLCr(c, trg)
	if _, ok := c.Flags.Custom[reactable.LcrExtraHitOverride]; ok {
		t.Fatal("leaving the domain did not clear the lunar-crystallize chance")
	}
	if c.Flags.Custom[reactable.LcExtraHitOverride] != 0.33 {
		t.Fatal("clearing lunar-crystallize also cleared lunar-charged")
	}
	emitLC(c, trg)
	if _, ok := c.Flags.Custom[reactable.LcExtraHitOverride]; ok {
		t.Fatal("leaving the domain did not clear the lunar-charged chance")
	}
	if _, ok := c.Flags.Custom[reactable.LcrExtraHitOverride]; ok {
		t.Fatal("lunar-charged cleanup restored lunar-crystallize")
	}
}

func TestExtraHitsFollowOnlyTheirOwnChance(t *testing.T) {
	const strikes = 40
	lcOnly := strikesOf(t, strikes, true, false)
	lcrOnly := strikesOf(t, strikes, false, true)
	both := strikesOf(t, strikes, true, true)
	neither := strikesOf(t, strikes, false, false)

	// Seed 1, 40 strikes. Base counts are 40 Lunar-Charged hits and 120
	// Lunar-Crystallize hits. The extra hits above those bases are the 33% rolls.
	// Lunar-Charged is rolled first. Its extra rolls move the later
	// Lunar-Crystallize sequence from 161 to 160. That is one shared generator,
	// and each reader still checks only its own chance.
	if lcOnly != (strikeCounts{lc: 54, lcr: strikes * 3}) {
		t.Fatalf("lunar-charged only %+v", lcOnly)
	}
	if lcrOnly != (strikeCounts{lc: strikes, lcr: 161}) {
		t.Fatalf("lunar-crystallize only %+v", lcrOnly)
	}
	if both != (strikeCounts{lc: 54, lcr: 160}) {
		t.Fatalf("both %+v", both)
	}
	if neither != (strikeCounts{lc: strikes, lcr: strikes * 3}) {
		t.Fatalf("neither %+v", neither)
	}
}

type strikeCounts struct {
	lc  int
	lcr int
}

func strikesOf(t *testing.T, n int, lc, lcr bool) strikeCounts {
	t.Helper()
	c, trg, _ := newCore(t)
	hits := watch(c)
	if lc {
		c.Flags.Custom[reactable.LcExtraHitOverride] = 0.33
	}
	if lcr {
		c.Flags.Custom[reactable.LcrExtraHitOverride] = 0.33
	}
	return strikeCounts{
		lc:  lcStrikes(c, trg, hits, n),
		lcr: lcrStrikes(c, trg, hits, n),
	}
}

func TestClearingOneChanceLeavesTheOtherStrike(t *testing.T) {
	c, trg, _ := newCore(t)
	hits := watch(c)
	c.Flags.Custom[reactable.LcExtraHitOverride] = 0.33
	c.Flags.Custom[reactable.LcrExtraHitOverride] = 0.33
	delete(c.Flags.Custom, reactable.LcrExtraHitOverride)
	if got := lcStrikes(c, trg, hits, 40); got != 54 {
		t.Fatalf("lunar-charged strikes after lunar-crystallize was cleared: %d", got)
	}
	if got := lcrStrikes(c, trg, hits, 40); got != 120 {
		t.Fatalf("lunar-crystallize strikes after its chance was cleared: %d", got)
	}

	c, trg, _ = newCore(t)
	hits = watch(c)
	c.Flags.Custom[reactable.LcExtraHitOverride] = 0.33
	c.Flags.Custom[reactable.LcrExtraHitOverride] = 0.33
	delete(c.Flags.Custom, reactable.LcExtraHitOverride)
	if got := lcStrikes(c, trg, hits, 40); got != 40 {
		t.Fatalf("lunar-charged strikes after its chance was cleared: %d", got)
	}
	if got := lcrStrikes(c, trg, hits, 40); got != 161 {
		t.Fatalf("lunar-crystallize strikes after lunar-charged was cleared: %d", got)
	}
}

func TestExtraHitsStopWhenTheDomainBuffExpires(t *testing.T) {
	c, trg, col := newCore(t)
	hits := watch(c)
	col.applyBurstBuff()
	emitLC(c, trg)
	emitLCr(c, trg)
	if got := lcStrikes(c, trg, hits, 40); got != 54 {
		t.Fatalf("strikes while the domain is up: %d", got)
	}
	step(c, withinTimer+2)
	if col.StatusIsActive(burstBuffKey) {
		t.Fatal("domain buff still active")
	}
	if _, ok := c.Flags.Custom[reactable.LcExtraHitOverride]; ok {
		t.Fatal("lunar-charged chance survived the domain")
	}
	if _, ok := c.Flags.Custom[reactable.LcrExtraHitOverride]; ok {
		t.Fatal("lunar-crystallize chance survived the domain")
	}
	if got := lcStrikes(c, trg, hits, 40); got != 40 {
		t.Fatalf("lunar-charged strikes after expiry %d", got)
	}
	if got := lcrStrikes(c, trg, hits, 40); got != 120 {
		t.Fatalf("lunar-crystallize strikes after expiry %d", got)
	}
}

func TestBloomDoesNotGainExtraHitsAndDewFollowsTheDomain(t *testing.T) {
	c, trg, col := newCore(t)
	hits := watch(c)
	col.applyBurstBuff()
	c.Flags.Custom[reactable.LcExtraHitOverride] = 0.33
	c.Flags.Custom[reactable.LcrExtraHitOverride] = 0.33
	bloom(trg)
	step(c, 5)
	if c.Player.MoonridgeDew() != 1 {
		t.Fatalf("moonridge dew %d", c.Player.MoonridgeDew())
	}
	if count(*hits, string(info.ReactionTypeLunarCharged)) != 0 || count(*hits, string(info.ReactionTypeLunarCrystallize)) != 0 {
		t.Fatalf("bloom produced lunar strikes %v", *hits)
	}

	c, trg, col = newCore(t)
	col.applyBurstBuff()
	step(c, withinTimer+2)
	bloom(trg)
	if c.Player.MoonridgeDew() != 0 {
		t.Fatalf("dew after the domain expired %d", c.Player.MoonridgeDew())
	}
}

func bloom(trg *enemy.Enemy) {
	trg.AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Dendro, Durability: 25}})
	trg.React(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Hydro, Durability: 25, ActorIndex: 0}})
}
