package reactable_test

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/ineffa"
	_ "github.com/genshinsim/gcsim/internal/characters/kaeya"
	_ "github.com/genshinsim/gcsim/internal/characters/linnea"
	_ "github.com/genshinsim/gcsim/internal/characters/xingqiu"
	_ "github.com/genshinsim/gcsim/internal/characters/zhongli"
	_ "github.com/genshinsim/gcsim/internal/characters/zibai"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/construct"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/reactable"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func lunarCore(t *testing.T) (*core.Core, *enemy.Enemy) {
	t.Helper()
	c, err := core.New(core.Opt{Seed: 1, Debug: false})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{X: 0, Y: 0}, 1))
	e := enemy.New(c, info.EnemyProfile{
		Level:  100,
		Resist: make(map[attributes.Element]float64),
		Pos:    info.Coord{X: 0, Y: 0, R: 1},
	})
	c.Combat.AddEnemy(e)
	c.Combat.DefaultTarget = e.Key()
	c.Player.SetActive(0)
	return c, e
}

func addLunarChar(t *testing.T, c *core.Core, key keys.Char, flatDEF float64) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(key, testhelper.TestWeaponKey)
	prof.Base.Cons = 0
	prof.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	if flatDEF != 0 {
		prof.Stats[attributes.DEF] = flatDEF
	}
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("add %v: %v", key, err)
	}
	return c.Player.ByIndex(idx)
}

func initLunar(t *testing.T, c *core.Core) {
	t.Helper()
	c.Player.SetActive(0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
}

func step(c *core.Core, n int) {
	for range n {
		c.F++
		c.Tick()
	}
}

func apply(c *core.Core, trg *enemy.Enemy, actor int, ele attributes.Element, dur float64, tag attacks.AttackTag, abil string) {
	c.QueueAttack(info.AttackInfo{
		ActorIndex: actor,
		Abil:       abil,
		AttackTag:  tag,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    ele,
		Durability: info.Durability(dur),
		Mult:       1,
	}, combat.NewSingleTargetHit(trg.Key()), 0, 0)
}

func lcrDEFBonus(ch *character.CharWrapper) float64 {
	return math.Min(ch.TotalDef(false)/100*0.007, 0.14)
}

func TestHydroCrystallizeStaysOrdinaryWithoutTheFlag(t *testing.T) {
	c, trg := lunarCore(t)
	zhong := addLunarChar(t, c, keys.Zhongli, 0)
	initLunar(t, c)
	if _, ok := c.Flags.Custom[reactable.LunarCrystallizeEnableKey]; ok {
		t.Fatal("lunar crystallize flag was set without a lunar-crystallize character")
	}
	var hydro, lunar int
	c.Events.Subscribe(event.OnCrystallizeHydro, func(...any) { hydro++ }, "hydro")
	c.Events.Subscribe(event.OnLunarCrystallize, func(...any) { lunar++ }, "lunar")
	trg.AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Hydro, Durability: 50}})
	step(c, 1)
	apply(c, trg, zhong.Index(), attributes.Geo, 25, attacks.AttackTagElementalArt, "geo")
	if hydro != 1 || lunar != 0 {
		t.Fatalf("hydro %d lunar %d", hydro, lunar)
	}
	if got, _ := c.Constructs.ConstructsByType(construct.GeoConstructLunarCrystallize); len(got) != 0 {
		t.Fatalf("moondrift constructs %d", len(got))
	}
}

func TestLunarCrystallizeEventsAndDamageTags(t *testing.T) {
	c, trg := lunarCore(t)
	zibai := addLunarChar(t, c, keys.Zibai, 0)
	initLunar(t, c)
	if c.Flags.Custom[reactable.LunarCrystallizeEnableKey] != 1 {
		t.Fatal("zibai did not enable lunar crystallize")
	}
	var hydro, lunar, harmony int
	var reaction, landed, direct info.AttackInfo
	c.Events.Subscribe(event.OnCrystallizeHydro, func(...any) { hydro++ }, "hydro")
	c.Events.Subscribe(event.OnLunarCrystallize, func(...any) { lunar++ }, "lunar")
	c.Events.Subscribe(event.OnMoondriftHarmony, func(...any) { harmony++ }, "harmony")
	c.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.AttackTag == attacks.AttackTagReactionLunarCrystallize {
			reaction = atk.Info
		}
	}, "reaction")
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent).Info
		switch atk.Abil {
		case "direct-probe":
			direct = atk
		case string(info.ReactionTypeLunarCrystallize):
			landed = atk
		}
	}, "hits")

	for range 3 {
		trg.AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Hydro, Durability: 50}})
		step(c, 1)
		apply(c, trg, zibai.Index(), attributes.Geo, 25, attacks.AttackTagElementalArt, "geo-probe")
	}
	for c.F < 80 && (reaction.Abil == "" || landed.Abil == "") {
		step(c, 1)
	}
	if hydro != 0 || lunar != 3 || harmony != 1 {
		t.Fatalf("hydro %d lunar %d harmony %d", hydro, lunar, harmony)
	}
	moondrift, _ := c.Constructs.ConstructsByType(construct.GeoConstructLunarCrystallize)
	if len(moondrift) < 3 {
		t.Fatalf("moondrift constructs %d", len(moondrift))
	}
	if reaction.AttackTag != attacks.AttackTagReactionLunarCrystallize || reaction.Abil != string(info.ReactionTypeLunarCrystallize) || reaction.Element != attributes.Geo {
		t.Fatalf("reaction info %+v", reaction)
	}
	if landed.AttackTag != attacks.AttackTagReactionLunarCrystallize || landed.Element != attributes.Geo || landed.IgnoreDefPercent != 1 {
		t.Fatalf("landed reaction %+v", landed)
	}
	want := lcrDEFBonus(zibai)
	if math.Abs(reaction.BaseDmgBonus-want) > 1e-9 {
		t.Fatalf("zibai reaction bonus %v want %v", reaction.BaseDmgBonus, want)
	}

	apply(c, trg, zibai.Index(), attributes.Geo, 0, attacks.AttackTagDirectLunarCrystallize, "direct-probe")
	if direct.AttackTag != attacks.AttackTagDirectLunarCrystallize || direct.Abil == string(info.ReactionTypeLunarCrystallize) {
		t.Fatalf("direct info %+v", direct)
	}
	if math.Abs(direct.BaseDmgBonus-want) > 1e-9 {
		t.Fatalf("zibai direct bonus %v want %v", direct.BaseDmgBonus, want)
	}
}

func TestZibaiAndLinneaLunarBonusesStack(t *testing.T) {
	c, trg := lunarCore(t)
	zibai := addLunarChar(t, c, keys.Zibai, 20000)
	linnea := addLunarChar(t, c, keys.Linnea, 20000)
	initLunar(t, c)
	zBonus := lcrDEFBonus(zibai)
	lBonus := lcrDEFBonus(linnea)
	if zBonus != 0.14 || lBonus != 0.14 {
		t.Fatalf("expected each DEF bonus to reach its own 0.14 cap, zibai %v linnea %v", zBonus, lBonus)
	}
	var reaction, direct float64
	var sawReaction, sawDirect bool
	c.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.AttackTag != attacks.AttackTagReactionLunarCrystallize {
			return
		}
		reaction = atk.Info.BaseDmgBonus
		sawReaction = true
	}, "reaction")
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent).Info
		if atk.Abil != "direct-probe" {
			return
		}
		direct = atk.BaseDmgBonus
		sawDirect = true
	}, "direct")

	for range 3 {
		trg.AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Hydro, Durability: 50}})
		step(c, 1)
		apply(c, trg, zibai.Index(), attributes.Geo, 25, attacks.AttackTagElementalArt, "geo-probe")
	}
	for c.F < 80 && !sawReaction {
		step(c, 1)
	}
	apply(c, trg, zibai.Index(), attributes.Geo, 0, attacks.AttackTagDirectLunarCrystallize, "direct-probe")
	if !sawReaction || !sawDirect {
		t.Fatal("missing a lunar-crystallize bonus event")
	}
	if math.Abs(reaction-(zBonus+lBonus)) > 1e-9 || math.Abs(direct-(zBonus+lBonus)) > 1e-9 {
		t.Fatalf("reaction %v direct %v want %v", reaction, direct, zBonus+lBonus)
	}
}

func TestLunarChargedRequiresTheEnableFlag(t *testing.T) {
	c, trg := lunarCore(t)
	xingqiu := addLunarChar(t, c, keys.Xingqiu, 0)
	kaeya := addLunarChar(t, c, keys.Kaeya, 0)
	initLunar(t, c)
	if _, ok := c.Flags.Custom[reactable.LunarChargeEnableKey]; ok {
		t.Fatal("lunar charged flag was set without ineffa, flins, or columbina")
	}
	var ec, lc int
	c.Events.Subscribe(event.OnElectroCharged, func(...any) { ec++ }, "ec")
	c.Events.Subscribe(event.OnLunarCharged, func(...any) { lc++ }, "lc")
	trg.AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Hydro, Durability: 50, ActorIndex: xingqiu.Index()}})
	step(c, 1)
	apply(c, trg, kaeya.Index(), attributes.Electro, 25, attacks.AttackTagElementalArt, "electro")
	if ec != 1 || lc != 0 || c.Status.Duration(reactable.LcKey) != 0 {
		t.Fatalf("electrocharged %d lunar %d cloud %d", ec, lc, c.Status.Duration(reactable.LcKey))
	}
}

func TestIneffaLunarChargedCloudAndReactionTag(t *testing.T) {
	c, trg := lunarCore(t)
	ineffa := addLunarChar(t, c, keys.Ineffa, 0)
	xingqiu := addLunarChar(t, c, keys.Xingqiu, 0)
	initLunar(t, c)
	if c.Flags.Custom[reactable.LunarChargeEnableKey] != 1 {
		t.Fatal("ineffa did not enable lunar charged")
	}
	var ec, lc int
	var reaction info.AttackInfo
	var landed info.AttackInfo
	c.Events.Subscribe(event.OnElectroCharged, func(...any) { ec++ }, "ec")
	c.Events.Subscribe(event.OnLunarCharged, func(...any) { lc++ }, "lc")
	c.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.AttackTag == attacks.AttackTagReactionLunarCharge {
			reaction = atk.Info
		}
	}, "reaction")
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		atk := args[1].(*info.AttackEvent).Info
		if atk.Abil == string(info.ReactionTypeLunarCharged) {
			landed = atk
		}
	}, "landed")

	trg.AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Hydro, Durability: 50, ActorIndex: xingqiu.Index()}})
	step(c, 1)
	hydroBefore := trg.GetAuraDurability(info.ReactionModKeyHydro)
	apply(c, trg, ineffa.Index(), attributes.Electro, 25, attacks.AttackTagElementalArt, "electro")
	electroBefore := trg.GetAuraDurability(info.ReactionModKeyElectro)
	if ec != 0 || lc != 1 || c.Status.Duration(reactable.LcKey) == 0 {
		t.Fatalf("electrocharged %d lunar %d cloud %d", ec, lc, c.Status.Duration(reactable.LcKey))
	}
	for c.F < 40 && landed.Abil == "" {
		step(c, 1)
	}
	if landed.AttackTag != attacks.AttackTagReactionLunarCharge || landed.Element != attributes.Electro || landed.IgnoreDefPercent != 1 {
		t.Fatalf("landed reaction %+v", landed)
	}
	want := math.Min(ineffa.TotalAtk()/100*0.007, 0.14)
	if want <= 0 || math.Abs(reaction.BaseDmgBonus-want) > 1e-9 {
		t.Fatalf("ineffa bonus %v want %v", reaction.BaseDmgBonus, want)
	}
	if trg.GetAuraDurability(info.ReactionModKeyHydro) > hydroBefore-10+1e-6 {
		t.Fatalf("hydro aura %v from %v", trg.GetAuraDurability(info.ReactionModKeyHydro), hydroBefore)
	}
	if trg.GetAuraDurability(info.ReactionModKeyElectro) > electroBefore-10+1e-6 {
		t.Fatalf("electro aura %v from %v", trg.GetAuraDurability(info.ReactionModKeyElectro), electroBefore)
	}
}

func TestFrozenRejectsLunarCharged(t *testing.T) {
	c, trg := lunarCore(t)
	ineffa := addLunarChar(t, c, keys.Ineffa, 0)
	initLunar(t, c)
	var ec, lc int
	c.Events.Subscribe(event.OnElectroCharged, func(...any) { ec++ }, "ec")
	c.Events.Subscribe(event.OnLunarCharged, func(...any) { lc++ }, "lc")
	trg.AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25}})
	trg.React(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Hydro, Durability: 25}})
	if trg.GetAuraDurability(info.ReactionModKeyFrozen) <= info.ZeroDur {
		t.Fatal("frozen aura was not present")
	}
	apply(c, trg, ineffa.Index(), attributes.Electro, 25, attacks.AttackTagElementalArt, "electro")
	if ec != 0 || lc != 0 || c.Status.Duration(reactable.LcKey) != 0 {
		t.Fatalf("frozen still reacted electrocharged %d lunar %d cloud %d", ec, lc, c.Status.Duration(reactable.LcKey))
	}
}

func TestLunarExtraHitOverrideKeysDiffer(t *testing.T) {
	c, _ := lunarCore(t)
	if reactable.LcExtraHitOverride == reactable.LcrExtraHitOverride {
		t.Fatal("lunar-charged and lunar-crystallize extra-hit keys are still the same string")
	}
	c.Flags.Custom[reactable.LcExtraHitOverride] = 0.33
	if _, ok := c.Flags.Custom[reactable.LcrExtraHitOverride]; ok {
		t.Fatal("setting the lunar-charged chance also set the lunar-crystallize chance")
	}
	c.Flags.Custom[reactable.LcrExtraHitOverride] = 0.33
	delete(c.Flags.Custom, reactable.LcExtraHitOverride)
	if c.Flags.Custom[reactable.LcrExtraHitOverride] != 0.33 {
		t.Fatal("clearing the lunar-charged chance cleared the lunar-crystallize chance")
	}
}

// The 1.6 and 3 multipliers are the checked-in CalcSpecialReactionDmg branches.
// This batch did not re-derive them from a new source.
func TestCheckedInSpecialReactionMultipliers(t *testing.T) {
	base := combat.CalcReactionBaseDmg(90)
	lcr := info.AttackInfo{AttackTag: attacks.AttackTagReactionLunarCrystallize}
	lc := info.AttackInfo{AttackTag: attacks.AttackTagReactionLunarCharge}
	if math.Abs(combat.CalcSpecialReactionDmg(90, 0, lcr, 0)-1.6*base) > 1e-6 {
		t.Fatalf("lunar crystallize multiplier moved")
	}
	if math.Abs(combat.CalcSpecialReactionDmg(90, 0, lc, 0)-3*base) > 1e-6 {
		t.Fatalf("lunar charged multiplier moved")
	}
}
