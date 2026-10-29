package zibai

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/columbina"
	_ "github.com/genshinsim/gcsim/internal/characters/xingqiu"
	_ "github.com/genshinsim/gcsim/internal/characters/zhongli"
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
)

func init() {
	testhelper.RegisterTestWeapon()
}

func makeCore(trgCount int) (*core.Core, []*enemy.Enemy) {
	c, err := core.New(core.Opt{Seed: 1, Debug: false})
	if err != nil {
		panic(err)
	}
	a := avatar.New(c, info.Point{X: 0, Y: 0}, 1)
	c.Combat.SetPlayer(a)
	var trgs []*enemy.Enemy
	for range trgCount {
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

func addProfile(t *testing.T, c *core.Core, key keys.Char, cons, talent int) *character.CharWrapper {
	t.Helper()
	return addProfileAt(t, c, key, cons, talent, 90, 90)
}

func addProfileAt(t *testing.T, c *core.Core, key keys.Char, cons, talent, level, maxLevel int) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(key, testhelper.TestWeaponKey)
	prof.Base.Cons = cons
	prof.Base.Level = level
	prof.Base.MaxLevel = maxLevel
	prof.Talents = info.TalentProfile{Attack: talent, Skill: talent, Burst: talent}
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatalf("add %v: %v", key, err)
	}
	return c.Player.ByIndex(idx)
}

func asZibai(ch *character.CharWrapper) *char {
	return ch.Character.(*char)
}

func mustInit(t *testing.T, c *core.Core) {
	t.Helper()
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
}

func hold(c *core.Core, next action.Action) int {
	n := 0
	for c.Player.IsAnimationLocked(next) {
		advance(c, 1)
		n++
		if n > 300 {
			return n
		}
	}
	return n
}

type hit struct {
	abil   string
	tag    attacks.AttackTag
	ele    attributes.Element
	group  attacks.ICDGroup
	flat   float64
	mult   float64
	ignore float64
	dur    float64
	useDef bool
	noInf  bool
	elev   float64
	bonus  float64
	strike attacks.StrikeType
	poise  float64
}

func watch(c *core.Core) *[]hit {
	hits := &[]hit{}
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		*hits = append(*hits, hit{
			abil:   atk.Info.Abil,
			tag:    atk.Info.AttackTag,
			ele:    atk.Info.Element,
			group:  atk.Info.ICDGroup,
			flat:   atk.Info.FlatDmg,
			mult:   atk.Info.Mult,
			ignore: atk.Info.IgnoreDefPercent,
			dur:    float64(atk.Info.Durability),
			useDef: atk.Info.UseDef,
			noInf:  atk.Info.IgnoreInfusion,
			elev:   atk.Info.Elevation,
			bonus:  atk.Info.BaseDmgBonus,
			strike: atk.Info.StrikeType,
			poise:  atk.Info.PoiseDMG,
		})
	}, "zibai-hits")
	return hits
}

func countAbil(hits []hit, name string) int {
	n := 0
	for _, h := range hits {
		if h.abil == name {
			n++
		}
	}
	return n
}

func findAbil(hits []hit, name string) (hit, bool) {
	for _, h := range hits {
		if h.abil == name {
			return h, true
		}
	}
	return hit{}, false
}

func exec(t *testing.T, c *core.Core, a action.Action, k keys.Char) {
	t.Helper()
	if err := c.Player.Exec(a, k, nil); err != nil {
		t.Fatalf("%v: %v", a, err)
	}
}

func enterMode(t *testing.T, c *core.Core, ch *character.CharWrapper) {
	t.Helper()
	exec(t, c, action.ActionSkill, ch.Base.Key)
	advance(c, skillHitmark)
	if !ch.StatusIsActive(skillKey) {
		t.Fatal("lunar phase shift did not start")
	}
}

// weave fills Phase Shift Radiance from the ticker, Normal Attacks, and
// Lunar-Crystallize. It does not write the radiance field.
func weave(t *testing.T, c *core.Core, ch *character.CharWrapper, trg *enemy.Enemy, want float64) {
	t.Helper()
	limit := c.F + 20*60
	for c.F < limit && asZibai(ch).radiance < want {
		if !ch.StatusIsActive(skillKey) {
			t.Fatalf("mode ended at frame %d radiance %v", c.F, asZibai(ch).radiance)
		}
		if !c.Player.IsAnimationLocked(action.ActionAttack) {
			trg.AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Hydro, Durability: 50}})
			exec(t, c, action.ActionAttack, ch.Base.Key)
			continue
		}
		advance(c, 1)
	}
	if asZibai(ch).radiance < want {
		t.Fatalf("radiance %v at frame %d, want %v", asZibai(ch).radiance, c.F, want)
	}
}

func TestIdentity(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Zibai, 0, 1)
	mustInit(t, c)
	if ch.Base.Element != attributes.Geo || ch.Weapon.Class != info.WeaponClassSword || ch.EnergyMax != 60 {
		t.Fatalf("identity ele %v weapon %v energy %v", ch.Base.Element, ch.Weapon.Class, ch.EnergyMax)
	}
	if ch.Moonsign != 1 || c.Player.GetMoonsignLevel() != 1 {
		t.Fatalf("moonsign char %d level %d", ch.Moonsign, c.Player.GetMoonsignLevel())
	}
	if ch.Energy != ch.EnergyMax {
		t.Fatalf("starting energy %v", ch.Energy)
	}
	if math.Abs(ch.BaseStats[attributes.CD]-(0.5+0.384)) > 1e-6 {
		t.Fatalf("cd %v", ch.BaseStats[attributes.CD])
	}
	if c.Flags.Custom[reactable.LunarCrystallizeEnableKey] != 1 {
		t.Fatal("lunar crystallize was not enabled")
	}
	got, err := ch.Condition([]string{"radiance"})
	if err != nil || radianceOf(got) != 0 {
		t.Fatalf("radiance condition %v %v", got, err)
	}
}

func ticksBetween(start, end int) int {
	n := 0
	for f := start + 1; f <= end; f++ {
		if f >= skillHitmark && (f-skillHitmark)%6 == 0 {
			n++
		}
	}
	return n
}

func radianceOf(v any) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return math.NaN()
	}
}

func TestAscensionZeroStillEnablesLunarCrystallize(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfileAt(t, c, keys.Zibai, 0, 1, 20, 20)
	mustInit(t, c)
	if ch.Base.Ascension != 0 {
		t.Fatalf("ascension %d", ch.Base.Ascension)
	}
	if c.Flags.Custom[reactable.LunarCrystallizeEnableKey] != 1 {
		t.Fatal("ascension 0 did not enable lunar crystallize")
	}
}

func TestTalentCons(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Zibai, 6, 1)
	mustInit(t, c)
	if ch.TalentLvlAttack() != 0 || ch.TalentLvlSkill() != 3 || ch.TalentLvlBurst() != 3 {
		t.Fatalf("c6 talents na %d skill %d burst %d", ch.TalentLvlAttack(), ch.TalentLvlSkill(), ch.TalentLvlBurst())
	}
	c3, _ := makeCore(1)
	only := addProfile(t, c3, keys.Zibai, 3, 1)
	mustInit(t, c3)
	if only.TalentLvlSkill() != 3 || only.TalentLvlBurst() != 0 {
		t.Fatalf("c3 talents skill %d burst %d", only.TalentLvlSkill(), only.TalentLvlBurst())
	}
}

func TestC0NormalAndCharge(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Zibai, 0, 1)
	mustInit(t, c)
	hits := watch(c)
	exec(t, c, action.ActionAttack, keys.Zibai)
	advance(c, attackHitmarks[0][0]+2)
	h, ok := findAbil(*hits, "Normal 0")
	if !ok || h.ele != attributes.Physical || h.tag != attacks.AttackTagNormal || h.group != attacks.ICDGroupDefault || h.useDef {
		t.Fatalf("outside normal %+v", h)
	}
	if math.Abs(h.mult-attack[0][0][ch.TalentLvlAttack()]) > 1e-9 {
		t.Fatalf("normal mult %v", h.mult)
	}
	hold(c, action.ActionCharge)
	exec(t, c, action.ActionCharge, keys.Zibai)
	advance(c, chargeHitmarks[0]+2)
	charges := 0
	for _, hit := range *hits {
		if hit.abil != "Charge" {
			continue
		}
		charges++
		if hit.ele != attributes.Physical || hit.tag != attacks.AttackTagExtra || hit.group != attacks.ICDGroupDefault {
			t.Fatalf("outside charge %+v", hit)
		}
	}
	if charges != 2 {
		t.Fatalf("charge hits %d", charges)
	}
}

func TestPhaseShiftNormalsAndTicker(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Zibai, 0, 1)
	mustInit(t, c)
	ready, reason := ch.ActionReady(action.ActionSkill, nil)
	if !ready || reason != action.NoFailure {
		t.Fatalf("opening skill ready %v %v", ready, reason)
	}
	enterMode(t, c, ch)
	if asZibai(ch).radiance != 1 {
		t.Fatalf("first tick radiance %v", asZibai(ch).radiance)
	}
	if ch.Cooldown(action.ActionSkill) < 18*60-5 {
		t.Fatalf("skill cd %d", ch.Cooldown(action.ActionSkill))
	}
	ready, reason = ch.ActionReady(action.ActionSkill, nil)
	if ready || reason != action.InsufficientStamina {
		t.Fatalf("stride before 70 ready %v %v rad %v", ready, reason, asZibai(ch).radiance)
	}
	advance(c, 6)
	if asZibai(ch).radiance != 2 {
		t.Fatalf("second tick radiance %v", asZibai(ch).radiance)
	}
	got, _ := ch.Condition([]string{"radiance"})
	if radianceOf(got) != asZibai(ch).radiance {
		t.Fatalf("condition %v field %v", got, asZibai(ch).radiance)
	}

	hits := watch(c)
	hold(c, action.ActionAttack)
	before := asZibai(ch).radiance
	start := c.F
	exec(t, c, action.ActionAttack, keys.Zibai)
	advance(c, attackHitmarks[0][0]+2)
	h, ok := findAbil(*hits, "Normal 0")
	if !ok || h.ele != attributes.Geo || !h.useDef || !h.noInf || h.group != attacks.ICDGroupDefault || h.strike != attacks.StrikeTypeBlunt {
		t.Fatalf("phase normal %+v", h)
	}
	if math.Abs(h.mult-skillAttack[0][0][ch.TalentLvlAttack()]) > 1e-9 {
		t.Fatalf("phase normal mult %v", h.mult)
	}
	if asZibai(ch).radiance != before+5+float64(ticksBetween(start, c.F)) {
		t.Fatalf("na radiance %v before %v frames %d-%d", asZibai(ch).radiance, before, start, c.F)
	}
	held := asZibai(ch).radiance
	start = c.F
	hold(c, action.ActionAttack)
	exec(t, c, action.ActionAttack, keys.Zibai)
	advance(c, attackHitmarks[1][0]+2)
	if asZibai(ch).radiance != held+float64(ticksBetween(start, c.F)) {
		t.Fatalf("na icd did not hold, radiance %v from %v", asZibai(ch).radiance, held)
	}

	hold(c, action.ActionCharge)
	*hits = nil
	exec(t, c, action.ActionCharge, keys.Zibai)
	advance(c, chargeHitmarks[0]+2)
	charges := 0
	for _, hit := range *hits {
		if hit.abil != "Charge" {
			continue
		}
		charges++
		if hit.ele != attributes.Geo || !hit.useDef || !hit.noInf || hit.group != attacks.ICDGroupDefault || hit.poise != 60 {
			t.Fatalf("phase charge %+v", hit)
		}
		if math.Abs(hit.mult-skillCharge[charges-1][ch.TalentLvlSkill()]) > 1e-9 {
			t.Fatalf("phase charge mult %v", hit.mult)
		}
	}
	if charges != 2 {
		t.Fatalf("phase charge hits %d", charges)
	}
}

func TestTickerAloneMissesA1Window(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Zibai, 0, 1)
	mustInit(t, c)
	enterMode(t, c, ch)
	for asZibai(ch).radiance < 70 && c.F < 20*60 {
		advance(c, 1)
	}
	if asZibai(ch).radiance != 70 || c.F != skillHitmark+69*6 {
		t.Fatalf("ticker radiance %v frame %d", asZibai(ch).radiance, c.F)
	}
	if ch.StatusIsActive(a1Key) {
		t.Fatal("a1 still active when ticker alone reaches 70")
	}
	ready, reason := ch.ActionReady(action.ActionSkill, nil)
	if !ready || reason != action.NoFailure {
		t.Fatalf("stride ready %v %v", ready, reason)
	}
}

func TestC0StrideBurstAndFourUses(t *testing.T) {
	c, trg := makeCore(1)
	ch := addProfile(t, c, keys.Zibai, 0, 1)
	mustInit(t, c)
	hits := watch(c)
	applied := map[string]float64{}
	c.Events.Subscribe(event.OnApplyAttack, func(args ...any) {
		atk := args[0].(*info.AttackEvent)
		if _, ok := applied[atk.Info.Abil]; !ok {
			applied[atk.Info.Abil] = float64(atk.Info.Durability)
		}
	}, "zibai-gauge")
	enterMode(t, c, ch)
	weave(t, c, ch, trg[0], 70)
	if !ch.StatusIsActive(a1Key) {
		t.Fatal("woven radiance reached 70 after a1 expired")
	}
	ready, reason := ch.ActionReady(action.ActionSkill, nil)
	if !ready || reason != action.NoFailure {
		t.Fatalf("first stride ready %v %v", ready, reason)
	}
	before := asZibai(ch).radiance
	exec(t, c, action.ActionSkill, keys.Zibai)
	if asZibai(ch).radiance != before-70 || asZibai(ch).skillsUsed != 1 {
		t.Fatalf("spend rad %v used %d", asZibai(ch).radiance, asZibai(ch).skillsUsed)
	}
	advance(c, skillStrideHitmark[1]+2)
	first, ok1 := findAbil(*hits, skillAbil1)
	second, ok2 := findAbil(*hits, skillAbil2)
	if !ok1 || !ok2 {
		t.Fatalf("stride hits %v", *hits)
	}
	if first.tag != attacks.AttackTagElementalArt || first.ele != attributes.Geo || !first.useDef || first.group != attacks.ICDGroupDefault || first.poise != 75 {
		t.Fatalf("stride 1 %+v", first)
	}
	if applied[skillAbil1] != 25 {
		t.Fatalf("stride 1 gauge at apply %v", applied[skillAbil1])
	}
	if math.Abs(first.mult-skillStride[0][ch.TalentLvlSkill()]) > 1e-9 {
		t.Fatalf("stride 1 mult %v", first.mult)
	}
	if second.tag != attacks.AttackTagDirectLunarCrystallize || second.dur != 0 || second.ignore != 1 || second.group != attacks.ICDGroupDefault || second.poise != 50 {
		t.Fatalf("stride 2 %+v", second)
	}
	wantFlat := 0.6 * ch.TotalDef(false)
	if math.Abs(second.flat-wantFlat) > 1e-6 {
		t.Fatalf("a1 flat %v want %v", second.flat, wantFlat)
	}
	wantBonus := math.Min(ch.TotalDef(false)/100.0*0.007, 0.14)
	if math.Abs(second.bonus-wantBonus) > 1e-9 {
		t.Fatalf("def bonus %v want %v", second.bonus, wantBonus)
	}

	// Three more natural strides. The fourth spends the mode.
	for asZibai(ch).skillsUsed < 4 {
		weave(t, c, ch, trg[0], 70)
		hold(c, action.ActionSkill)
		ready, reason = ch.ActionReady(action.ActionSkill, nil)
		if !ready {
			t.Fatalf("stride %d not ready: %v rad %v", asZibai(ch).skillsUsed+1, reason, asZibai(ch).radiance)
		}
		exec(t, c, action.ActionSkill, keys.Zibai)
	}
	if ch.StatusIsActive(skillKey) || asZibai(ch).skillsUsed != 4 {
		t.Fatalf("mode after 4 used %d active %v frame %d", asZibai(ch).skillsUsed, ch.StatusIsActive(skillKey), c.F)
	}
	ready, reason = ch.ActionReady(action.ActionSkill, nil)
	if ready || reason != action.SkillCD {
		t.Fatalf("reentry ready %v %v cd %d", ready, reason, ch.Cooldown(action.ActionSkill))
	}

	// Burst outside the mode still has two hits. Energy started full.
	*hits = nil
	exec(t, c, action.ActionBurst, keys.Zibai)
	if ch.Cooldown(action.ActionBurst) != 15*60 {
		t.Fatalf("burst cd %d", ch.Cooldown(action.ActionBurst))
	}
	advance(c, 8)
	if ch.Energy != 0 {
		t.Fatalf("energy after burst %v", ch.Energy)
	}
	advance(c, burstHitmarks[0])
	b1, ok1 := findAbil(*hits, "Burst")
	b2, ok2 := findAbil(*hits, "Burst"+lunarCrystallizeAbil)
	if !ok1 || !ok2 || b1.tag != attacks.AttackTagElementalBurst || b2.tag != attacks.AttackTagDirectLunarCrystallize || b2.ignore != 1 || b2.dur != 0 || b2.poise != 100 {
		t.Fatalf("burst %+v %+v names %v", b1, b2, *hits)
	}
	if b1.group != attacks.ICDGroupDefault || b2.group != attacks.ICDGroupDefault {
		t.Fatal("burst icd")
	}
	ready, reason = ch.ActionReady(action.ActionBurst, nil)
	if ready || reason != action.InsufficientEnergy {
		t.Fatalf("burst after drain ready %v %v", ready, reason)
	}
}

func TestBurstExtendsMode(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Zibai, 0, 1)
	mustInit(t, c)
	enterMode(t, c, ch)
	before := ch.StatusDuration(skillKey)
	exec(t, c, action.ActionBurst, keys.Zibai)
	if ch.StatusDuration(skillKey) != before+102 {
		t.Fatalf("extension before %d after %d", before, ch.StatusDuration(skillKey))
	}
}

func TestSwapClearsMode(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Zibai, 0, 1)
	addProfile(t, c, keys.Zhongli, 0, 1)
	mustInit(t, c)
	enterMode(t, c, ch)
	exec(t, c, action.ActionSwap, keys.Zhongli)
	advance(c, 5)
	if c.Player.Active() == ch.Index() || ch.StatusIsActive(skillKey) {
		t.Fatalf("active %d status %v", c.Player.Active(), ch.StatusIsActive(skillKey))
	}
	got, _ := ch.Condition([]string{"radiance"})
	if radianceOf(got) != 0 {
		t.Fatalf("off-field condition %v", got)
	}
}

func TestHydroConvertsToLunarCrystallize(t *testing.T) {
	plain, ptrg := makeCore(1)
	addProfile(t, plain, keys.Zhongli, 0, 1)
	mustInit(t, plain)
	var hydro, lunar int
	plain.Events.Subscribe(event.OnCrystallizeHydro, func(...any) { hydro++ }, "cry")
	plain.Events.Subscribe(event.OnLunarCrystallize, func(...any) { lunar++ }, "lcr")
	ptrg[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Hydro, Durability: 50}})
	advance(plain, 1)
	gauge(plain, ptrg[0], 0, attributes.Geo)
	if hydro != 1 || lunar != 0 {
		t.Fatalf("without zibai hydro %d lunar %d", hydro, lunar)
	}

	live, ltrg := makeCore(1)
	ch := addProfile(t, live, keys.Zibai, 0, 1)
	mustInit(t, live)
	hydro, lunar = 0, 0
	live.Events.Subscribe(event.OnCrystallizeHydro, func(...any) { hydro++ }, "cry")
	live.Events.Subscribe(event.OnLunarCrystallize, func(...any) { lunar++ }, "lcr")
	ltrg[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Hydro, Durability: 50}})
	advance(live, 1)
	gauge(live, ltrg[0], ch.Index(), attributes.Geo)
	if hydro != 0 || lunar != 1 {
		t.Fatalf("with zibai hydro %d lunar %d", hydro, lunar)
	}
}

func TestA4PartyStats(t *testing.T) {
	solo, _ := makeCore(1)
	alone := addProfile(t, solo, keys.Zibai, 0, 1)
	mustInit(t, solo)

	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Zibai, 0, 1)
	addProfile(t, c, keys.Zhongli, 0, 1)
	addProfile(t, c, keys.Xingqiu, 0, 1)
	mustInit(t, c)
	if math.Abs(ch.Stat(attributes.DEFP)-alone.Stat(attributes.DEFP)-0.15) > 1e-9 {
		t.Fatalf("def%% %v solo %v", ch.Stat(attributes.DEFP), alone.Stat(attributes.DEFP))
	}
	if math.Abs(ch.Stat(attributes.EM)-alone.Stat(attributes.EM)-60) > 1e-9 {
		t.Fatalf("em %v solo %v", ch.Stat(attributes.EM), alone.Stat(attributes.EM))
	}
}

func TestC1ImmediateRadianceAndFiveStrides(t *testing.T) {
	c, trg := makeCore(1)
	ch := addProfile(t, c, keys.Zibai, 1, 1)
	mustInit(t, c)
	enterMode(t, c, ch)
	if asZibai(ch).radiance != 100 || !ch.StatusIsActive(c1Key) {
		t.Fatalf("c1 rad %v status %v", asZibai(ch).radiance, ch.StatusIsActive(c1Key))
	}
	ready, reason := ch.ActionReady(action.ActionSkill, nil)
	if !ready || reason != action.NoFailure {
		t.Fatalf("c1 stride ready %v %v", ready, reason)
	}
	ai := info.AttackInfo{ActorIndex: ch.Index(), Abil: skillAbil2, AttackTag: attacks.AttackTagDirectLunarCrystallize}
	if math.Abs(ch.ReactBonus(ai)-2.2) > 1e-9 {
		t.Fatalf("c1 bonus %v", ch.ReactBonus(ai))
	}
	advance(c, 1)
	if ch.StatusIsActive(c1Key) {
		t.Fatal("c1 status survived the first stride bonus")
	}
	hold(c, action.ActionSkill)
	exec(t, c, action.ActionSkill, keys.Zibai)
	if asZibai(ch).radiance != 30 || asZibai(ch).skillsUsed != 1 || !ch.StatusIsActive(skillKey) {
		t.Fatalf("after first rad %v used %d", asZibai(ch).radiance, asZibai(ch).skillsUsed)
	}
	for asZibai(ch).skillsUsed < 5 {
		weave(t, c, ch, trg[0], 70)
		hold(c, action.ActionSkill)
		exec(t, c, action.ActionSkill, keys.Zibai)
	}
	if ch.StatusIsActive(skillKey) || asZibai(ch).skillsUsed != 5 {
		t.Fatalf("c1 cap used %d active %v frame %d", asZibai(ch).skillsUsed, ch.StatusIsActive(skillKey), c.F)
	}
}

func TestC2ReactBonusAndGleamA1(t *testing.T) {
	c, trg := makeCore(1)
	ch := addProfile(t, c, keys.Zibai, 2, 1)
	bina := addProfile(t, c, keys.Columbina, 0, 1)
	mustInit(t, c)
	if c.Player.GetMoonsignLevel() != 2 {
		t.Fatalf("moonsign %d", c.Player.GetMoonsignLevel())
	}
	direct := info.AttackInfo{AttackTag: attacks.AttackTagDirectLunarCrystallize}
	react := info.AttackInfo{AttackTag: attacks.AttackTagReactionLunarCrystallize}
	normal := info.AttackInfo{AttackTag: attacks.AttackTagNormal}
	if ch.ReactBonus(direct) != 0 || bina.ReactBonus(react) != 0 {
		t.Fatal("c2 bonus applied outside the mode")
	}
	enterMode(t, c, ch)
	if math.Abs(ch.ReactBonus(direct)-0.3) > 1e-9 || math.Abs(bina.ReactBonus(react)-0.3) > 1e-9 || ch.ReactBonus(normal) != 0 {
		t.Fatalf("c2 in mode zibai %v columbina %v normal %v", ch.ReactBonus(direct), bina.ReactBonus(react), ch.ReactBonus(normal))
	}
	weave(t, c, ch, trg[0], 70)
	if !ch.StatusIsActive(a1Key) {
		c.Events.Emit(event.OnMoondriftHarmony, trg[0])
	}
	if !ch.StatusIsActive(a1Key) {
		t.Fatal("a1 did not refresh")
	}
	hits := watch(c)
	hold(c, action.ActionSkill)
	exec(t, c, action.ActionSkill, keys.Zibai)
	advance(c, skillStrideHitmark[1]+2)
	second, ok := findAbil(*hits, skillAbil2)
	want := 5.5 * ch.TotalDef(false)
	if !ok || math.Abs(second.flat-want) > 1e-4 {
		t.Fatalf("c2 flat %v want %v ok %v", second.flat, want, ok)
	}
}

func TestC4CounterAndN4(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Zibai, 4, 1)
	addProfile(t, c, keys.Columbina, 0, 1)
	mustInit(t, c)
	enterMode(t, c, ch)
	hold(c, action.ActionAttack)
	exec(t, c, action.ActionAttack, keys.Zibai)
	if ch.NormalCounter != 1 {
		t.Fatalf("counter %d", ch.NormalCounter)
	}
	ch.ResetNormalCounter()
	if ch.NormalCounter != 1 {
		t.Fatalf("c4 reset counter to %d", ch.NormalCounter)
	}

	plain, _ := makeCore(1)
	base := addProfile(t, plain, keys.Zibai, 0, 1)
	mustInit(t, plain)
	enterMode(t, plain, base)
	base.NormalCounter = 2
	base.ResetNormalCounter()
	if base.NormalCounter != 0 {
		t.Fatalf("c0 reset counter %d", base.NormalCounter)
	}

	// Gleam N4 extra. C4's status multiplies it by 2.5 after a stride hit.
	live, trg := makeCore(1)
	z := addProfile(t, live, keys.Zibai, 4, 1)
	addProfile(t, live, keys.Columbina, 0, 1)
	mustInit(t, live)
	hits := watch(live)
	enterMode(t, live, z)
	weave(t, live, z, trg[0], 70)
	hold(live, action.ActionSkill)
	exec(t, live, action.ActionSkill, keys.Zibai)
	advance(live, skillStrideHitmark[1]+2)
	if !z.StatusIsActive(c4Key) {
		t.Fatal("c4 status missing after stride hit")
	}
	for z.NormalCounter != 3 {
		hold(live, action.ActionAttack)
		exec(t, live, action.ActionAttack, keys.Zibai)
	}
	hold(live, action.ActionAttack)
	exec(t, live, action.ActionAttack, keys.Zibai)
	advance(live, attackHitmarks[3][0]+20+5)
	extra, ok := findAbil(*hits, "Normal 4"+lunarCrystallizeAbil)
	want := skillLastAttackBonus[z.TalentLvlAttack()] * 2.5
	if !ok || extra.tag != attacks.AttackTagDirectLunarCrystallize || extra.group != attacks.ICDGroupDefault || math.Abs(extra.mult-want) > 1e-9 {
		t.Fatalf("c4 n4 %+v want %v", extra, want)
	}
	if z.StatusIsActive(c4Key) {
		t.Fatal("c4 status was not consumed")
	}
}

func TestC6GainSpendAndElevation(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Zibai, 6, 1)
	mustInit(t, c)
	enterMode(t, c, ch)
	// C6 includes C1. The opening 100 is scaled by 1.5 and then capped.
	if asZibai(ch).radiance != 100 {
		t.Fatalf("c6 opening radiance %v", asZibai(ch).radiance)
	}
	var elev []float64
	c.Events.Subscribe(event.OnApplyAttack, func(args ...any) {
		atk := args[0].(*info.AttackEvent)
		if atk.Info.Abil == skillAbil2 {
			elev = append(elev, atk.Info.Elevation)
		}
	}, "c6-elev")
	hold(c, action.ActionSkill)
	exec(t, c, action.ActionSkill, keys.Zibai)
	if asZibai(ch).radiance != 0 {
		t.Fatalf("c6 leftover %v", asZibai(ch).radiance)
	}
	want := (100.0 - 70.0) * 0.016
	if math.Abs(asZibai(ch).c6Elev-want) > 1e-9 || !ch.StatusIsActive(c6Key) {
		t.Fatalf("elev stored %v status %v", asZibai(ch).c6Elev, ch.StatusIsActive(c6Key))
	}
	sawTick := false
	for range 6 {
		advance(c, 1)
		if asZibai(ch).radiance == 0 {
			continue
		}
		if asZibai(ch).radiance != 1.5 {
			t.Fatalf("c6 tick %v", asZibai(ch).radiance)
		}
		sawTick = true
		break
	}
	if !sawTick {
		t.Fatal("c6 ticker did not add 1.5")
	}
	advance(c, skillStrideHitmark[1])
	if len(elev) != 1 || math.Abs(elev[0]-want) > 1e-9 {
		t.Fatalf("applied elev %v", elev)
	}
}

func TestParticlesAreGeoAndICDLimited(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Zibai, 0, 1)
	mustInit(t, c)
	ch.Energy = 0
	var particles []character.Particle
	c.Events.Subscribe(event.OnParticleReceived, func(args ...any) {
		particles = append(particles, args[0].(character.Particle))
	}, "particles")
	enterMode(t, c, ch)
	limit := c.F + 15*60
	for len(particles) == 0 && c.F < limit {
		if !c.Player.IsAnimationLocked(action.ActionAttack) {
			exec(t, c, action.ActionAttack, keys.Zibai)
			continue
		}
		advance(c, 1)
	}
	if len(particles) != 1 || particles[0].Num != 1 || particles[0].Ele != attributes.Geo {
		t.Fatalf("particle %+v", particles)
	}
	if ch.Energy <= 0 {
		t.Fatal("particle did not restore energy")
	}
	locked := c.F
	for c.F < locked+60 {
		if !c.Player.IsAnimationLocked(action.ActionAttack) {
			exec(t, c, action.ActionAttack, keys.Zibai)
			continue
		}
		advance(c, 1)
	}
	if len(particles) != 1 {
		t.Fatalf("particle inside icd %+v", particles)
	}
}

func gauge(c *core.Core, trg *enemy.Enemy, actor int, ele attributes.Element) {
	c.QueueAttack(info.AttackInfo{
		ActorIndex: actor,
		Abil:       "probe",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagNone,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    ele,
		Durability: 25,
		Mult:       1,
	}, combat.NewSingleTargetHit(trg.Key()), 0, 0)
}
