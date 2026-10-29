package vesna

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/bennett"
	_ "github.com/genshinsim/gcsim/internal/characters/kaeya"
	_ "github.com/genshinsim/gcsim/internal/characters/xiangling"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"github.com/genshinsim/gcsim/pkg/reactable"
	"github.com/genshinsim/gcsim/pkg/testhelper"
)

func init() {
	testhelper.RegisterTestWeapon()
}

func makeCore() (*core.Core, []*enemy.Enemy) {
	c, err := core.New(core.Opt{Seed: 1, Debug: false})
	if err != nil {
		panic(err)
	}
	a := avatar.New(c, info.Point{X: 0, Y: 0}, 1)
	c.Combat.SetPlayer(a)
	var trgs []*enemy.Enemy
	for range 1 {
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

func addProfile(t *testing.T, c *core.Core, key keys.Char, cons int) *character.CharWrapper {
	t.Helper()
	return addProfileAt(t, c, key, cons, 90, 90)
}

func addProfileAt(t *testing.T, c *core.Core, key keys.Char, cons, level, maxLevel int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(key, testhelper.TestWeaponKey)
	prof.Base.Cons = cons
	prof.Base.Level = level
	prof.Base.MaxLevel = maxLevel
	prof.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("add %v: %v", key, err)
	}
	return c.Player.ByIndex(idx)
}

func asVesna(ch *character.CharWrapper) *char {
	return ch.Character.(*char)
}

func mustExec(t *testing.T, c *core.Core, a action.Action, k keys.Char) {
	t.Helper()
	if err := c.Player.Exec(a, k, nil); err != nil {
		t.Fatalf("%v: %v", a, err)
	}
}

func near(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

func condNum(t *testing.T, ch *character.CharWrapper, key string) float64 {
	t.Helper()
	v, err := ch.Condition([]string{key})
	if err != nil {
		t.Fatal(err)
	}
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		t.Fatalf("%s type %T", key, v)
		return 0
	}
}

func forceCrit(ch *character.CharWrapper, cr, anemo float64) {
	m := make([]float64, attributes.EndStatType)
	m[attributes.CR] = cr
	m[attributes.AnemoP] = anemo
	ch.AddStatMod(character.StatMod{
		Base: modifier.NewBase("vesna-test-stats", -1),
		Amount: func() []float64 {
			return m
		},
	})
}

func TestIdentity(t *testing.T) {
	c, _ := makeCore()
	ch := addProfile(t, c, keys.Vesna, 0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	if ch.Base.Element != attributes.Anemo || ch.Weapon.Class != info.WeaponClassSword || ch.EnergyMax != 60 {
		t.Fatalf("identity ele %v weapon %v energy %v", ch.Base.Element, ch.Weapon.Class, ch.EnergyMax)
	}
	if ch.Base.Ascension != 6 {
		t.Fatalf("ascension %d", ch.Base.Ascension)
	}
	if math.Abs(ch.BaseStats[attributes.CR]-(0.05+0.192)) > 1e-4 {
		t.Fatalf("cr %v", ch.BaseStats[attributes.CR])
	}
	// Yatta level 90 display, before the weapon's attack is added.
	if math.Abs(ch.BaseStats[attributes.BaseHP]-13262) > 2 {
		t.Fatalf("hp %v", ch.BaseStats[attributes.BaseHP])
	}
	if math.Abs(ch.BaseStats[attributes.BaseDEF]-730) > 2 {
		t.Fatalf("def %v", ch.BaseStats[attributes.BaseDEF])
	}
	if _, ok := c.Flags.Custom[reactable.StellarSwirlEnableKey]; !ok {
		t.Fatal("stellar swirl flag")
	}
	if _, ok := c.Flags.Custom[reactable.StellarConductEnableKey]; ok {
		t.Fatal("vesna must not enable stellar conduct")
	}
	if ch.TalentLvlSkill() != 0 || ch.TalentLvlBurst() != 0 {
		t.Fatalf("talent index skill %d burst %d", ch.TalentLvlSkill(), ch.TalentLvlBurst())
	}
}

func TestTalentCons(t *testing.T) {
	c, _ := makeCore()
	ch := addProfile(t, c, keys.Vesna, 6)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	// Talent 1 plus C3/C5 is index 3, the level-4 row.
	if ch.TalentLvlSkill() != 3 || ch.TalentLvlBurst() != 3 {
		t.Fatalf("c6 indexes skill %d burst %d", ch.TalentLvlSkill(), ch.TalentLvlBurst())
	}
	if !near(skillDMG[ch.TalentLvlSkill()], 0.5) || !near(burstDMG[ch.TalentLvlBurst()], 3.29) {
		t.Fatalf("raised ratios skill %v burst %v", skillDMG[ch.TalentLvlSkill()], burstDMG[ch.TalentLvlBurst()])
	}
}

func TestArmedInfusionAndParticles(t *testing.T) {
	c, trg := makeCore()
	ch := addProfile(t, c, keys.Vesna, 0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	ch.ParticleDelay = 0
	c.Player.SetActive(ch.Index())
	var particles []character.Particle
	c.Events.Subscribe(event.OnParticleReceived, func(args ...any) {
		particles = append(particles, args[0].(character.Particle))
	}, "particles")
	var hits []info.AttackInfo
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.ActorIndex == ch.Index() {
			hits = append(hits, ai)
		}
	}, "hits")

	mustExec(t, c, action.ActionAttack, keys.Vesna)
	advance(c, attackHitmarks[0][0]+2)
	if len(hits) != 1 || hits[0].Element != attributes.Physical || hits[0].IgnoreInfusion {
		t.Fatalf("physical normal %+v", hits)
	}
	if trg[0].AuraContains(attributes.Anemo) || len(particles) != 0 {
		t.Fatalf("outside armed aura %v particles %d", trg[0].AuraContains(attributes.Anemo), len(particles))
	}

	c.Player.AddWeaponInfuse(ch.Index(), "test-pyro", attributes.Pyro, 600, true, attacks.AttackTagNormal, attacks.AttackTagExtra, attacks.AttackTagPlunge)
	asVesna(ch).ResetNormalCounter()
	hits = nil
	mustExec(t, c, action.ActionAttack, keys.Vesna)
	advance(c, attackHitmarks[0][0]+2)
	if len(hits) != 1 || hits[0].Element != attributes.Pyro || hits[0].IgnoreInfusion {
		t.Fatalf("infused normal %+v", hits)
	}

	hits = nil
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	if condNum(t, ch, "armed") != 1 || !near(condNum(t, ch, "essence"), essenceSkill) || condNum(t, ch, "windborne") != 1 {
		t.Fatalf("enter armed %v essence %v level %v", condNum(t, ch, "armed"), condNum(t, ch, "essence"), condNum(t, ch, "windborne"))
	}
	cd := ch.Cooldown(action.ActionSkill)
	if cd <= 0 {
		t.Fatal("entry skill did not start the cooldown")
	}
	ready, _ := ch.ActionReady(action.ActionSkill, nil)
	if !ready {
		t.Fatal("lv1 should be ready from the entry essence")
	}
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	if ch.Cooldown(action.ActionSkill) != cd {
		t.Fatal("windborne refreshed the entry cooldown")
	}
	advance(c, skillHitmark+2)
	var armedNormal, skillHit, pinion int
	for _, ai := range hits {
		switch ai.Abil {
		case "The Art of Victory":
			skillHit++
			if ai.Element != attributes.Anemo || ai.Durability != 25 || !near(ai.Mult, skillDMG[0]) {
				t.Fatalf("skill hit %+v", ai)
			}
		case "Windborne Sword Lv1":
			if ai.Element != attributes.Anemo || !near(ai.Mult, skillDMG[0]) || ai.AttackTag != attacks.AttackTagElementalArt {
				t.Fatalf("lv1 %+v", ai)
			}
		case "Wind Pinion":
			pinion++
		case "Armed Normal 0", "Armed Normal 1":
			armedNormal++
			if ai.Element != attributes.Anemo || !ai.IgnoreInfusion {
				t.Fatalf("armed normal %+v", ai)
			}
		}
	}
	if skillHit != 1 || pinion != 0 {
		t.Fatalf("skill %d pinion %d hits %d", skillHit, pinion, len(hits))
	}
	if len(particles) != 1 || particles[0].Ele != attributes.Anemo || particles[0].Num != 1 {
		t.Fatalf("particles %+v", particles)
	}
	// The infused normal left Pyro. The Anemo skill swirls it, so Anemo does not remain.

	hits = nil
	mustExec(t, c, action.ActionAttack, keys.Vesna)
	advance(c, attackHitmarks[0][0]+2)
	for _, ai := range hits {
		if ai.Abil == "Armed Normal 0" || ai.Abil == "Armed Normal 1" || ai.Abil == "Armed Normal 2" {
			if ai.Element != attributes.Anemo || !ai.IgnoreInfusion {
				t.Fatalf("infusion overrode armed %+v", ai)
			}
			armedNormal++
		}
	}
	if armedNormal == 0 {
		t.Fatal("no armed normal")
	}
	// The normal itself is not in the particle list. Its pinion is, and the ICD is still up.
	if len(particles) != 1 {
		t.Fatalf("particle icd broken, count %d", len(particles))
	}
}

func TestNaturalWindborne(t *testing.T) {
	c, _ := makeCore()
	ch := addProfile(t, c, keys.Vesna, 0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	var blades, pinions int
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		switch ai.Abil {
		case "Spirit Blade Lv2", "Spirit Blade Lv3", "Spirit Blade Lv3 Final":
			blades++
			if ai.AttackTag != attacks.AttackTagElementalArt || ai.Durability != 25 || ai.IgnoreDefPercent != 0 {
				t.Fatalf("ordinary blade %+v", ai)
			}
		case "Wind Pinion":
			pinions++
		}
	}, "hits")

	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	if !near(condNum(t, ch, "essence"), 30) || condNum(t, ch, "windborne") != 2 || condNum(t, ch, "discipline") != 1 {
		t.Fatalf("after lv1 essence %v level %v disc %v", condNum(t, ch, "essence"), condNum(t, ch, "windborne"), condNum(t, ch, "discipline"))
	}
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	if !near(condNum(t, ch, "essence"), 0) || condNum(t, ch, "windborne") != 3 || condNum(t, ch, "discipline") != 2 {
		t.Fatalf("after lv2 essence %v level %v disc %v", condNum(t, ch, "essence"), condNum(t, ch, "windborne"), condNum(t, ch, "discipline"))
	}
	ready, fail := ch.ActionReady(action.ActionSkill, nil)
	if ready || fail != action.SkillCD {
		t.Fatalf("lv3 without essence ready %v fail %v", ready, fail)
	}
	if err := c.Player.Exec(action.ActionSkill, keys.Vesna, nil); err == nil {
		t.Fatal("lv3 cast with no essence")
	}
	advance(c, lv2Blade+2)
	if blades != 1 || pinions != 0 {
		t.Fatalf("blades %d pinions %d", blades, pinions)
	}

	for range 6 {
		mustExec(t, c, action.ActionAttack, keys.Vesna)
	}
	if !near(condNum(t, ch, "essence"), 30) {
		t.Fatalf("pinions gave essence %v", condNum(t, ch, "essence"))
	}
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	if !near(condNum(t, ch, "essence"), 0) || condNum(t, ch, "windborne") != 3 {
		t.Fatalf("lv3 essence %v level %v", condNum(t, ch, "essence"), condNum(t, ch, "windborne"))
	}
	advance(c, lv3Hitmarks[len(lv3Hitmarks)-1]+2)
	if blades != 6 {
		t.Fatalf("blades after lv3 %d", blades)
	}
	advance(c, pinionHitmark+2)
	if pinions != 6 {
		t.Fatalf("pinions %d", pinions)
	}
}

func TestEssenceCapAndBurst(t *testing.T) {
	c, _ := makeCore()
	ch := addProfile(t, c, keys.Vesna, 0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	// The config does not start with a full burst bar. This test fills it on purpose.
	ch.AddEnergy("test", ch.EnergyMax)
	var blades []info.AttackInfo
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.Abil == "For the Tsaritsa!" {
			blades = append(blades, ai)
		}
	}, "burst")

	mustExec(t, c, action.ActionBurst, keys.Vesna)
	if condNum(t, ch, "armed") != 0 || !near(condNum(t, ch, "essence"), essenceBurst) || condNum(t, ch, "discipline") != 1 {
		t.Fatalf("burst outside armed %v essence %v disc %v", condNum(t, ch, "armed"), condNum(t, ch, "essence"), condNum(t, ch, "discipline"))
	}
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	// Entry clears the burst's disc, then adds 60 onto the 30 that was already stored.
	if !near(condNum(t, ch, "essence"), essenceMax) || condNum(t, ch, "discipline") != 0 {
		t.Fatalf("capped essence %v disc %v", condNum(t, ch, "essence"), condNum(t, ch, "discipline"))
	}
	ch.AddEnergy("test", ch.EnergyMax)
	ch.ResetActionCooldown(action.ActionBurst)
	mustExec(t, c, action.ActionBurst, keys.Vesna)
	if !near(condNum(t, ch, "essence"), essenceMax) || condNum(t, ch, "discipline") != 1 {
		t.Fatalf("burst at cap essence %v disc %v", condNum(t, ch, "essence"), condNum(t, ch, "discipline"))
	}
	for range 6 {
		mustExec(t, c, action.ActionAttack, keys.Vesna)
	}
	if !near(condNum(t, ch, "essence"), essenceMax) {
		t.Fatalf("pinions broke the cap %v", condNum(t, ch, "essence"))
	}
	advance(c, burstBladeDelay+2)
	if len(blades) != 2 {
		t.Fatalf("burst blades %d", len(blades))
	}
	want := burstDMG[0] * (1 + discPerStack)
	if !near(blades[1].Mult, want) || blades[1].AttackTag != attacks.AttackTagElementalBurst || blades[1].Durability != 25 {
		t.Fatalf("burst blade %+v want %v", blades[1], want)
	}
}

func TestC1UsesAndDuration(t *testing.T) {
	c, _ := makeCore()
	ch := addProfile(t, c, keys.Vesna, 1)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	vc := asVesna(ch)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	// First Lv3 is free. Essence was already 0 after Lv2, and the flag is spent.
	if vc.c1Free || !near(vc.essence, 0) || vc.e3Uses != 1 || vc.closing {
		t.Fatalf("first lv3 free %v essence %v uses %d closing %v", vc.c1Free, vc.essence, vc.e3Uses, vc.closing)
	}
	for use := 2; use <= e3C1; use++ {
		for range 6 {
			mustExec(t, c, action.ActionAttack, keys.Vesna)
		}
		if !near(vc.essence, essenceCost) {
			t.Fatalf("use %d essence %v", use, vc.essence)
		}
		mustExec(t, c, action.ActionSkill, keys.Vesna)
		if vc.e3Uses != use || !near(vc.essence, 0) {
			t.Fatalf("use %d recorded %d essence %v", use, vc.e3Uses, vc.essence)
		}
	}
	if !vc.closing {
		t.Fatal("fourth lv3 did not close windborne")
	}
	ready, _ := ch.ActionReady(action.ActionSkill, nil)
	if ready {
		t.Fatal("windborne stayed ready while closing")
	}
	if condNum(t, ch, "armed") != 1 {
		t.Fatal("armed ended before the placeholder animation")
	}
	advance(c, windborneEnd)
	if condNum(t, ch, "armed") != 0 || !near(condNum(t, ch, "essence"), 0) {
		t.Fatalf("after close armed %v essence %v", condNum(t, ch, "armed"), condNum(t, ch, "essence"))
	}

	// A fresh entry still lasts 15s when it is not closed by Lv3.
	// The 18s cooldown is still running, so this test clears it.
	ch.ResetActionCooldown(action.ActionSkill)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	advance(c, armedDuration-1)
	if condNum(t, ch, "armed") != 1 {
		t.Fatal("armed ended early")
	}
	advance(c, 1)
	if condNum(t, ch, "armed") != 0 || !near(condNum(t, ch, "essence"), 0) {
		t.Fatalf("duration exit armed %v essence %v", condNum(t, ch, "armed"), condNum(t, ch, "essence"))
	}
}

func TestSwapResets(t *testing.T) {
	c, _ := makeCore()
	ch := addProfile(t, c, keys.Vesna, 6)
	addProfile(t, c, keys.Kaeya, 0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	if condNum(t, ch, "transpose") != 1 || condNum(t, ch, "discipline") != discMax {
		t.Fatalf("before swap transpose %v disc %v", condNum(t, ch, "transpose"), condNum(t, ch, "discipline"))
	}
	mustExec(t, c, action.ActionSwap, keys.Kaeya)
	advance(c, 2)
	if condNum(t, ch, "armed") != 0 || !near(condNum(t, ch, "essence"), 0) || condNum(t, ch, "discipline") != 0 || condNum(t, ch, "transpose") != 0 {
		t.Fatalf("after swap armed %v essence %v disc %v transpose %v", condNum(t, ch, "armed"), condNum(t, ch, "essence"), condNum(t, ch, "discipline"), condNum(t, ch, "transpose"))
	}
	mustExec(t, c, action.ActionSwap, keys.Vesna)
	advance(c, 2)
	if condNum(t, ch, "armed") != 0 {
		t.Fatal("swap back resumed armed")
	}
}

func TestC2AndTruth(t *testing.T) {
	c, _ := makeCore()
	ch := addProfile(t, c, keys.Vesna, 2)
	kaeya := addProfile(t, c, keys.Kaeya, 0)
	ben := addProfile(t, c, keys.Bennett, 0)
	xl := addProfile(t, c, keys.Xiangling, 0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	before := ch.Stat(attributes.ATKP)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	if condNum(t, ch, "discipline") != discMax {
		t.Fatalf("c2 discs %v", condNum(t, ch, "discipline"))
	}
	if math.Abs(ch.Stat(attributes.ATKP)-(before+c2ATK)) > 1e-9 {
		t.Fatalf("c2 atk %v before %v", ch.Stat(attributes.ATKP), before)
	}
	advance(c, armedDuration)
	if condNum(t, ch, "armed") != 0 || condNum(t, ch, "discipline") != discMax {
		t.Fatalf("discs cleared with armed, armed %v disc %v", condNum(t, ch, "armed"), condNum(t, ch, "discipline"))
	}
	if math.Abs(ch.Stat(attributes.ATKP)-(before+c2ATK)) > 1e-9 {
		t.Fatal("c2 atk dropped while six discs remained")
	}
	advance(c, discDuration-armedDuration)
	if condNum(t, ch, "discipline") != 0 || math.Abs(ch.Stat(attributes.ATKP)-before) > 1e-9 {
		t.Fatalf("after disc expiry disc %v atk %v", condNum(t, ch, "discipline"), ch.Stat(attributes.ATKP))
	}
	_ = kaeya
	_ = ben
	_ = xl

	c2, _ := makeCore()
	low := addProfileAt(t, c2, keys.Vesna, 2, 1, 20)
	if err := c2.Init(); err != nil {
		t.Fatal(err)
	}
	c2.Player.SetActive(low.Index())
	mustExec(t, c2, action.ActionSkill, keys.Vesna)
	if condNum(t, low, "discipline") != 0 || low.Stat(attributes.ATKP) != 0 {
		t.Fatalf("c2 without a1 disc %v atk %v", condNum(t, low, "discipline"), low.Stat(attributes.ATKP))
	}

	c4, trg := makeCore()
	main := addProfile(t, c4, keys.Vesna, 4)
	addProfile(t, c4, keys.Kaeya, 0)
	addProfile(t, c4, keys.Bennett, 0)
	addProfile(t, c4, keys.Xiangling, 0)
	if err := c4.Init(); err != nil {
		t.Fatal(err)
	}
	c4.Player.SetActive(main.Index())
	mustExec(t, c4, action.ActionSkill, keys.Vesna)
	atkArmed := main.Stat(attributes.ATKP)
	emArmed := main.Stat(attributes.EM)
	trg[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25}})
	advance(c4, 1)
	// The entry hit was queued at frame 14 and has already passed. Apply a fresh Anemo hit.
	c4.QueueAttack(info.AttackInfo{
		ActorIndex: main.Index(),
		Abil:       "anemo-gauge",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		Element:    attributes.Anemo,
		Durability: 25,
	}, combat.NewSingleTargetHit(trg[0].Key()), 0, 0)
	if condNum(t, main, "radiance") != 1 {
		t.Fatal("radiance did not start")
	}
	// Two Cryo/Anemo (Vesna, Kaeya) and two others. C4 triples 6% and 25.
	if math.Abs(main.Stat(attributes.ATKP)-(atkArmed+a4ATK*2*c4Mult)) > 1e-9 {
		t.Fatalf("c4 atk %v from %v", main.Stat(attributes.ATKP), atkArmed)
	}
	if math.Abs(main.Stat(attributes.EM)-(emArmed+a4EM*2*c4Mult)) > 1e-6 {
		t.Fatalf("c4 em %v from %v", main.Stat(attributes.EM), emArmed)
	}

	c0, trg0 := makeCore()
	plain := addProfile(t, c0, keys.Vesna, 0)
	addProfile(t, c0, keys.Kaeya, 0)
	addProfile(t, c0, keys.Bennett, 0)
	addProfile(t, c0, keys.Xiangling, 0)
	if err := c0.Init(); err != nil {
		t.Fatal(err)
	}
	c0.Player.SetActive(plain.Index())
	em0 := plain.Stat(attributes.EM)
	atk0 := plain.Stat(attributes.ATKP)
	trg0[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25}})
	advance(c0, 1)
	c0.QueueAttack(info.AttackInfo{
		ActorIndex: plain.Index(),
		Abil:       "anemo-gauge",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		Element:    attributes.Anemo,
		Durability: 25,
	}, combat.NewSingleTargetHit(trg0[0].Key()), 0, 0)
	if math.Abs(plain.Stat(attributes.ATKP)-(atk0+a4ATK*2)) > 1e-9 || math.Abs(plain.Stat(attributes.EM)-(em0+a4EM*2)) > 1e-6 {
		t.Fatalf("c0 truth atk %v em %v", plain.Stat(attributes.ATKP), plain.Stat(attributes.EM))
	}

	low4, trg4 := makeCore()
	early := addProfileAt(t, low4, keys.Vesna, 4, 40, 50)
	if err := low4.Init(); err != nil {
		t.Fatal(err)
	}
	low4.Player.SetActive(early.Index())
	beforeEarly := early.Stat(attributes.ATKP)
	trg4[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25}})
	advance(low4, 1)
	low4.QueueAttack(info.AttackInfo{
		ActorIndex: early.Index(),
		Abil:       "anemo-gauge",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		Element:    attributes.Anemo,
		Durability: 25,
	}, combat.NewSingleTargetHit(trg4[0].Key()), 0, 0)
	if condNum(t, early, "radiance") != 1 || math.Abs(early.Stat(attributes.ATKP)-beforeEarly) > 1e-9 {
		t.Fatalf("truth below a4 radiance %v atk %v before %v asc %d", condNum(t, early, "radiance"), early.Stat(attributes.ATKP), beforeEarly, early.Base.Ascension)
	}
}

func TestC6Transpose(t *testing.T) {
	c, _ := makeCore()
	ch := addProfile(t, c, keys.Vesna, 6)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	vc := asVesna(ch)
	var hits []info.AttackInfo
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.ActorIndex == ch.Index() {
			hits = append(hits, ai)
		}
	}, "hits")

	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	if condNum(t, ch, "transpose") != 1 || vc.e3Uses != 1 || !near(vc.essence, 0) {
		t.Fatalf("window %v uses %d essence %v", condNum(t, ch, "transpose"), vc.e3Uses, vc.essence)
	}
	uses := vc.e3Uses
	discs := vc.discCount()
	mustExec(t, c, action.ActionAttack, keys.Vesna)
	if condNum(t, ch, "transpose") != 0 || vc.e3Uses != uses || vc.discCount() != discs {
		t.Fatalf("transpose consumed wrong uses %d disc %d window %v", vc.e3Uses, vc.discCount(), condNum(t, ch, "transpose"))
	}
	if !near(vc.essence, essencePinion) {
		t.Fatalf("armed transpose essence %v", vc.essence)
	}
	advance(c, transposeBladeDelay+2)
	var slash, blade, pinion int
	for _, ai := range hits {
		switch ai.Abil {
		case "Windborne Sword: Transpose":
			slash++
			if !near(ai.Mult, transposeTalent) || ai.AttackTag != attacks.AttackTagElementalArt || ai.Durability != 25 {
				t.Fatalf("transpose slash %+v", ai)
			}
		case "Transpose Spirit Blade":
			blade++
			if !near(ai.Mult, transposeBlade*(1+discPerStack*float64(discMax))) || ai.Durability != 25 {
				t.Fatalf("transpose blade %+v", ai)
			}
		case "Wind Pinion":
			pinion++
		}
	}
	if slash != 1 || blade != 1 || pinion != 1 {
		t.Fatalf("transpose hits slash %d blade %d pinion %d", slash, blade, pinion)
	}

	hits = nil
	mustExec(t, c, action.ActionAttack, keys.Vesna)
	advance(c, attackHitmarks[0][0]+2)
	for _, ai := range hits {
		if ai.Abil == "Windborne Sword: Transpose" {
			t.Fatal("window repeated")
		}
	}

	// The capping Lv3 ends Armed at the placeholder, and the 5s window stays up.
	for use := vc.e3Uses + 1; use <= e3C1; use++ {
		for vc.essence < essenceCost {
			mustExec(t, c, action.ActionAttack, keys.Vesna)
		}
		mustExec(t, c, action.ActionSkill, keys.Vesna)
	}
	advance(c, windborneEnd)
	if condNum(t, ch, "armed") != 0 || condNum(t, ch, "transpose") != 1 || !near(condNum(t, ch, "essence"), 0) {
		t.Fatalf("after final lv3 armed %v window %v essence %v", condNum(t, ch, "armed"), condNum(t, ch, "transpose"), condNum(t, ch, "essence"))
	}
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	if !near(condNum(t, ch, "essence"), 0) || condNum(t, ch, "transpose") != 0 {
		t.Fatalf("unarmed transpose essence %v window %v", condNum(t, ch, "essence"), condNum(t, ch, "transpose"))
	}
}

func TestStellarTagsAndFormula(t *testing.T) {
	c, trg := makeCore()
	ch := addProfile(t, c, keys.Vesna, 0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	ch.ParticleDelay = 0
	forceCrit(ch, -1, 1)
	c.Player.SetActive(ch.Index())
	trg[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Cryo, Durability: 50}})

	var ordinary, stellar float64
	var stellarAI info.AttackInfo
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		amt := args[2].(float64)
		switch atk.Info.Abil {
		case "The Art of Victory":
			ordinary = amt
			want := talentDamage(atk, 100)
			if math.Abs(amt-want) > 1e-4 {
				t.Fatalf("ordinary %v want %v anemo %v", amt, want, atk.Snapshot.Stats[attributes.AnemoP])
			}
		case "Spirit Blade Lv2":
			stellar = amt
			stellarAI = atk.Info
			want := directDamage(atk, ch.ReactBonus(atk.Info))
			if math.Abs(amt-want) > 1e-4 {
				t.Fatalf("stellar %v want %v bonus %v elev %v", amt, want, atk.Info.BaseDmgBonus, atk.Info.Elevation)
			}
		}
	}, "dmg")

	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	advance(c, lv2Blade+2)
	if ordinary == 0 || stellar == 0 {
		t.Fatalf("missing damage ordinary %v stellar %v", ordinary, stellar)
	}
	if stellarAI.AttackTag != attacks.AttackTagDirectStellarSwirl || stellarAI.Durability != 0 || stellarAI.IgnoreDefPercent != 1 {
		t.Fatalf("stellar tags %+v", stellarAI)
	}
	if stellarAI.Elevation != 0 {
		t.Fatalf("c0 elevation %v", stellarAI.Elevation)
	}
	bonus := asVesna(ch).jubileeBonus()
	if math.Abs(stellarAI.BaseDmgBonus-bonus) > 1e-9 || bonus <= 0 || bonus > jubileeCap {
		t.Fatalf("jubilee %v info %v", bonus, stellarAI.BaseDmgBonus)
	}
	// Same talent percentage as the ordinary row, with one disc from each of Lv1 and Lv2.
	if !near(stellarAI.Mult, wbLv2Blade[0]*(1+discPerStack*2)) {
		t.Fatalf("disc mult %v", stellarAI.Mult)
	}
	_ = trg
}

func TestC6DirectAndReaction(t *testing.T) {
	c, trg := makeCore()
	ch := addProfileAt(t, c, keys.Vesna, 6, 1, 20)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	forceCrit(ch, -1, 1)
	c.Player.SetActive(ch.Index())
	trg[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Cryo, Durability: 50}})

	var reactElev, landedElev float64
	var reactBonus float64
	var sawReact bool
	c.Events.Subscribe(event.OnSpecialReactionAttack, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		if atk.Info.ActorIndex != ch.Index() || atk.Info.AttackTag != attacks.AttackTagReactionStellarSwirl {
			return
		}
		reactElev = atk.Info.Elevation
		reactBonus = atk.Info.BaseDmgBonus
		sawReact = true
	}, "react")
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		amt := args[2].(float64)
		switch atk.Info.Abil {
		case "The Art of Victory":
			want := talentDamage(atk, 100)
			if math.Abs(amt-want) > 1e-4 {
				t.Fatalf("c6 ordinary picked up stellar terms %v want %v elev %v", amt, want, atk.Info.Elevation)
			}
		case "Spirit Blade Lv2":
			if atk.Info.Elevation != c6Elevation {
				t.Fatalf("direct elevation %v", atk.Info.Elevation)
			}
			want := directDamage(atk, ch.ReactBonus(atk.Info))
			if math.Abs(ch.ReactBonus(atk.Info)-c1React) > 1e-9 {
				t.Fatalf("c1 react %v", ch.ReactBonus(atk.Info))
			}
			if math.Abs(amt-want) > 1e-4 {
				t.Fatalf("c6 stellar %v want %v", amt, want)
			}
		case "Stellar Swirl":
			landedElev = atk.Info.Elevation
		}
	}, "dmg")

	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	advance(c, lv2Blade+2)
	if !sawReact || math.Abs(reactElev-c6Elevation) > 1e-9 {
		t.Fatalf("reaction elevation %v saw %v", reactElev, sawReact)
	}
	if math.Abs(reactBonus-asVesna(ch).jubileeBonus()) > 1e-6 {
		t.Fatalf("reaction jubilee %v", reactBonus)
	}
	if landedElev != 0 {
		t.Fatalf("queued swirl elevation %v", landedElev)
	}
}

func TestSwirlSplit(t *testing.T) {
	c, trg := makeCore()
	ch := addProfile(t, c, keys.Vesna, 0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	var stellar, cryoSwirl, pyroSwirl, conduct, super int
	c.Events.Subscribe(event.OnStellarSwirl, func(args ...any) { stellar++ }, "ss")
	c.Events.Subscribe(event.OnSwirlCryo, func(args ...any) { cryoSwirl++ }, "scryo")
	c.Events.Subscribe(event.OnSwirlPyro, func(args ...any) { pyroSwirl++ }, "spyro")
	c.Events.Subscribe(event.OnStellarConduct, func(args ...any) { conduct++ }, "cond")
	c.Events.Subscribe(event.OnSuperconduct, func(args ...any) { super++ }, "sup")

	trg[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25}})
	advance(c, 1)
	c.QueueAttack(info.AttackInfo{
		ActorIndex: ch.Index(),
		Abil:       "anemo",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		Element:    attributes.Anemo,
		Durability: 25,
	}, combat.NewSingleTargetHit(trg[0].Key()), 0, 0)
	if stellar != 1 || cryoSwirl != 0 || condNum(t, ch, "radiance") != 1 {
		t.Fatalf("cryo stellar %d plain %d radiance %v", stellar, cryoSwirl, condNum(t, ch, "radiance"))
	}

	advance(c, 30)
	trg[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Pyro, Durability: 25}})
	advance(c, 1)
	c.QueueAttack(info.AttackInfo{
		ActorIndex: ch.Index(),
		Abil:       "anemo",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		Element:    attributes.Anemo,
		Durability: 25,
	}, combat.NewSingleTargetHit(trg[0].Key()), 0, 0)
	if pyroSwirl != 1 || stellar != 1 {
		t.Fatalf("pyro swirl %d stellar %d", pyroSwirl, stellar)
	}

	advance(c, 30)
	trg[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Cryo, Durability: 50}})
	advance(c, 1)
	c.QueueAttack(info.AttackInfo{
		ActorIndex: ch.Index(),
		Abil:       "electro",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		Element:    attributes.Electro,
		Durability: 25,
	}, combat.NewSingleTargetHit(trg[0].Key()), 0, 0)
	if super != 1 || conduct != 0 {
		t.Fatalf("superconduct %d conduct %d", super, conduct)
	}
}

func TestPlungeAndCharge(t *testing.T) {
	c, _ := makeCore()
	ch := addProfile(t, c, keys.Vesna, 0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	var hits []info.AttackInfo
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hits = append(hits, args[1].(*info.AttackEvent).Info)
	}, "hits")
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	before := condNum(t, ch, "essence")
	if err := c.Player.SetAirborne(player.AirborneXianyun); err != nil {
		t.Fatal(err)
	}
	if err := c.Player.Exec(action.ActionLowPlunge, keys.Vesna, map[string]int{"collision": 1}); err != nil {
		t.Fatal(err)
	}
	if !near(condNum(t, ch, "essence"), before+essencePinion) {
		t.Fatalf("plunge pinion essence %v", condNum(t, ch, "essence"))
	}
	advance(c, lowPlungeHitmark+2)
	if len(hits) < 2 {
		t.Fatalf("plunge hits %d", len(hits))
	}
	var impact, collisionHit, pinion int
	for _, ai := range hits {
		switch ai.Abil {
		case "Armed Low Plunge":
			impact++
			if ai.Element != attributes.Anemo || !ai.IgnoreInfusion || ai.Durability != 25 || !near(ai.Mult, lowPlunge[0]) {
				t.Fatalf("impact %+v", ai)
			}
		case "Plunge Collision":
			collisionHit++
			if ai.Durability != 0 || ai.Element != attributes.Anemo {
				t.Fatalf("collision %+v", ai)
			}
		case "Wind Pinion":
			pinion++
		}
	}
	if impact != 1 || collisionHit != 1 || pinion != 1 {
		t.Fatalf("impact %d collision %d pinion %d", impact, collisionHit, pinion)
	}

	hits = nil
	before = condNum(t, ch, "essence")
	mustExec(t, c, action.ActionCharge, keys.Vesna)
	advance(c, chargeHitmarks[0]+2)
	var charged int
	for _, ai := range hits {
		if ai.Abil == "Armed Charged Attack" {
			charged++
			if ai.Element != attributes.Anemo || !ai.IgnoreInfusion || !near(ai.Mult, attack7[0]) {
				t.Fatalf("charged %+v", ai)
			}
		}
	}
	if charged != 1 || !near(condNum(t, ch, "essence"), before+essencePinion) {
		t.Fatalf("charged %d essence %v", charged, condNum(t, ch, "essence"))
	}
}

// TestWindborneICD is a mechanics check. Lv3's five blades are queued together.
// Windborne and the ordinary pinion share a 2s counter, first hit only.
// The entry hit has its own counter and still applies inside that window.
func TestWindborneICD(t *testing.T) {
	c, _ := makeCore()
	ch := addProfileAt(t, c, keys.Vesna, 1, 1, 20)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	type landed struct {
		abil string
		dur  info.Durability
		tag  attacks.AttackTag
	}
	var hits []landed
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.ActorIndex != ch.Index() {
			return
		}
		hits = append(hits, landed{ai.Abil, ai.Durability, ai.AttackTag})
	}, "icd")

	// C1 reaches Lv3 with no pinion, so the skill counter is only Windborne hits.
	for range 4 {
		mustExec(t, c, action.ActionSkill, keys.Vesna)
	}
	advance(c, lv3Hitmarks[len(lv3Hitmarks)-1]+2)

	var entry, lv1 bool
	var lv3 int
	for _, h := range hits {
		switch h.abil {
		case "The Art of Victory":
			entry = h.dur > 0 && h.tag == attacks.AttackTagElementalArt
		case "Windborne Sword Lv1":
			lv1 = h.dur > 0 && h.tag == attacks.AttackTagElementalArt
		case "Windborne Sword Lv2", "Spirit Blade Lv2":
			if h.dur > 0 || h.tag != attacks.AttackTagElementalArt {
				t.Fatalf("shared hit applied %+v", h)
			}
		case "Spirit Blade Lv3", "Spirit Blade Lv3 Final":
			lv3++
			if h.dur > 0 || h.tag != attacks.AttackTagElementalArt {
				t.Fatalf("lv3 applied %+v", h)
			}
		}
	}
	if !entry || !lv1 || lv3 != 5 {
		t.Fatalf("entry %v lv1 %v lv3 %d hits %+v", entry, lv1, lv3, hits)
	}

	// The 2s counter armed on Lv1. The reset is timer-1, same as every group.
	resetAt := lv1Hitmark + attacks.ICDGroupResetTimer[attacks.ICDGroupVesnaSkill] - 1
	for c.F < resetAt-pinionHitmark {
		advance(c, 1)
	}
	mustExec(t, c, action.ActionAttack, keys.Vesna)
	advance(c, pinionHitmark+2)
	var pinion bool
	for _, h := range hits {
		if h.abil == "Wind Pinion" {
			pinion = h.dur > 0 && h.tag == attacks.AttackTagElementalArt
		}
	}
	if !pinion {
		t.Fatalf("no application after reset at %d, frame %d", resetAt, c.F)
	}

	cryo, trg := makeCore()
	cryoCh := addProfileAt(t, cryo, keys.Vesna, 1, 1, 20)
	if err := cryo.Init(); err != nil {
		t.Fatal(err)
	}
	cryo.Player.SetActive(cryoCh.Index())
	trg[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Cryo, Durability: 200}})
	var swirls []string
	var blades int
	cryo.Events.Subscribe(event.OnStellarSwirl, func(args ...any) {
		swirls = append(swirls, args[1].(*info.AttackEvent).Info.Abil)
	}, "swirl")
	cryo.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.Abil == "Spirit Blade Lv3" || ai.Abil == "Spirit Blade Lv3 Final" {
			blades++
			if ai.AttackTag != attacks.AttackTagDirectStellarSwirl || ai.Durability != 0 {
				t.Fatalf("radiance blade %+v", ai)
			}
		}
	}, "blades")
	for range 4 {
		mustExec(t, cryo, action.ActionSkill, keys.Vesna)
	}
	advance(cryo, lv3Hitmarks[len(lv3Hitmarks)-1]+2)
	if blades != 5 {
		t.Fatalf("lv3 blades %d", blades)
	}
	for _, name := range swirls {
		if name == "Spirit Blade Lv3" || name == "Spirit Blade Lv3 Final" {
			t.Fatalf("lv3 swirled %v", swirls)
		}
	}
	if len(swirls) != 2 || swirls[0] != "Windborne Sword Lv1" || swirls[1] != "The Art of Victory" {
		t.Fatalf("cryo applications %v", swirls)
	}
}

// TestEntryICD checks the entry hit's own counter: attenuation group "default",
// 2.5s/3-hit, independent of Windborne's 2s counter and of Transpose's standard group.
func TestEntryICD(t *testing.T) {
	if attacks.ICDGroupResetTimer[attacks.ICDGroupVesnaSkillStart] != attacks.ICDGroupResetTimer[attacks.ICDGroupDefault] {
		t.Fatalf("entry reset %d", attacks.ICDGroupResetTimer[attacks.ICDGroupVesnaSkillStart])
	}
	startSeq := attacks.ICDGroupEleApplicationSequence[attacks.ICDGroupVesnaSkillStart]
	defSeq := attacks.ICDGroupEleApplicationSequence[attacks.ICDGroupDefault]
	if len(startSeq) != len(defSeq) {
		t.Fatalf("entry sequence length %d", len(startSeq))
	}
	for i, v := range defSeq {
		if startSeq[i] != v {
			t.Fatalf("entry sequence %v", startSeq)
		}
	}
	if len(attacks.ICDGroupDamageSequence[attacks.ICDGroupVesnaSkillStart]) != len(attacks.ICDGroupDamageSequence[attacks.ICDGroupDefault]) {
		t.Fatal("entry damage sequence length")
	}
	if attacks.ICDGroupResetTimer[attacks.ICDGroupVesnaSkill] != 120 || attacks.ICDGroupEleApplicationSequence[attacks.ICDGroupVesnaSkill][0] != 1 || attacks.ICDGroupEleApplicationSequence[attacks.ICDGroupVesnaSkill][1] != 0 {
		t.Fatal("windborne row changed")
	}

	c, trg := makeCore()
	ch := addProfileAt(t, c, keys.Vesna, 0, 1, 20)
	addProfile(t, c, keys.Kaeya, 0)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	type landed struct {
		frame int
		dur   info.Durability
		group attacks.ICDGroup
	}
	var hits []landed
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		switch ai.Abil {
		case "The Art of Victory", "windborne-probe", "transpose-probe":
			hits = append(hits, landed{c.F, ai.Durability, ai.ICDGroup})
		}
	}, "entry-icd")

	reenter := func() {
		t.Helper()
		if condNum(t, ch, "armed") == 1 {
			mustExec(t, c, action.ActionSwap, keys.Kaeya)
			advance(c, 2)
			if condNum(t, ch, "armed") != 0 {
				t.Fatal("swap did not clear armed")
			}
		}
		ch.ResetActionCooldown(action.ActionSkill)
		if c.Player.Active() != ch.Index() {
			mustExec(t, c, action.ActionSwap, keys.Vesna)
			advance(c, 2)
		}
		mustExec(t, c, action.ActionSkill, keys.Vesna)
		advance(c, skillHitmark+2)
	}
	reenter()
	if len(hits) != 1 || hits[0].dur <= 0 || hits[0].group != attacks.ICDGroupVesnaSkillStart {
		t.Fatalf("first entry %+v", hits)
	}
	// Same frame as the entry application is already past. These two probes use
	// the other counters and must still apply. OnEnemyDamage is a same-frame
	// task, so it is not visible until the next tick.
	c.QueueAttack(info.AttackInfo{
		ActorIndex: ch.Index(),
		Abil:       "windborne-probe",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupVesnaSkill,
		Element:    attributes.Anemo,
		Durability: 25,
	}, combat.NewSingleTargetHit(trg[0].Key()), 0, 0)
	c.QueueAttack(info.AttackInfo{
		ActorIndex: ch.Index(),
		Abil:       "transpose-probe",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		Element:    attributes.Anemo,
		Durability: 25,
	}, combat.NewSingleTargetHit(trg[0].Key()), 0, 0)
	advance(c, 1)
	if len(hits) != 3 || hits[1].dur <= 0 || hits[1].group != attacks.ICDGroupVesnaSkill || hits[2].dur <= 0 || hits[2].group != attacks.ICDGroupDefault {
		t.Fatalf("counters not independent %+v", hits)
	}

	for range 3 {
		reenter()
	}
	entry := make([]landed, 0, 4)
	for _, h := range hits {
		if h.group == attacks.ICDGroupVesnaSkillStart {
			entry = append(entry, h)
		}
	}
	if len(entry) != 4 || entry[0].dur <= 0 || entry[1].dur != 0 || entry[2].dur != 0 || entry[3].dur <= 0 {
		t.Fatalf("3-hit sequence %+v", entry)
	}

	first := entry[0].frame
	reset25 := first + attacks.ICDGroupResetTimer[attacks.ICDGroupVesnaSkillStart] - 1
	reset2 := first + attacks.ICDGroupResetTimer[attacks.ICDGroupVesnaSkill] - 1
	// A 2s reset would already have fired. The 2.5s counter must still be closed.
	target := reset2 + 10
	if target >= reset25 {
		t.Fatalf("no gap between 2s reset %d and 2.5s reset %d", reset2, reset25)
	}
	if condNum(t, ch, "armed") == 1 {
		mustExec(t, c, action.ActionSwap, keys.Kaeya)
		advance(c, 2)
	}
	ch.ResetActionCooldown(action.ActionSkill)
	if c.Player.Active() != ch.Index() {
		mustExec(t, c, action.ActionSwap, keys.Vesna)
		advance(c, 2)
	}
	wantCast := target - skillHitmark
	if c.F > wantCast {
		t.Fatalf("cast frame %d is past %d", c.F, wantCast)
	}
	advance(c, wantCast-c.F)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	advance(c, skillHitmark+2)
	mid := hits[len(hits)-1]
	if mid.frame != target || mid.dur != 0 || mid.group != attacks.ICDGroupVesnaSkillStart {
		t.Fatalf("hit inside 2.5s %+v target %d reset %d", mid, target, reset25)
	}

	if condNum(t, ch, "armed") == 1 {
		mustExec(t, c, action.ActionSwap, keys.Kaeya)
		advance(c, 2)
	}
	ch.ResetActionCooldown(action.ActionSkill)
	if c.Player.Active() != ch.Index() {
		mustExec(t, c, action.ActionSwap, keys.Vesna)
		advance(c, 2)
	}
	wantCast = reset25 + 1 - skillHitmark
	if c.F > wantCast {
		t.Fatalf("reset cast frame %d is past %d", c.F, wantCast)
	}
	advance(c, wantCast-c.F)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	advance(c, skillHitmark+2)
	after := hits[len(hits)-1]
	if after.frame != reset25+1 || after.dur <= 0 {
		t.Fatalf("hit after 2.5s reset %+v reset %d", after, reset25)
	}
}

func TestPinionOrdinaryAndRadiance(t *testing.T) {
	c, trg := makeCore()
	ch := addProfile(t, c, keys.Vesna, 6)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	var hits []info.AttackInfo
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		hits = append(hits, args[1].(*info.AttackEvent).Info)
	}, "pinion")
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	mustExec(t, c, action.ActionAttack, keys.Vesna)
	advance(c, pinionHitmark+2)
	ratio := pinion[ch.TalentLvlSkill()]
	var ordinary info.AttackInfo
	for _, ai := range hits {
		if ai.Abil == "Wind Pinion" {
			ordinary = ai
		}
	}
	if ordinary.AttackTag != attacks.AttackTagElementalArt || ordinary.Durability != 25 || ordinary.ICDGroup != attacks.ICDGroupVesnaSkill || !near(ordinary.Mult, ratio) || ordinary.Elevation != 0 {
		t.Fatalf("ordinary pinion %+v ratio %v", ordinary, ratio)
	}

	trg[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Cryo, Durability: 50}})
	advance(c, 1)
	c.QueueAttack(info.AttackInfo{
		ActorIndex: ch.Index(),
		Abil:       "anemo-gauge",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		Element:    attributes.Anemo,
		Durability: 25,
	}, combat.NewSingleTargetHit(trg[0].Key()), 0, 0)
	if condNum(t, ch, "radiance") != 1 {
		t.Fatal("radiance did not start")
	}
	hits = nil
	mustExec(t, c, action.ActionAttack, keys.Vesna)
	advance(c, pinionHitmark+2)
	var star info.AttackInfo
	for _, ai := range hits {
		if ai.Abil == "Wind Pinion Stellar Swirl" {
			star = ai
		}
	}
	// C2, which C6 includes, grants six stacks on entry. Lv1 refreshes one.
	// The star mixin includes them. The ordinary row does not.
	want := ratio * (1 + discPerStack*float64(discMax))
	if star.AttackTag != attacks.AttackTagDirectStellarSwirl || star.Durability != 0 || star.IgnoreDefPercent != 1 || star.Elevation != c6Elevation || !near(star.Mult, want) {
		t.Fatalf("star pinion %+v want %v", star, want)
	}
}

func TestDiscCapRefreshesOldest(t *testing.T) {
	c, _ := makeCore()
	ch := addProfile(t, c, keys.Vesna, 2)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	if condNum(t, ch, "discipline") != discMax {
		t.Fatalf("c2 discs %v", condNum(t, ch, "discipline"))
	}
	advance(c, 100)
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	if condNum(t, ch, "discipline") != discMax {
		t.Fatalf("refresh kept %v", condNum(t, ch, "discipline"))
	}
	advance(c, discDuration-100)
	// Five untouched stacks have expired. C2's 40% ATK requires all six.
	if condNum(t, ch, "discipline") != 1 || math.Abs(ch.Stat(attributes.ATKP)) > 1e-9 {
		t.Fatalf("oldest refreshed disc %v atk %v", condNum(t, ch, "discipline"), ch.Stat(attributes.ATKP))
	}
	advance(c, 100)
	if condNum(t, ch, "discipline") != 0 {
		t.Fatalf("refreshed stack remained %v", condNum(t, ch, "discipline"))
	}
}

// TestBurstNoICD grants energy and resets the cooldown. It is not a natural rotation.
// The released burst row is 1U with no ICD, and each cast is one damage event.
func TestBurstNoICD(t *testing.T) {
	c, _ := makeCore()
	ch := addProfileAt(t, c, keys.Vesna, 0, 1, 20)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	c.Player.SetActive(ch.Index())
	var durs []info.Durability
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ai := args[1].(*info.AttackEvent).Info
		if ai.Abil == "For the Tsaritsa!" {
			durs = append(durs, ai.Durability)
		}
	}, "burst-icd")
	for range 2 {
		ch.AddEnergy("test", ch.EnergyMax)
		ch.ResetActionCooldown(action.ActionBurst)
		mustExec(t, c, action.ActionBurst, keys.Vesna)
		advance(c, 40)
	}
	advance(c, burstBladeDelay)
	if len(durs) != 2 || durs[0] <= 0 || durs[1] <= 0 {
		t.Fatalf("burst applications %v", durs)
	}
}

// TestBurstFromParticles earns the 60 cost from Vesna's own Anemo particles.
// It is a mechanics check, not a damage benchmark.
func TestBurstFromParticles(t *testing.T) {
	c, _ := makeCore()
	ch := addProfileAt(t, c, keys.Vesna, 0, 1, 20)
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
	ch.ParticleDelay = 0
	// AddChar fills the burst bar. Empty it so only particles count.
	ch.Energy = 0
	c.Player.SetActive(ch.Index())
	mustExec(t, c, action.ActionSkill, keys.Vesna)
	advance(c, skillHitmark+2)
	if ch.Energy <= 0 || ch.Energy >= ch.EnergyMax {
		t.Fatalf("one particle energy %v max %v", ch.Energy, ch.EnergyMax)
	}
	ready, _ := ch.ActionReady(action.ActionBurst, nil)
	if ready {
		t.Fatal("burst ready after one particle")
	}
	for c.F < 90*60 && ch.Energy < ch.EnergyMax {
		if condNum(t, ch, "armed") == 1 {
			if err := c.Player.ReadyCheck(action.ActionAttack, keys.Vesna, nil); err == nil {
				mustExec(t, c, action.ActionAttack, keys.Vesna)
			}
		} else if err := c.Player.ReadyCheck(action.ActionSkill, keys.Vesna, nil); err == nil {
			mustExec(t, c, action.ActionSkill, keys.Vesna)
		}
		advance(c, 1)
	}
	if ch.Energy < ch.EnergyMax {
		t.Fatalf("natural energy %v after frame %d", ch.Energy, c.F)
	}
	var blades int
	c.Events.Subscribe(event.OnEnemyHit, func(args ...any) {
		if args[1].(*info.AttackEvent).Info.Abil == "For the Tsaritsa!" {
			blades++
		}
	}, "burst")
	mustExec(t, c, action.ActionBurst, keys.Vesna)
	advance(c, burstBladeDelay+2)
	if blades != 1 || ch.Energy > 1e-6 {
		t.Fatalf("blades %d energy left %v", blades, ch.Energy)
	}
}

func talentDamage(atk *info.AttackEvent, enemyLvl int) float64 {
	a := atk.Snapshot.Stats.TotalATK()
	dmgP := atk.Snapshot.Stats[attributes.DmgP]
	var ele float64
	switch atk.Info.Element {
	case attributes.Anemo:
		ele = atk.Snapshot.Stats[attributes.AnemoP]
	case attributes.Physical:
		ele = atk.Snapshot.Stats[attributes.PhyP]
	}
	damage := atk.Info.Mult * a * (1 + atk.Info.BaseDmgBonus) * (1 + ele + dmgP)
	defmod := float64(atk.Snapshot.CharLvl+100) / float64(atk.Snapshot.CharLvl+100+enemyLvl+100)
	damage *= defmod
	return applyCrit(damage, atk)
}

func directDamage(atk *info.AttackEvent, react float64) float64 {
	a := atk.Snapshot.Stats.TotalATK()
	em := atk.Snapshot.Stats[attributes.EM]
	emBonus := (6 * em) / (2000 + em)
	damage := atk.Info.Mult * a * (1 + atk.Info.BaseDmgBonus)
	damage *= 1 + emBonus + react
	damage *= 1 + atk.Info.Elevation
	return applyCrit(damage, atk)
}

func applyCrit(damage float64, atk *info.AttackEvent) float64 {
	cr := atk.Snapshot.Stats[attributes.CR]
	if cr < 0 {
		cr = 0
	}
	if cr > 1 {
		cr = 1
	}
	if cr > 0 {
		damage *= 1 + atk.Snapshot.Stats[attributes.CD]
	}
	return damage
}
