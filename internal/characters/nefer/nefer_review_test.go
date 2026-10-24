package nefer

import (
	"math"
	"testing"

	"github.com/genshinsim/gcsim/internal/characters/nilou"
	"github.com/genshinsim/gcsim/internal/template/dendrocore"
	"github.com/genshinsim/gcsim/pkg/avatar"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/testhelper"

	_ "github.com/genshinsim/gcsim/internal/characters/collei"
	_ "github.com/genshinsim/gcsim/internal/characters/nilou"
)

func asNefer(ch *character.CharWrapper) *char { return ch.Character.(*char) }

func addNefer(t *testing.T, c *core.Core, cons int, em float64) *character.CharWrapper {
	t.Helper()
	prof := testhelper.DefaultProfile(keys.Nefer, testhelper.TestWeaponKey)
	prof.Base.Cons = cons
	prof.Talents = info.TalentProfile{Attack: 1, Skill: 1, Burst: 1}
	prof.Stats[attributes.EM] = em
	prof.Stats[attributes.DendroP] = 0.5
	idx, err := c.AddChar(prof)
	if err != nil {
		t.Fatal(err)
	}
	return c.Player.ByIndex(idx)
}

type recorded struct {
	info info.AttackInfo
	dmg  float64
	snap attributes.Stats
	lvl  int
}

func collectDmg(c *core.Core) *[]recorded {
	hits := []recorded{}
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ae := args[1].(*info.AttackEvent)
		hits = append(hits, recorded{
			info: ae.Info,
			dmg:  args[2].(float64),
			snap: ae.Snapshot.Stats,
			lvl:  ae.Snapshot.CharLvl,
		})
	}, "dmg")
	return &hits
}

func findHit(hits []recorded, name string) (recorded, bool) {
	for _, h := range hits {
		if h.info.Abil == name {
			return h, true
		}
	}
	return recorded{}, false
}

func atkOf(ch *character.CharWrapper) float64 {
	return ch.BaseStats[attributes.BaseATK]*(1+ch.Stat(attributes.ATKP)) + ch.Stat(attributes.ATK)
}

func near(t *testing.T, got, want float64) {
	t.Helper()
	tol := 1e-6 * math.Max(1, math.Abs(want))
	if math.Abs(got-want) > tol {
		t.Fatalf("got %v want %v", got, want)
	}
}

// ordinary talent damage: (ATK ratio * ATK + EM ratio * EM) * (1 + base multiplier) * (1 + DMG bonus) * def.
func ordinaryDmg(atk, em, atkR, emR, baseMult, dmgBonus float64, lvl int) float64 {
	base := (atkR*atk + emR*em) * (1 + baseMult)
	defmod := float64(lvl+100) / (float64(lvl+100) + 200)
	return base * (1 + dmgBonus) * defmod
}

// direct lunar-bloom: ratio * EM * (1 + moonsign + veil) * (1 + EM reaction bonus) * (1 + elevation).
// Defense is ignored and DMG Bonus is not part of this result.
func lunarDmg(em, ratio, veil, elevation float64) float64 {
	moonsign := math.Min(em*0.000175, 0.14)
	raw := ratio * em * (1 + moonsign + veil)
	return raw * (1 + (6*em)/(2000+em)) * (1 + elevation)
}

func openDance(t *testing.T, c *core.Core, ch *character.CharWrapper) {
	t.Helper()
	c.Player.SetVerdantDew(1)
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillAnim+1)
	ch.BaseStats[attributes.CR] = 0
}

func TestVeilIsBaseDamageMultiplier(t *testing.T) {
	c, _ := makeCore(1)
	ch := addNefer(t, c, 0, 800)
	start(t, c, ch.Index())
	openDance(t, c, ch)
	asNefer(ch).addVeils(3)
	if asNefer(ch).veilCount() != 3 {
		t.Fatalf("veils %d", asNefer(ch).veilCount())
	}
	atk, em := atkOf(ch), ch.Stat(attributes.EM)
	hits := collectDmg(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, ppAnim+1)

	neferHit, ok := findHit(*hits, "Phantasm Performance (Nefer 1)")
	if !ok {
		t.Fatal("missing nefer hit")
	}
	shade, ok := findHit(*hits, "Phantasm Performance (Shade 1)")
	if !ok {
		t.Fatal("missing shade")
	}
	lvl := ch.TalentLvlSkill()
	const veil = 0.24
	near(t, neferHit.dmg, ordinaryDmg(atk, em, ppN1ATK[lvl], ppN1EM[lvl], veil, 0.5, neferHit.lvl))
	wrong := ordinaryDmg(atk, em, ppN1ATK[lvl], ppN1EM[lvl], 0, 0.5+veil, neferHit.lvl)
	if math.Abs(neferHit.dmg-wrong) <= 1e-4 {
		t.Fatalf("nefer hit matched the DMG Bonus bucket %v", neferHit.dmg)
	}
	if math.Abs(neferHit.snap[attributes.DmgP]) > 1e-12 || math.Abs(neferHit.info.BaseDmgBonus-veil) > 1e-12 {
		t.Fatalf("nefer buckets dmg%% %v base %v", neferHit.snap[attributes.DmgP], neferHit.info.BaseDmgBonus)
	}

	near(t, shade.dmg, lunarDmg(em, shade1[lvl], veil, 0))
	if math.Abs(shade.dmg-lunarDmg(em, shade1[lvl], 0, 0)) <= 1e-4 {
		t.Fatal("shade ignored the veil")
	}
	if math.Abs(shade.snap[attributes.DendroP]-0.5) > 1e-12 {
		t.Fatalf("dendro%% %v", shade.snap[attributes.DendroP])
	}
	moonsign := math.Min(em*0.000175, 0.14)
	if math.Abs(shade.info.BaseDmgBonus-(moonsign+veil)) > 1e-12 {
		t.Fatalf("shade base %v moonsign %v", shade.info.BaseDmgBonus, moonsign)
	}
}

func TestC1IsInsideTheVeilMultiplier(t *testing.T) {
	c, _ := makeCore(1)
	ch := addNefer(t, c, 1, 800)
	start(t, c, ch.Index())
	openDance(t, c, ch)
	asNefer(ch).addVeils(3)
	em := ch.Stat(attributes.EM)
	hits := collectDmg(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, ppAnim+1)
	shade, ok := findHit(*hits, "Phantasm Performance (Shade 1)")
	if !ok {
		t.Fatal("missing shade")
	}
	ratio := shade1[ch.TalentLvlSkill()] + 0.6
	const veil = 0.24
	near(t, shade.dmg, lunarDmg(em, ratio, veil, 0))
	if math.Abs(shade.info.Mult-ratio) > 1e-12 {
		t.Fatalf("mult %v", shade.info.Mult)
	}
	doubled := lunarDmg(em, ratio, veil, 0) * (1 + veil)
	if math.Abs(shade.dmg-doubled) <= 1e-4 {
		t.Fatal("c1 was multiplied by the veil twice")
	}
}

func TestC2FiveStacksAreThe140Percent(t *testing.T) {
	c, _ := makeCore(1)
	ch := addNefer(t, c, 2, 800)
	start(t, c, ch.Index())
	openDance(t, c, ch)
	// The skill already granted 2. Three more reach the C2 cap.
	asNefer(ch).addVeils(3)
	if asNefer(ch).veilCount() != 5 {
		t.Fatalf("veils %d", asNefer(ch).veilCount())
	}
	em := ch.Stat(attributes.EM)
	hits := collectDmg(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, ppAnim+1)
	shade, ok := findHit(*hits, "Phantasm Performance (Shade 1)")
	if !ok {
		t.Fatal("missing shade")
	}
	const veil = 0.40
	ratio := shade1[ch.TalentLvlSkill()]
	if ch.Base.Cons >= 1 {
		ratio += 0.6
	}
	want := lunarDmg(em, ratio, veil, 0)
	near(t, shade.dmg, want)
	if math.Abs(shade.dmg-want*1.4) <= 1e-4 {
		t.Fatal("a second 140% multiplier was applied")
	}
	plain := lunarDmg(em, ratio, 0, 0)
	moonsign := math.Min(em*0.000175, 0.14)
	// "140% of original DMG" is this 40% base multiplier. Moonsign sits in the same bucket, so the
	// ratio against an unbuffed shade is (1 + moonsign + 0.40) / (1 + moonsign), not an extra 1.4.
	if math.Abs(shade.dmg/plain-(1+moonsign+veil)/(1+moonsign)) > 1e-9 {
		t.Fatalf("five stacks scaled the shade by %v", shade.dmg/plain)
	}
	if math.Abs(shade.info.BaseDmgBonus-(moonsign+veil)) > 1e-12 {
		t.Fatalf("base %v", shade.info.BaseDmgBonus)
	}
}

func TestC6VeilAndElevationStaySeparate(t *testing.T) {
	c, _ := makeCore(1)
	ch := addNefer(t, c, 6, 800)
	start(t, c, ch.Index())
	// Gleam is the party's Moonsign. Set the flag directly so Columbina's own buffs stay out of this check.
	asNefer(ch).gleam = true
	openDance(t, c, ch)
	asNefer(ch).addVeils(3)
	if asNefer(ch).veilCount() != 5 {
		t.Fatalf("veils %d", asNefer(ch).veilCount())
	}
	em := ch.Stat(attributes.EM)
	hits := collectDmg(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, ppAnim+1)
	conv, ok := findHit(*hits, "Phantasm Performance C6 (Nefer 2)")
	if !ok {
		t.Fatal("missing c6 hit")
	}
	extra, ok := findHit(*hits, "Phantasm Performance C6 (Extra)")
	if !ok {
		t.Fatal("missing c6 extra")
	}
	const veil = 0.40
	// C6 includes C4, and Shadow Dance is still up, so Dendro RES is -20%.
	// Negative RES uses 1 - res/2, which is 1.1 here.
	const resmod = 1.1
	near(t, conv.dmg, lunarDmg(em, 0.85+0.6, veil, 0.15)*resmod)
	near(t, extra.dmg, lunarDmg(em, 1.2+0.6, veil, 0.15)*resmod)
	if conv.info.Elevation != 0.15 || extra.info.Elevation != 0.15 {
		t.Fatalf("elevation %v %v", conv.info.Elevation, extra.info.Elevation)
	}
	folded := lunarDmg(em, 1.45, veil+0.15, 0) * resmod
	if math.Abs(conv.dmg-folded) <= 1e-4 {
		t.Fatal("elevation was added into the base multiplier")
	}
}

func TestBurstVeilIsDamageBonus(t *testing.T) {
	c, _ := makeCore(1)
	ch := addNefer(t, c, 0, 800)
	start(t, c, ch.Index())
	ch.BaseStats[attributes.CR] = 0
	ch.Energy = ch.EnergyMax
	asNefer(ch).addVeils(2)
	atk, em := atkOf(ch), ch.Stat(attributes.EM)
	hits := collectDmg(c)
	if err := c.Player.Exec(action.ActionBurst, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, burstAnim+1)
	if asNefer(ch).veilCount() != 0 {
		t.Fatal("burst left veils behind")
	}
	first, ok := findHit(*hits, "Sacred Vow (1)")
	if !ok {
		t.Fatal("missing burst hit")
	}
	lvl := ch.TalentLvlBurst()
	bonus := 2 * burstVeil[lvl]
	near(t, first.dmg, ordinaryDmg(atk, em, burst1ATK[lvl], burst1EM[lvl], 0, 0.5+bonus, first.lvl))
	if math.Abs(first.info.BaseDmgBonus) > 1e-12 || math.Abs(first.snap[attributes.DmgP]-bonus) > 1e-12 {
		t.Fatalf("burst base %v dmg%% %v", first.info.BaseDmgBonus, first.snap[attributes.DmgP])
	}
	asBase := ordinaryDmg(atk, em, burst1ATK[lvl], burst1EM[lvl], bonus, 0.5, first.lvl)
	if math.Abs(first.dmg-asBase) <= 1e-4 {
		t.Fatal("burst veil was applied as a base multiplier")
	}
}

func countParticles(t *testing.T, seed int64, hit bool, enemies int) (float64, attributes.Element, int) {
	t.Helper()
	c, err := core.New(core.Opt{Seed: seed, Debug: false})
	if err != nil {
		t.Fatal(err)
	}
	c.Combat.SetPlayer(avatar.New(c, info.Point{X: 0, Y: 0}, 1))
	var first *enemy.Enemy
	for range enemies {
		e := enemy.New(c, info.EnemyProfile{
			Level:  100,
			Resist: make(map[attributes.Element]float64),
			Pos:    info.Coord{X: 0, Y: 0, R: 1},
		})
		c.Combat.AddEnemy(e)
		if first == nil {
			first = e
			c.Combat.DefaultTarget = e.Key()
		}
	}
	ch := addChar(t, c, keys.Nefer, 0, 1)
	start(t, c, ch.Index())
	ch.ParticleDelay = 0
	var n float64
	var ele attributes.Element
	var events int
	c.Events.Subscribe(event.OnParticleReceived, func(args ...any) {
		p := args[0].(character.Particle)
		n += p.Num
		ele = p.Ele
		events++
	}, "particles")
	if !hit && first != nil {
		first.Kill()
	}
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillHit+1)
	return n, ele, events
}

func TestSkillParticlesOnHit(t *testing.T) {
	const trials = 120
	var total, threes float64
	for seed := int64(1); seed <= trials; seed++ {
		n, ele, events := countParticles(t, seed, true, 1)
		if events != 1 || (n != 2 && n != 3) || ele != attributes.Dendro {
			t.Fatalf("seed %d events %d num %v ele %v", seed, events, n, ele)
		}
		total += n
		if n == 3 {
			threes++
		}
	}
	avg := total / trials
	share := threes / trials
	if avg < 2.55 || avg > 2.78 || share < 0.55 || share > 0.78 {
		t.Fatalf("average %v share of 3 %v", avg, share)
	}
}

func TestSkillParticlesMiss(t *testing.T) {
	n, _, events := countParticles(t, 1, false, 1)
	if events != 0 || n != 0 {
		t.Fatalf("miss generated %v across %d events", n, events)
	}
}

func TestSkillParticlesOncePerCast(t *testing.T) {
	c, trg := makeCore(2)
	trg[1].SetPos(info.Point{X: 0, Y: 0})
	ch := addChar(t, c, keys.Nefer, 0, 1)
	start(t, c, ch.Index())
	ch.ParticleDelay = 0
	var events int
	c.Events.Subscribe(event.OnParticleReceived, func(args ...any) {
		events++
	}, "particles")
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillHit+1)
	if events != 1 {
		t.Fatalf("one cast, two enemies, events %d", events)
	}
	c.Player.SetVerdantDew(1)
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, ppAnim+1)
	if events != 1 {
		t.Fatalf("phantasm generated particles, events %d", events)
	}
}

func reactionCount(c *core.Core, name string) *int {
	n := 0
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		ae := args[1].(*info.AttackEvent)
		if ae.Info.Abil == name {
			n++
		}
	}, name)
	return &n
}

func TestOrdinaryBloomWithoutNefer(t *testing.T) {
	c, _ := makeCore(1)
	col := addChar(t, c, keys.Collei, 0, 1)
	start(t, c, col.Index())
	blooms := reactionCount(c, "bloom")
	spawnOwnedCore(c, col.Index())
	advance(c, 400)
	if *blooms == 0 {
		t.Fatal("ordinary bloom did not explode")
	}
}

func eventCount(c *core.Core, ev event.Event, key string) *int {
	n := 0
	c.Events.Subscribe(ev, func(args ...any) { n++ }, key)
	return &n
}

// Lunar-Bloom replaces the Bloom reaction event and still spawns a Dendro Core.
// The core's later explosion is damage labeled "bloom"; it is not a second reaction.
func TestLunarBloomEventIsDistinctFromCoreExplosion(t *testing.T) {
	c, _ := makeCore(1)
	col := addChar(t, c, keys.Collei, 0, 1)
	start(t, c, col.Index())
	onBloom := eventCount(c, event.OnBloom, "ordinary-bloom")
	onLunar := eventCount(c, event.OnLunarBloom, "ordinary-lunar")
	explosion := reactionCount(c, "bloom")
	spawnOwnedCore(c, col.Index())
	advance(c, 400)
	if *onBloom == 0 || *onLunar != 0 || *explosion == 0 {
		t.Fatalf("ordinary bloom: reaction %d lunar %d explosion %d", *onBloom, *onLunar, *explosion)
	}

	c, _ = makeCore(1)
	col = addChar(t, c, keys.Collei, 0, 1)
	addChar(t, c, keys.Nefer, 0, 1)
	start(t, c, col.Index())
	onBloom = eventCount(c, event.OnBloom, "nefer-bloom")
	onLunar = eventCount(c, event.OnLunarBloom, "nefer-lunar")
	explosion = reactionCount(c, "bloom")
	spawnOwnedCore(c, col.Index())
	advance(c, 400)
	if *onBloom != 0 || *onLunar == 0 || *explosion == 0 {
		t.Fatalf("lunar bloom: reaction %d lunar %d explosion %d", *onBloom, *onLunar, *explosion)
	}
}

func TestOrdinaryHyperbloomAndBurgeon(t *testing.T) {
	c, trg := makeCore(1)
	col := addChar(t, c, keys.Collei, 0, 1)
	start(t, c, col.Index())
	hyper := reactionCount(c, "hyperbloom")
	burgeon := reactionCount(c, "burgeon")
	spawnOwnedCore(c, col.Index())
	c.QueueAttackEvent(&info.AttackEvent{
		Info:    info.AttackInfo{ActorIndex: col.Index(), Element: attributes.Electro, Durability: 25},
		Pattern: combat.NewCircleHitOnTarget(trg[0], nil, 10),
	}, 0)
	advance(c, 70)
	if *hyper != 1 {
		t.Fatalf("hyperbloom %d", *hyper)
	}
	spawnOwnedCore(c, col.Index())
	c.QueueAttackEvent(&info.AttackEvent{
		Info:    info.AttackInfo{ActorIndex: col.Index(), Element: attributes.Pyro, Durability: 50},
		Pattern: combat.NewCircleHitOnTarget(trg[0], nil, 10),
	}, 0)
	advance(c, 5)
	if *burgeon != 1 {
		t.Fatalf("burgeon %d", *burgeon)
	}
}

func armNilou(t *testing.T, c *core.Core) {
	t.Helper()
	// The stance entry does not count. A1 is queued on the third dance step.
	for range 4 {
		if err := c.Player.Exec(action.ActionSkill, keys.Nilou, nil); err != nil {
			t.Fatal(err)
		}
		advance(c, 5)
	}
	advance(c, 45)
	nl, ok := c.Player.ByKey(keys.Nilou)
	if !ok || !nl.StatusIsActive("nilou-a1") {
		t.Fatal("nilou a1 did not arm")
	}
}

func spawnOwnedCore(c *core.Core, owner int) {
	c.QueueAttackEvent(&info.AttackEvent{
		Info:    info.AttackInfo{ActorIndex: owner, Element: attributes.Dendro, Durability: 25},
		Pattern: combat.NewCircleHitOnTarget(info.Point{}, nil, 100),
	}, 0)
	advance(c, 1)
	c.QueueAttackEvent(&info.AttackEvent{
		Info:    info.AttackInfo{ActorIndex: owner, Element: attributes.Hydro, Durability: 50},
		Pattern: combat.NewCircleHitOnTarget(info.Point{}, nil, 100),
	}, 0)
	advance(c, dendrocore.Delay+1)
}

func bountifulCores(c *core.Core) []*nilou.BountifulCore {
	var out []*nilou.BountifulCore
	for _, g := range c.Combat.Gadgets() {
		if b, ok := g.(*nilou.BountifulCore); ok {
			out = append(out, b)
		}
	}
	return out
}

func TestBountifulCoreExplodesWithoutDeceit(t *testing.T) {
	c, _ := makeCore(1)
	nl := addChar(t, c, keys.Nilou, 0, 1)
	col := addChar(t, c, keys.Collei, 0, 1)
	start(t, c, nl.Index())
	armNilou(t, c)
	blooms := reactionCount(c, "bloom")
	spawnOwnedCore(c, col.Index())
	advance(c, 40)
	if *blooms == 0 {
		t.Fatal("bountiful core did not explode")
	}
}

func TestBountifulCoreDeceitDoesNotExplode(t *testing.T) {
	c, _ := makeCore(1)
	nl := addChar(t, c, keys.Nilou, 0, 1)
	addChar(t, c, keys.Columbina, 0, 1)
	nef := addChar(t, c, keys.Nefer, 0, 1)
	start(t, c, nl.Index())
	armNilou(t, c)
	c.Player.SetActive(nef.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, skillAnim+1)
	blooms := reactionCount(c, "bloom")
	spawnOwnedCore(c, nef.Index())
	cores := bountifulCores(c)
	if len(cores) != 1 || !cores[0].IsDeceit() {
		t.Fatalf("deceit bountiful cores %d", len(cores))
	}
	advance(c, 40)
	if *blooms != 0 {
		t.Fatalf("deceit bountiful exploded %d", *blooms)
	}
}

func TestAbsorbRefreshesOneVeil(t *testing.T) {
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
	first := c.F
	advance(c, 1)
	n := asNefer(nef)
	if n.veilCount() != 3 || math.Abs(nef.Stat(attributes.EM)-before-100) > 1e-6 {
		t.Fatalf("veils %d em %v before %v", n.veilCount(), nef.Stat(attributes.EM), before)
	}
	advance(c, 400)
	if math.Abs(nef.Stat(attributes.EM)-before-100) > 1e-6 {
		t.Fatal("em fell off before the refresh")
	}
	spawnCore(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	second := c.F
	advance(c, 1)
	old, fresh := 0, 0
	for _, until := range n.veils {
		switch until {
		case first + veilDur:
			old++
		case second + veilDur:
			fresh++
		}
	}
	if n.veilCount() != 3 || old != 2 || fresh != 1 {
		t.Fatalf("refresh old %d fresh %d veils %v frame %d/%d", old, fresh, n.veils, first, second)
	}
	advance(c, 400)
	if math.Abs(nef.Stat(attributes.EM)-before-100) > 1e-6 {
		t.Fatal("refreshed em expired early")
	}
	advance(c, 90)
	if math.Abs(nef.Stat(attributes.EM)-before) > 1e-6 {
		t.Fatalf("em still %v base %v", nef.Stat(attributes.EM), before)
	}
}

func dewFrame(t *testing.T, c *core.Core) int {
	t.Helper()
	for i := 1; i <= 200; i++ {
		advance(c, 1)
		if c.Player.VerdantDew() > 0 {
			return i
		}
	}
	t.Fatal("dew did not generate")
	return -1
}

func TestC4HasteEndsOnSwap(t *testing.T) {
	c, trg := makeCore(1)
	nef := addChar(t, c, keys.Nefer, 4, 1)
	addChar(t, c, keys.Columbina, 0, 1)
	start(t, c, nef.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 1)
	c.Player.OnLunarBloom()
	if err := c.Player.Exec(action.ActionSwap, keys.Columbina, nil); err != nil {
		t.Fatal(err)
	}
	if got := dewFrame(t, c); got != 147 {
		t.Fatalf("dew after swap in %d frames", got)
	}
	if !trg[0].ResistModIsActive("nefer-c4") {
		t.Fatal("c4 shred did not linger after the swap")
	}
}

func TestC4HasteEndsWhenDanceEnds(t *testing.T) {
	c, _ := makeCore(1)
	nef := addChar(t, c, keys.Nefer, 4, 1)
	start(t, c, nef.Index())
	if err := c.Player.Exec(action.ActionSkill, keys.Nefer, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 1)
	asNefer(nef).exitDance()
	if nef.StatusIsActive(shadowKey) {
		t.Fatal("dance still active")
	}
	c.Player.OnLunarBloom()
	if got := dewFrame(t, c); got != 147 {
		t.Fatalf("dew after dance in %d frames", got)
	}
}
