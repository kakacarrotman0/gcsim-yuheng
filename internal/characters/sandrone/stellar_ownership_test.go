package sandrone

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/traveler/cryo/lumine"
	_ "github.com/genshinsim/gcsim/internal/characters/vesna"
	_ "github.com/genshinsim/gcsim/internal/characters/wriothesley"
	_ "github.com/genshinsim/gcsim/internal/characters/xiangling"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/reactable"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func addChar(t *testing.T, c *core.Core, key keys.Char, atk float64, params map[string]int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(key, testhelper.TestWeaponKey)
	prof.Base.Level = 90
	prof.Base.MaxLevel = 90
	prof.Talents = info.TalentProfile{Attack: 10, Skill: 10, Burst: 10}
	if atk != 0 {
		prof.Stats[attributes.ATK] = atk
	}
	for k, v := range params {
		prof.Params[k] = v
	}
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("add %v: %v", key, err)
	}
	return c.Player.ByIndex(idx)
}

func swirlCryo(trg *enemy.Enemy, applier, trigger int) {
	trg.AttachOrRefill(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25, ActorIndex: applier},
	})
	trg.React(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Anemo, Durability: 25, ActorIndex: trigger},
	})
}

func vortexes(c *core.Core) int {
	n := 0
	for _, g := range c.Combat.Gadgets() {
		if g != nil && g.GadgetTyp() == info.GadgetTypStellarVortex {
			n++
		}
	}
	return n
}

func detonateContributors(t *testing.T, c *core.Core) [info.MaxChars]bool {
	t.Helper()
	var got [info.MaxChars]bool
	var n int
	c.Events.Subscribe(event.OnStellarVortexDetonate, func(args ...any) {
		n++
		got = *args[1].(*[info.MaxChars]bool)
	}, "vortex-contributors")
	advance(c, 200)
	if n != 1 {
		t.Fatalf("vortex detonations %d", n)
	}
	return got
}

func TestOdetteAndSandroneShareOneSwirl(t *testing.T) {
	c, trg := makeCore()
	sand := addChar(t, c, keys.Sandrone, 8000, nil)
	od := addChar(t, c, keys.Odette, 8000, nil)
	kaeya := addChar(t, c, keys.Kaeya, 0, nil)
	sucrose := addChar(t, c, keys.Sucrose, 0, nil)
	mustInit(t, c)
	if c.Flags.Custom[reactable.StellarSwirlEnableKey] != 1 || c.Flags.Custom[reactable.StellarConductEnableKey] != 1 {
		t.Fatalf("flags swirl %v conduct %v", c.Flags.Custom[reactable.StellarSwirlEnableKey], c.Flags.Custom[reactable.StellarConductEnableKey])
	}

	var stellar, plain int
	var bonus float64
	var snaps int
	c.Events.Subscribe(event.OnStellarSwirl, func(...any) { stellar++ }, "own-swirl")
	c.Events.Subscribe(event.OnSwirlCryo, func(...any) { plain++ }, "own-plain")
	c.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.AttackTag != attacks.AttackTagReactionStellarSwirl {
			return
		}
		snaps++
		bonus = atk.Info.BaseDmgBonus
	}, "own-bonus")

	swirlCryo(trg[0], kaeya.Index(), sucrose.Index())
	if stellar != 1 || plain != 0 || vortexes(c) != 1 {
		t.Fatalf("stellar %d plain %d vortex %d", stellar, plain, vortexes(c))
	}
	if sand.StatusDuration(radianceKey) != reactable.RadianceStellarSwirlBase {
		t.Fatalf("sandrone radiance %d", sand.StatusDuration(radianceKey))
	}
	if od.StatusDuration(radianceKey) != reactable.RadianceStellarSwirlBase {
		t.Fatalf("odette radiance %d", od.StatusDuration(radianceKey))
	}
	if sucrose.StatusIsActive(radianceKey) || kaeya.StatusIsActive(radianceKey) {
		t.Fatal("swirl radiance spread to a character without the passive")
	}
	want := 0.14 + 0.14
	if snaps == 0 || math.Abs(bonus-want) > 1e-9 {
		t.Fatalf("stacked base bonus %v snaps %d", bonus, snaps)
	}
}

func TestSandroneBonusAlone(t *testing.T) {
	c, trg := makeCore()
	addChar(t, c, keys.Sandrone, 8000, nil)
	kaeya := addChar(t, c, keys.Kaeya, 0, nil)
	sucrose := addChar(t, c, keys.Sucrose, 0, nil)
	mustInit(t, c)
	var bonus float64
	c.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.AttackTag == attacks.AttackTagReactionStellarSwirl {
			bonus = atk.Info.BaseDmgBonus
		}
	}, "sand-only")
	swirlCryo(trg[0], kaeya.Index(), sucrose.Index())
	if math.Abs(bonus-0.14) > 1e-9 {
		t.Fatalf("sandrone bonus %v", bonus)
	}
}

func TestFrozenSwirlContributors(t *testing.T) {
	c, trg := makeCore()
	sand := addChar(t, c, keys.Sandrone, 0, nil)
	kaeya := addChar(t, c, keys.Kaeya, 0, nil)
	xiangling := addChar(t, c, keys.Xiangling, 0, nil)
	sucrose := addChar(t, c, keys.Sucrose, 0, nil)
	mustInit(t, c)

	var stellar int
	c.Events.Subscribe(event.OnStellarSwirl, func(...any) { stellar++ }, "frz-swirl")
	trg[0].AttachOrRefill(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25, ActorIndex: kaeya.Index()},
	})
	trg[0].React(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Hydro, Durability: 25, ActorIndex: kaeya.Index()},
	})
	trg[0].AttachOrRefill(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25, ActorIndex: xiangling.Index()},
	})
	trg[0].React(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Hydro, Durability: 25, ActorIndex: xiangling.Index()},
	})
	before := trg[0].GetAuraDurability(info.ReactionModKeyFrozen)
	trg[0].React(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Anemo, Durability: 25, ActorIndex: sucrose.Index()},
	})
	if stellar != 1 {
		t.Fatalf("stellar swirls %d", stellar)
	}
	after := trg[0].GetAuraDurability(info.ReactionModKeyFrozen)
	if after >= before {
		t.Fatalf("frozen %v -> %v", before, after)
	}
	if sand.StatusDuration(radianceKey) != reactable.RadianceStellarSwirlBase {
		t.Fatalf("radiance from frozen swirl %d", sand.StatusDuration(radianceKey))
	}

	got := detonateContributors(t, c)
	if !got[sucrose.Index()] || !got[kaeya.Index()] {
		t.Fatalf("contributors %+v", got)
	}
	if got[xiangling.Index()] || got[sand.Index()] {
		t.Fatalf("unexpected contributors %+v", got)
	}
}

func TestCryoAuraContributorIgnoresElement(t *testing.T) {
	c, trg := makeCore()
	sand := addChar(t, c, keys.Sandrone, 0, nil)
	kaeya := addChar(t, c, keys.Kaeya, 0, nil)
	xiangling := addChar(t, c, keys.Xiangling, 0, nil)
	sucrose := addChar(t, c, keys.Sucrose, 0, nil)
	mustInit(t, c)
	swirlCryo(trg[0], xiangling.Index(), sucrose.Index())
	got := detonateContributors(t, c)
	if !got[sucrose.Index()] || !got[xiangling.Index()] {
		t.Fatalf("cryo contributors %+v", got)
	}
	if got[kaeya.Index()] || got[sand.Index()] {
		t.Fatalf("unexpected cryo contributors %+v", got)
	}
}

func TestQiqiRevelationRadiance(t *testing.T) {
	c, trg := makeCore()
	addChar(t, c, keys.Sandrone, 0, nil)
	qiqi := addChar(t, c, keys.Qiqi, 0, nil)
	yae := addChar(t, c, keys.YaeMiko, 0, nil)
	mustInit(t, c)
	c.Player.SetActive(qiqi.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Qiqi, nil); err != nil {
		t.Fatal(err)
	}
	swirlCryo(trg[0], qiqi.Index(), qiqi.Index())
	if qiqi.StatusDuration(radianceKey) != reactable.RadianceStellarSwirlBase {
		t.Fatalf("qiqi radiance %d", qiqi.StatusDuration(radianceKey))
	}
	swirlBonus := qiqi.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStellarSwirl})
	conductBonus := qiqi.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagDirectStellarConduct})
	if math.Abs(swirlBonus-0.5) > 1e-9 || conductBonus != 0 {
		t.Fatalf("swirl bonus %v conduct bonus %v", swirlBonus, conductBonus)
	}

	trg[0].AttachOrRefill(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25, ActorIndex: qiqi.Index()},
	})
	trg[0].React(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Electro, Durability: 25, ActorIndex: yae.Index()},
	})
	if !qiqi.StatusIsActive(reactable.PolestarFieldKey) {
		t.Fatal("polestar status")
	}
	swirlBonus = qiqi.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStellarSwirl})
	conductBonus = qiqi.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagDirectStellarConduct})
	superBonus := qiqi.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagSuperconductDamage})
	if swirlBonus != 0 || math.Abs(conductBonus-0.5) > 1e-9 || math.Abs(superBonus-0.5) > 1e-9 {
		t.Fatalf("after polestar swirl %v conduct %v super %v", swirlBonus, conductBonus, superBonus)
	}
}

func TestQiqiWithoutRevelationDoesNotEnter(t *testing.T) {
	c, trg := makeCore()
	addChar(t, c, keys.Sandrone, 0, nil)
	qiqi := addChar(t, c, keys.Qiqi, 0, map[string]int{"revelation": 0})
	mustInit(t, c)
	c.Player.SetActive(qiqi.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Qiqi, nil); err != nil {
		t.Fatal(err)
	}
	swirlCryo(trg[0], qiqi.Index(), qiqi.Index())
	if qiqi.StatusIsActive(radianceKey) {
		t.Fatal("qiqi entered radiance without revelation")
	}
	if qiqi.ReactBonus(info.AttackInfo{AttackTag: attacks.AttackTagReactionStellarSwirl}) != 0 {
		t.Fatal("react bonus without revelation")
	}
}

func TestVodyanitsaExtendsEnterOnly(t *testing.T) {
	c, trg := makeCore()
	sand := addChar(t, c, keys.Sandrone, 0, nil)
	vody := addChar(t, c, keys.Vodyanitsa, 0, nil)
	mustInit(t, c)
	c.Player.SetActive(vody.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Vodyanitsa, nil); err != nil {
		t.Fatal(err)
	}
	if c.Flags.Custom[reactable.VodyanitsaSongUntilKey] <= float64(c.F) {
		t.Fatal("song flag")
	}
	swirlCryo(trg[0], sand.Index(), sand.Index())
	if sand.StatusDuration(radianceKey) != reactable.RadianceStellarSwirlBase+reactable.RadianceStellarSwirlExtend {
		t.Fatalf("enter duration %d", sand.StatusDuration(radianceKey))
	}
	if vody.StatusIsActive(radianceKey) {
		t.Fatal("vodyanitsa entered swirl radiance")
	}
	swirlCryo(trg[0], sand.Index(), sand.Index())
	if sand.StatusDuration(radianceKey) != reactable.RadianceStellarSwirlBase {
		t.Fatalf("refresh duration %d", sand.StatusDuration(radianceKey))
	}
}

func TestCryoTravelerEntersRadiance(t *testing.T) {
	c, trg := makeCore()
	trv := addChar(t, c, keys.LumineCryo, 8000, nil)
	sucrose := addChar(t, c, keys.Sucrose, 0, nil)
	mustInit(t, c)
	if c.Flags.Custom[reactable.StellarSwirlEnableKey] != 1 || c.Flags.Custom[reactable.StellarConductEnableKey] != 1 {
		t.Fatal("traveler enable flags")
	}
	var bonus float64
	c.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.AttackTag == attacks.AttackTagReactionStellarSwirl {
			bonus = atk.Info.BaseDmgBonus
		}
	}, "trav-bonus")
	swirlCryo(trg[0], trv.Index(), sucrose.Index())
	if trv.StatusDuration(radianceKey) != reactable.RadianceStellarSwirlBase {
		t.Fatalf("traveler radiance %d", trv.StatusDuration(radianceKey))
	}
	if sucrose.StatusIsActive(radianceKey) {
		t.Fatal("sucrose entered radiance")
	}
	if math.Abs(bonus-0.07) > 1e-9 {
		t.Fatalf("traveler bonus %v", bonus)
	}
}

func TestVesnaEnablesSwirlOnly(t *testing.T) {
	c, trg := makeCore()
	ves := addChar(t, c, keys.Vesna, 8000, nil)
	sucrose := addChar(t, c, keys.Sucrose, 0, nil)
	mustInit(t, c)
	if c.Flags.Custom[reactable.StellarSwirlEnableKey] != 1 {
		t.Fatal("vesna swirl flag")
	}
	if _, ok := c.Flags.Custom[reactable.StellarConductEnableKey]; ok {
		t.Fatal("vesna enabled stellar conduct")
	}
	swirlCryo(trg[0], ves.Index(), sucrose.Index())
	if ves.StatusDuration("vesna-radiance") != reactable.RadianceStellarSwirlBase {
		t.Fatalf("vesna radiance %d", ves.StatusDuration("vesna-radiance"))
	}
	if sucrose.StatusIsActive("vesna-radiance") || sucrose.StatusIsActive(radianceKey) {
		t.Fatal("vesna radiance spread")
	}
}

func TestPolestarPrecedesSwirlRadiance(t *testing.T) {
	c, trg := makeCore()
	sand := addChar(t, c, keys.Sandrone, 0, nil)
	yae := addChar(t, c, keys.YaeMiko, 0, nil)
	mustInit(t, c)
	trg[0].AttachOrRefill(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25, ActorIndex: sand.Index()},
	})
	var conduct, plain int
	c.Events.Subscribe(event.OnStellarConduct, func(...any) { conduct++ }, "polestar-conduct")
	c.Events.Subscribe(event.OnSuperconduct, func(...any) { plain++ }, "polestar-super")
	trg[0].React(&info.AttackEvent{
		Info: info.AttackInfo{Element: attributes.Electro, Durability: 25, ActorIndex: yae.Index()},
	})
	if conduct != 1 || plain != 0 {
		t.Fatalf("conduct %d superconduct %d", conduct, plain)
	}
	if !sand.StatusIsActive(reactable.PolestarFieldKey) {
		t.Fatal("polestar status missing")
	}
	if asSandrone(sand).getRadiance() != radianceStellarConduct {
		t.Fatal("radiance before swirl")
	}
	swirlCryo(trg[0], sand.Index(), sand.Index())
	if !sand.StatusIsActive(radianceKey) {
		t.Fatal("swirl radiance missing")
	}
	if asSandrone(sand).getRadiance() != radianceStellarConduct {
		t.Fatal("swirl radiance replaced polestar")
	}
}

func TestWriothesleyAndAlyoshaDoNotEnterSwirlRadiance(t *testing.T) {
	c, trg := makeCore()
	sand := addChar(t, c, keys.Sandrone, 0, nil)
	wrio := addChar(t, c, keys.Wriothesley, 0, nil)
	aly := addChar(t, c, keys.Alyosha, 0, nil)
	mustInit(t, c)
	if c.Flags.Custom[reactable.StellarSwirlEnableKey] != 1 || c.Flags.Custom[reactable.StellarConductEnableKey] != 1 {
		t.Fatal("enable flags")
	}
	swirlCryo(trg[0], sand.Index(), sand.Index())
	if !sand.StatusIsActive(radianceKey) {
		t.Fatal("sandrone radiance")
	}
	if wrio.StatusIsActive(radianceKey) || aly.StatusIsActive(radianceKey) {
		t.Fatal("swirl radiance applied without a character passive")
	}
	if wrio.StatusIsActive("vesna-radiance") || aly.StatusIsActive("vesna-radiance") {
		t.Fatal("unexpected radiance key")
	}
}
