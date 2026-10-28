package nefer

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/internal/template/dendrocore"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/reactable"
	"github.com/genshinsim/gcsim/pkg/testhelper"

	_ "github.com/genshinsim/gcsim/internal/characters/columbina"
)

func init() {
	testhelper.RegisterTestWeapon()
}

func makeCore(n int) (*core.Core, []*enemy.Enemy) {
	c, err := core.New(core.Opt{Seed: 1, Debug: false})
	if err != nil {
		panic(err)
	}
	a := avatar.New(c, info.Point{X: 0, Y: 0}, 1)
	c.Combat.SetPlayer(a)
	var trgs []*enemy.Enemy
	for range n {
		e := enemy.New(c, info.EnemyProfile{
			Level:  100,
			Resist: make(map[attributes.Element]float64),
			Pos:    info.Coord{X: 0, Y: 0, R: 1},
		})
		trgs = append(trgs, e)
		c.Combat.AddEnemy(e)
	}
	c.Player.SetActive(0)
	if len(trgs) > 0 {
		c.Combat.DefaultTarget = trgs[0].Key()
	}
	return c, trgs
}

func advance(c *core.Core, n int) {
	for range n {
		c.F++
		c.Tick()
	}
}

func addChar(t *testing.T, c *core.Core, k keys.Char, cons, talent int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(k, testhelper.TestWeaponKey)
	prof.Base.Cons = cons
	prof.Talents = info.TalentProfile{Attack: talent, Skill: talent, Burst: talent}
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("add %v: %v", k, err)
	}
	return c.Player.ByIndex(idx)
}

func start(t *testing.T, c *core.Core, active int) {
	t.Helper()
	c.Player.SetActive(active)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
}

func collect(c *core.Core) *[]info.AttackInfo {
	hits := []info.AttackInfo{}
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hits = append(hits, args[1].(*info.AttackEvent).Info)
	}, "hits")
	return &hits
}

func spawnCore(c *core.Core) {
	c.QueueAttackEvent(&info.AttackEvent{
		Info:    info.AttackInfo{Element: attributes.Dendro, Durability: 25},
		Pattern: combat.NewCircleHitOnTarget(info.Point{}, nil, 100),
	}, 0)
	advance(c, 1)
	c.QueueAttackEvent(&info.AttackEvent{
		Info:    info.AttackInfo{Element: attributes.Hydro, Durability: 50},
		Pattern: combat.NewCircleHitOnTarget(info.Point{}, nil, 100),
	}, 0)
	advance(c, dendrocore.Delay+1)
}

func TestIdentity(t *testing.T) {
	c, _ := makeCore(1)
	ch := addChar(t, c, keys.Nefer, 0, 1)
	start(t, c, ch.Index())
	if ch.Base.Element != attributes.Dendro {
		t.Fatalf("element %v", ch.Base.Element)
	}
	if ch.EnergyMax != 60 {
		t.Fatalf("energy %v", ch.EnergyMax)
	}
	if ch.Moonsign != 1 {
		t.Fatalf("moonsign %v", ch.Moonsign)
	}
	if math.Abs(ch.BaseStats[attributes.CD]-0.884) > 1e-6 {
		t.Fatalf("crit dmg %v", ch.BaseStats[attributes.CD])
	}
	if _, ok := c.Flags.Custom[reactable.LunarBloomEnableKey]; !ok {
		t.Fatal("lunar bloom was not enabled")
	}
}

func TestSkillOpensShadowDance(t *testing.T) {
	c, _ := makeCore(1)
	ch := addChar(t, c, keys.Nefer, 0, 1)
	start(t, c, ch.Index())
	hits := collect(c)
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillHit+1)
	if !ch.StatusIsActive(shadowKey) {
		t.Fatal("shadow dance inactive")
	}
	if len(*hits) != 1 {
		t.Fatalf("hits %d", len(*hits))
	}
	got := (*hits)[0]
	if got.Element != attributes.Dendro || got.Durability != 25 || got.ICDTag != attacks.ICDTagElementalArt {
		t.Fatalf("skill hit %+v", got)
	}
	if math.Abs(got.Mult-skillATK[0]) > 1e-9 {
		t.Fatalf("atk ratio %v", got.Mult)
	}
	if math.Abs(got.FlatDmg-skillEM[0]*ch.Stat(attributes.EM)) > 1e-6 {
		t.Fatalf("em flat %v em %v", got.FlatDmg, ch.Stat(attributes.EM))
	}
}

func TestPhantasmConsumesDew(t *testing.T) {
	c, _ := makeCore(1)
	ch := addChar(t, c, keys.Nefer, 0, 1)
	start(t, c, ch.Index())
	c.Player.SetVerdantDew(1)
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillAnim+1)
	hits := collect(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, ppAnim+1)
	if c.Player.Dew() != 0 {
		t.Fatalf("dew %v", c.Player.Dew())
	}
	want := []string{
		"Phantasm Performance (Nefer 1)",
		"Phantasm Performance (Nefer 2)",
		"Phantasm Performance (Shade 1)",
		"Phantasm Performance (Shade 2)",
		"Phantasm Performance (Shade 3)",
	}
	if len(*hits) != len(want) {
		t.Fatalf("hits %v", names(*hits))
	}
	for i, name := range want {
		if (*hits)[i].Abil != name {
			t.Fatalf("hit %d %v", i, (*hits)[i].Abil)
		}
	}
	shade := (*hits)[2]
	if shade.AttackTag != attacks.AttackTagDirectLunarBloom || !shade.UseEM || shade.Durability != 0 {
		t.Fatalf("shade %+v", shade)
	}
	if math.Abs(shade.Mult-shade1[ch.TalentLvlSkill()]) > 1e-9 {
		t.Fatalf("shade ratio %v", shade.Mult)
	}
	if !ch.StatusIsActive(shadowKey) {
		t.Fatal("one phantasm ended shadow dance")
	}
}

func TestThreePhantasmsEndDance(t *testing.T) {
	c, _ := makeCore(1)
	ch := addChar(t, c, keys.Nefer, 0, 1)
	start(t, c, ch.Index())
	c.Player.SetVerdantDew(3)
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillAnim+1)
	for range 3 {
		if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
			t.Fatal(err)
		}
		advance(c, ppAnim+1)
	}
	if ch.StatusIsActive(shadowKey) {
		t.Fatal("shadow dance still active")
	}
	if c.Player.Dew() != 0 {
		t.Fatalf("dew %v", c.Player.Dew())
	}
	hits := collect(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, caHit+1)
	if len(*hits) != 1 || (*hits)[0].Abil != "Charged Attack" {
		t.Fatalf("charge after three phantasms %v", names(*hits))
	}
	if c.Player.Dew() != 0 {
		t.Fatalf("ordinary charge spent dew %v", c.Player.Dew())
	}
}

func TestGleamSeedsBecomeDeceit(t *testing.T) {
	c, _ := makeCore(1)
	addChar(t, c, keys.Columbina, 0, 1)
	nef := addChar(t, c, keys.Nefer, 0, 1)
	start(t, c, nef.Index())
	if c.Player.GetMoonsignLevel() < 2 {
		t.Fatal("expected gleam")
	}
	spawnCore(c)
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillAnim+1)
	if !nef.StatusIsActive(deceitKey) {
		t.Fatal("deceit window inactive")
	}
	if c.Combat.GadgetCount() != 1 {
		t.Fatalf("skill removed the core, count %d", c.Combat.GadgetCount())
	}
	c.QueueAttackEvent(&info.AttackEvent{
		Info:    info.AttackInfo{Element: attributes.Electro, Durability: 25, ActorIndex: nef.Index()},
		Pattern: combat.NewCircleHitOnTarget(info.Point{}, nil, 100),
	}, 0)
	advance(c, 2)
	if c.Combat.GadgetCount() != 1 {
		t.Fatalf("deceit core was removed, count %d", c.Combat.GadgetCount())
	}
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 1)
	if c.Combat.GadgetCount() != 0 {
		t.Fatalf("absorb left %d", c.Combat.GadgetCount())
	}
}

func TestThreeSeedsGrantEM(t *testing.T) {
	c, _ := makeCore(1)
	addChar(t, c, keys.Columbina, 0, 1)
	nef := addChar(t, c, keys.Nefer, 0, 1)
	start(t, c, nef.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillAnim+1)
	for range 3 {
		spawnCore(c)
	}
	before := nef.Stat(attributes.EM)
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 1)
	if math.Abs(nef.Stat(attributes.EM)-before-100) > 1e-6 {
		t.Fatalf("em before %v after %v", before, nef.Stat(attributes.EM))
	}
}

func TestC6ConvertsSecondHit(t *testing.T) {
	c, _ := makeCore(1)
	addChar(t, c, keys.Columbina, 0, 1)
	nef := addChar(t, c, keys.Nefer, 6, 1)
	start(t, c, nef.Index())
	c.Player.SetVerdantDew(1)
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillAnim+1)
	hits := collect(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, ppAnim+1)
	got := names(*hits)
	want := "Phantasm Performance C6 (Nefer 2)"
	found := false
	for _, h := range *hits {
		if h.Abil == want {
			found = true
			if !h.UseEM || math.Abs(h.Mult-1.45) > 1e-9 || h.Elevation != 0.15 {
				t.Fatalf("c6 hit %+v", h)
			}
		}
	}
	if !found {
		t.Fatalf("missing %s in %v", want, got)
	}
}

func TestC4DewHaste(t *testing.T) {
	c, _ := makeCore(1)
	nef := addChar(t, c, keys.Nefer, 4, 1)
	start(t, c, nef.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 1)
	c.Player.OnLunarBloom()
	got := -1
	for i := 1; i <= 160; i++ {
		advance(c, 1)
		if c.Player.VerdantDew() > 0 {
			got = i
			break
		}
	}
	if got != 118 {
		t.Fatalf("dew gained after %d ticks", got)
	}
}

func TestC4ShredWhileShadowDance(t *testing.T) {
	c, trg := makeCore(1)
	nef := addChar(t, c, keys.Nefer, 4, 1)
	start(t, c, nef.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 1)
	if !trg[0].ResistModIsActive("nefer-c4") {
		t.Fatal("c4 shred inactive")
	}
}

func TestTalentCons(t *testing.T) {
	c, _ := makeCore(1)
	nef := addChar(t, c, keys.Nefer, 3, 1)
	start(t, c, nef.Index())
	if nef.TalentLvlSkill() != 3 {
		t.Fatalf("c3 skill index %d", nef.TalentLvlSkill())
	}
	c5, _ := makeCore(1)
	ch := addChar(t, c5, keys.Nefer, 5, 1)
	start(t, c5, ch.Index())
	if ch.TalentLvlBurst() != 3 {
		t.Fatalf("c5 burst index %d", ch.TalentLvlBurst())
	}
}

func TestShadowDanceEndsOnSwap(t *testing.T) {
	c, _ := makeCore(1)
	nef := addChar(t, c, keys.Nefer, 0, 1)
	col := addChar(t, c, keys.Columbina, 0, 1)
	start(t, c, nef.Index())
	c.Player.SetVerdantDew(2)
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillHit+1)
	if !nef.StatusIsActive(shadowKey) {
		t.Fatal("shadow dance did not start")
	}
	c.Events.Emit(event.OnCharacterSwap, col.Index(), nef.Index())
	if !nef.StatusIsActive(shadowKey) || c.Player.Active() != nef.Index() {
		t.Fatal("another character's swap cleared shadow dance")
	}

	if err := c.Player.Exec(action.ActionSwap, keys.Columbina, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 1)
	if c.Player.Active() != col.Index() {
		t.Fatalf("active %d", c.Player.Active())
	}
	if nef.StatusIsActive(shadowKey) {
		t.Fatal("shadow dance stayed up after nefer left")
	}
	if c.Player.Dew() != 2 {
		t.Fatalf("swap spent dew %v", c.Player.Dew())
	}

	advance(c, 61)
	if err := c.Player.Exec(action.ActionSwap, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 1)
	if c.Player.Active() != nef.Index() {
		t.Fatalf("active %d", c.Player.Active())
	}
	if nef.StatusIsActive(shadowKey) {
		t.Fatal("returning to the field restored shadow dance")
	}

	hits := collect(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, caHit+1)
	if len(*hits) != 1 || (*hits)[0].Abil != "Charged Attack" || (*hits)[0].AttackTag == attacks.AttackTagDirectLunarBloom {
		t.Fatalf("charge without shadow dance %+v", *hits)
	}
	if c.Player.Dew() != 2 {
		t.Fatalf("ordinary charge spent dew %v", c.Player.Dew())
	}

	for c.F < skillCD+skillHit+5 && !readySkill(nef) {
		advance(c, 1)
	}
	if !readySkill(nef) {
		t.Fatal("skill did not come off cooldown")
	}
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	if !nef.StatusIsActive(shadowKey) {
		t.Fatal("skill did not reopen shadow dance")
	}
	advance(c, skillHit+1)
	*hits = nil
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, ppAnim+1)
	if len(*hits) == 0 || (*hits)[0].Abil != "Phantasm Performance (Nefer 1)" {
		t.Fatalf("re-entered charge %v", names(*hits))
	}
	if c.Player.Dew() != 1 {
		t.Fatalf("phantasm dew %v", c.Player.Dew())
	}
}

func TestShadowDanceExpires(t *testing.T) {
	c, _ := makeCore(1)
	nef := addChar(t, c, keys.Nefer, 0, 1)
	start(t, c, nef.Index())
	c.Player.SetVerdantDew(1)
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, shadowDur+1)
	if nef.StatusIsActive(shadowKey) {
		t.Fatal("shadow dance outlasted its duration")
	}
	hits := collect(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, caHit+1)
	if len(*hits) != 1 || (*hits)[0].Abil != "Charged Attack" {
		t.Fatalf("charge after expiry %v", names(*hits))
	}
	if c.Player.Dew() != 1 {
		t.Fatalf("expired charge spent dew %v", c.Player.Dew())
	}
}

func TestQueuedPhantasmStillLandsAfterSwap(t *testing.T) {
	c, _ := makeCore(1)
	nef := addChar(t, c, keys.Nefer, 0, 1)
	addChar(t, c, keys.Columbina, 0, 1)
	start(t, c, nef.Index())
	c.Player.SetVerdantDew(1)
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillHit+1)
	hits := collect(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	if err := c.Player.Exec(action.ActionSwap, keys.Columbina, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 1)
	if nef.StatusIsActive(shadowKey) {
		t.Fatal("swap left shadow dance active")
	}
	advance(c, ppAnim)
	found := false
	for _, hit := range *hits {
		if hit.Abil == "Phantasm Performance (Shade 1)" {
			found = true
		}
	}
	if !found {
		t.Fatalf("queued shade did not land: %v", names(*hits))
	}
	if c.Player.Dew() != 0 {
		t.Fatalf("queued dew spend %v", c.Player.Dew())
	}
}

func readySkill(ch *character.CharWrapper) bool {
	ok, reason := ch.ActionReady(action.ActionSkill, nil)
	return ok && reason == action.NoFailure
}

func names(hits []info.AttackInfo) []string {
	out := make([]string, len(hits))
	for i, h := range hits {
		out[i] = h.Abil
	}
	return out
}
