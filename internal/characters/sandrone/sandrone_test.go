package sandrone

import (
	"math"
	"testing"

	_ "github.com/genshinsim/gcsim/internal/characters/alyosha"
	_ "github.com/genshinsim/gcsim/internal/characters/kaeya"
	_ "github.com/genshinsim/gcsim/internal/characters/odette"
	_ "github.com/genshinsim/gcsim/internal/characters/qiqi"
	_ "github.com/genshinsim/gcsim/internal/characters/sucrose"
	_ "github.com/genshinsim/gcsim/internal/characters/vodyanitsa"
	_ "github.com/genshinsim/gcsim/internal/characters/yaemiko"
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

func asSandrone(ch *character.CharWrapper) *char {
	return ch.Character.(*char)
}

func mustInit(t *testing.T, c *core.Core) {
	t.Helper()
	if err := c.Init(); err != nil {
		t.Fatal(err)
	}
}

type logbook struct {
	names []string
	cd    []float64
	mult  []float64
	elev  []float64
}

func (l *logbook) watch(c *core.Core) {
	c.Events.Subscribe(event.OnEnemyDamage, func(args ...any) {
		atk := args[1].(*info.AttackEvent)
		l.names = append(l.names, atk.Info.Abil)
		l.cd = append(l.cd, atk.Snapshot.Stats[attributes.CD])
		l.mult = append(l.mult, atk.Info.Mult)
	}, "sandrone-hits")
	c.Events.Subscribe(event.OnApplyAttack, func(args ...any) {
		atk := args[0].(*info.AttackEvent)
		l.elev = append(l.elev, atk.Info.Elevation)
	}, "sandrone-elev")
}

func (l *logbook) count(substr string) int {
	n := 0
	for _, name := range l.names {
		if name == substr {
			n++
		}
	}
	return n
}

func (l *logbook) countPrefix(prefix string) int {
	n := 0
	for _, name := range l.names {
		if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
			n++
		}
	}
	return n
}

func hold(c *core.Core, next action.Action) int {
	n := 0
	for c.Player.IsAnimationLocked(next) {
		advance(c, 1)
		n++
		if n > 2000 {
			return n
		}
	}
	return n
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

func TestDecodingOverdriveAndRecovery(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 0, 1)
	mustInit(t, c)
	var book logbook
	book.watch(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 180)
	if asSandrone(ch).laserCount != 3 || asSandrone(ch).mode != modeDecoding {
		t.Fatalf("at 180 lasers %d mode %v power %v", asSandrone(ch).laserCount, asSandrone(ch).mode, asSandrone(ch).power)
	}
	if math.Abs(asSandrone(ch).power-99) > 0.05 {
		t.Fatalf("power before the cap tick %v", asSandrone(ch).power)
	}
	advance(c, 12)
	s := asSandrone(ch)
	if s.mode != modeOverdrive || s.power != 100 {
		t.Fatalf("overdrive mode %v power %v", s.mode, s.power)
	}
	if book.count("Sweeping Fire") != 1 || book.count("Condensed Beam") != 3 {
		t.Fatalf("names %v", book.names)
	}
	over := book.count("Overdrive")
	advance(c, 48)
	if book.count("Overdrive") <= over {
		t.Fatalf("overdrive did not continue: %v", book.names)
	}
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, nil); err == nil {
		t.Fatal("charge during overdrive should fail")
	}
	// On-field decay is 2.5 every 30 frames. From 100, exit is not this window.
	// Drive the exit directly from just above 50.
	s.power = 52.5
	s.mode = modeOverdrive
	advance(c, 30)
	if s.mode != modeIdle || math.Abs(s.power-overdriveExit) > 1e-9 {
		t.Fatalf("exit mode %v power %v", s.mode, s.power)
	}
}

func TestC1SlowerGain(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 1, 1)
	mustInit(t, c)
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 360)
	s := asSandrone(ch)
	if s.laserCount != 6 || s.mode != modeDecoding {
		t.Fatalf("c1 at 360 lasers %d mode %v power %v", s.laserCount, s.mode, s.power)
	}
	if !ch.StatusIsActive(c1Key) {
		t.Fatal("c1 buff missing during decoding")
	}
	advance(c, 12)
	if s.mode != modeOverdrive || s.power != 100 {
		t.Fatalf("c1 cap mode %v power %v", s.mode, s.power)
	}
}

func TestSkillRepairAndCooldown(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 0, 1)
	mustInit(t, c)
	asSandrone(ch).power = 100
	asSandrone(ch).mode = modeOverdrive
	if err := c.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	s := asSandrone(ch)
	if math.Abs(s.power-30) > 1e-9 || s.mode != modeIdle {
		t.Fatalf("instant repair power %v mode %v", s.power, s.mode)
	}
	if ch.Cooldown(action.ActionSkill) != skillCDFrames {
		t.Fatalf("cd %v", ch.Cooldown(action.ActionSkill))
	}
	held := hold(c, action.ActionAttack)
	if held != skillEarly {
		t.Fatalf("attack cancel opened at %d", held)
	}
	if err := c.Player.Exec(action.ActionAttack, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	if s.power != 0 || s.tactics() != 10 {
		t.Fatalf("full repair power %v tactics %d reduce %v", s.power, s.tactics(), s.frostReduce)
	}
	advance(c, skillCDFrames)
	if ch.Cooldown(action.ActionSkill) != 0 {
		t.Fatalf("cd left %v", ch.Cooldown(action.ActionSkill))
	}
}

func TestEarlySkillCancelLeavesPower(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 0, 1)
	mustInit(t, c)
	asSandrone(ch).power = 100
	if err := c.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	held := hold(c, action.ActionDash)
	if held != skillDash {
		t.Fatalf("dash cancel opened at %d", held)
	}
	if err := c.Player.Exec(action.ActionDash, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	// Initial 70 plus the 0.1s ticks on frames 6 and 12. Frame 12's tick runs before the dash.
	if math.Abs(asSandrone(ch).power-10) > 1e-9 {
		t.Fatalf("early cancel power %v", asSandrone(ch).power)
	}
	// Repair is state-attached. The next decay is frame 30, so frame 29 still shows the cancelled bar.
	advance(c, 17)
	if c.F != 29 || math.Abs(asSandrone(ch).power-10) > 1e-9 {
		t.Fatalf("repair continued after dash: f %d power %v", c.F, asSandrone(ch).power)
	}
	// 10 remaining still reaches the third beam. 30 is the bar after the instant 70 and before any 0.1s tick.
	// That bar caps during the second second, so the third condensed beam does not fire.
	asSandrone(ch).power = 30
	asSandrone(ch).mode = modeIdle
	var book logbook
	book.watch(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 200)
	if asSandrone(ch).laserCount != 2 {
		t.Fatalf("partial bar produced %d beams, power %v", asSandrone(ch).laserCount, asSandrone(ch).power)
	}
}

func TestCondensedRadianceTags(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 0, 1)
	mustInit(t, c)
	var book logbook
	book.watch(c)
	ch.AddStatus(reactable.PolestarFieldKey, 600, true)
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 60)
	if book.count("Condensed Beam"+stellarConductText) != 1 || book.count("Condensed Beam") != 0 {
		t.Fatalf("conduct names %v", book.names)
	}
	ch.DeleteStatus(reactable.PolestarFieldKey)
	ch.AddStatus(radianceKey, 600, true)
	advance(c, 60)
	if book.count("Condensed Beam"+stellarSwirlText) != 1 {
		t.Fatalf("swirl names %v", book.names)
	}
}

func TestReactionsWithAndWithoutSandrone(t *testing.T) {
	plain, ptrg := makeCore(1)
	k := addProfile(t, plain, keys.Kaeya, 0, 1)
	addProfile(t, plain, keys.Sucrose, 0, 1)
	mustInit(t, plain)
	var sc, swirl int
	plain.Events.Subscribe(event.OnSuperconduct, func(...any) { sc++ }, "sc")
	plain.Events.Subscribe(event.OnSwirlCryo, func(...any) { swirl++ }, "sw")
	ptrg[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Electro, Durability: 25}})
	advance(plain, 1)
	gauge(plain, ptrg[0], k.Index(), attributes.Cryo)
	ptrg[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25}})
	advance(plain, 1)
	gauge(plain, ptrg[0], k.Index(), attributes.Anemo)
	if sc != 1 || swirl != 1 {
		t.Fatalf("plain reactions sc %d swirl %d", sc, swirl)
	}

	live, ltrg := makeCore(1)
	sk := addProfile(t, live, keys.Sandrone, 0, 1)
	mustInit(t, live)
	var stellar, plainSC, plainSwirl int
	live.Events.Subscribe(event.OnStellarConduct, func(...any) { stellar++ }, "ssc")
	live.Events.Subscribe(event.OnSuperconduct, func(...any) { plainSC++ }, "sc")
	live.Events.Subscribe(event.OnStellarSwirl, func(...any) { stellar++ }, "ssw")
	live.Events.Subscribe(event.OnSwirlCryo, func(...any) { plainSwirl++ }, "sw")
	ltrg[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Electro, Durability: 25}})
	advance(live, 1)
	gauge(live, ltrg[0], sk.Index(), attributes.Cryo)
	ltrg[0].AttachOrRefill(&info.AttackEvent{Info: info.AttackInfo{Element: attributes.Cryo, Durability: 25}})
	advance(live, 1)
	gauge(live, ltrg[0], sk.Index(), attributes.Anemo)
	if stellar != 2 || plainSC != 0 || plainSwirl != 0 {
		t.Fatalf("stellar %d superconduct %d swirl %d", stellar, plainSC, plainSwirl)
	}
	if live.Flags.Custom[reactable.StellarConductEnableKey] != 1 || live.Flags.Custom[reactable.StellarSwirlEnableKey] != 1 {
		t.Fatal("flags were not enabled")
	}
}

func TestSharedParticles(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 0, 1)
	mustInit(t, c)
	particles := 0
	c.Events.Subscribe(event.OnParticleReceived, func(...any) { particles++ }, "p")
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 180)
	if particles != 1 {
		t.Fatalf("beams should share one particle window, got %d", particles)
	}
	before := particles
	advance(c, 12) // enter overdrive and fire the attach hit
	if particles != before {
		t.Fatalf("overdrive generated a particle")
	}
	hold(c, action.ActionSkill)
	if err := c.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, prismFirst)
	if particles != before {
		t.Fatalf("skill prism inside the beam ICD generated a particle, count %d", particles)
	}
	advance(c, prismSecond-prismFirst)
	if particles != before+1 {
		t.Fatalf("later prism particles %d", particles)
	}
	_ = ch
}

func TestBurstEnergyAndRadiance(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 0, 1)
	mustInit(t, c)
	var book logbook
	book.watch(c)
	if ch.Energy != 60 {
		t.Fatalf("start energy %v", ch.Energy)
	}
	if err := c.Player.Exec(action.ActionBurst, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	if ch.Energy != 0 {
		t.Fatalf("energy %v", ch.Energy)
	}
	advance(c, 51)
	if book.count("Bombardment") != 3 || book.count("Convective Inhibition Ray") != 1 {
		t.Fatalf("burst names %v", book.names)
	}

	c2, _ := makeCore(1)
	ch2 := addProfile(t, c2, keys.Sandrone, 0, 1)
	mustInit(t, c2)
	s := asSandrone(ch2)
	s.frostReduce = 100
	s.tacticsUntil = c2.F + 100
	ch2.AddStatus(reactable.PolestarFieldKey, 300, true)
	var ray logbook
	ray.watch(c2)
	if err := c2.Player.Exec(action.ActionBurst, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	if s.tactics() != 0 {
		t.Fatalf("stacks left %d", s.tactics())
	}
	advance(c2, 51)
	want := burstRayConduct[0] * (1 + a1RayPer*10)
	if ray.count("Convective Inhibition Ray"+stellarConductText) != 1 {
		t.Fatalf("ray names %v", ray.names)
	}
	got := ray.mult[len(ray.mult)-1]
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("ray mult %v want %v", got, want)
	}
}

func TestA1StacksAndPrism(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 0, 1)
	mustInit(t, c)
	s := asSandrone(ch)
	s.power = 100
	var book logbook
	book.watch(c)
	ch.AddStatus(reactable.PolestarFieldKey, 600, true)
	if err := c.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, prismSecond)
	if book.count("Prism Shot") != 1 || book.count("Prism Shot"+stellarConductText) != 1 {
		t.Fatalf("prisms %v", book.names)
	}
	if math.Abs(book.mult[0]-skillOrdinary[0]) > 1e-9 {
		t.Fatalf("first prism %v", book.mult[0])
	}
	if math.Abs(book.mult[1]-skillConduct[0]*a1PrismMult) > 1e-9 {
		t.Fatalf("boosted prism %v", book.mult[1])
	}
	hold(c, action.ActionAttack)
	if s.tactics() != 10 {
		t.Fatalf("tactics %d reduce %v", s.tactics(), s.frostReduce)
	}
	left := s.tacticsUntil - c.F
	advance(c, left)
	if s.tactics() != 0 {
		t.Fatalf("stacks survived expiry %d", s.tactics())
	}

	low, _ := makeCore(1)
	lowCh := addProfile(t, low, keys.Sandrone, 0, 1)
	mustInit(t, low)
	asSandrone(lowCh).power = 40
	var lowBook logbook
	lowBook.watch(low)
	if err := low.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(low, prismSecond)
	if math.Abs(lowBook.mult[1]-skillOrdinary[0]) > 1e-9 {
		t.Fatalf("unboosted second prism %v", lowBook.mult[1])
	}
}

func TestA4CapDoesNotRecurse(t *testing.T) {
	low, _ := makeCore(1)
	lowCh := addProfileAt(t, low, keys.Sandrone, 0, 1, 20, 20)
	mustInit(t, low)
	if math.Abs(lowCh.Stat(attributes.EM)-100) > 1e-9 {
		t.Fatalf("pre-a4 em %v", lowCh.Stat(attributes.EM))
	}

	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 0, 1)
	mustInit(t, c)
	ch.BaseStats[attributes.ATK] += 10000
	before := ch.TotalAtk()
	em := ch.Stat(attributes.EM)
	if math.Abs(em-(100+a4Cap)) > 1e-6 {
		t.Fatalf("capped em %v", em)
	}
	ch.BaseStats[attributes.EM] += 500
	if ch.TotalAtk() != before {
		t.Fatalf("em changed atk %v -> %v", before, ch.TotalAtk())
	}
	if math.Abs(ch.Stat(attributes.EM)-(100+500+a4Cap)) > 1e-6 {
		t.Fatalf("em after extra %v", ch.Stat(attributes.EM))
	}
}

func TestConstellationLevels(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 6, 1)
	mustInit(t, c)
	if ch.TalentLvlAttack() != 3 || ch.TalentLvlBurst() != 3 || ch.TalentLvlSkill() != 0 {
		t.Fatalf("talents na %d skill %d burst %d", ch.TalentLvlAttack(), ch.TalentLvlSkill(), ch.TalentLvlBurst())
	}
}

func TestC2CritOnCondensedBeams(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 2, 1)
	mustInit(t, c)
	base := ch.Stat(attributes.CD)
	ch.AddStatus(reactable.PolestarFieldKey, 900, true)
	var book logbook
	book.watch(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 180)
	var got []float64
	for i, name := range book.names {
		if name == "Condensed Beam"+stellarConductText {
			got = append(got, book.cd[i])
		}
	}
	want := []float64{base + 0.60, base + 0.80, base + 1.00}
	if len(got) != 3 {
		t.Fatalf("beams %v names %v", got, book.names)
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			t.Fatalf("beam %d cd %v want %v", i+1, got[i], want[i])
		}
	}

	plain, _ := makeCore(1)
	plainCh := addProfile(t, plain, keys.Sandrone, 2, 1)
	mustInit(t, plain)
	base = plainCh.Stat(attributes.CD)
	var plainBook logbook
	plainBook.watch(plain)
	if err := plain.Player.Exec(action.ActionCharge, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(plain, 60)
	for i, name := range plainBook.names {
		if name == "Condensed Beam" && math.Abs(plainBook.cd[i]-base) > 1e-9 {
			t.Fatalf("ordinary beam cd %v base %v", plainBook.cd[i], base)
		}
	}
}

func TestC4CooldownDoesNotRecurse(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 4, 1)
	mustInit(t, c)
	var book logbook
	book.watch(c)
	probe := info.AttackInfo{
		ActorIndex:       ch.Index(),
		Abil:             "probe stellar",
		AttackTag:        attacks.AttackTagDirectStellarConduct,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		Element:          attributes.Cryo,
		Durability:       0,
		Mult:             1,
		IgnoreDefPercent: 1,
	}
	ap := combat.NewCircleHitOnTarget(c.Combat.Player(), nil, 2)
	c.QueueAttack(probe, ap, 0, 0)
	c.QueueAttack(probe, ap, 0, 0)
	advance(c, c4Delay+1)
	if book.countPrefix(c4Abil) != 1 {
		t.Fatalf("cannons %v", book.names)
	}
	advance(c, c4CD)
	c.QueueAttack(probe, ap, 0, 0)
	advance(c, c4Delay+1)
	if book.countPrefix(c4Abil) != 2 {
		t.Fatalf("second window cannons %v", book.names)
	}
}

func TestC6ClusterAndElevation(t *testing.T) {
	c, _ := makeCore(1)
	addProfile(t, c, keys.Sandrone, 6, 1)
	mustInit(t, c)
	var book logbook
	book.watch(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 120)
	if book.count("Condensed Cluster Beam") != 0 || book.count("Condensed Beam") != 2 {
		t.Fatalf("early names %v", book.names)
	}
	advance(c, 100)
	if book.count("Condensed Cluster Beam") != 4 || book.count("Condensed Beam") != 3 {
		t.Fatalf("cluster names %v", book.names)
	}

	rad, _ := makeCore(1)
	radCh := addProfile(t, rad, keys.Sandrone, 6, 1)
	mustInit(t, rad)
	radCh.AddStatus(reactable.PolestarFieldKey, 900, true)
	var radBook logbook
	radBook.watch(rad)
	if err := rad.Player.Exec(action.ActionCharge, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(rad, 220)
	if radBook.count("Condensed Cluster Beam"+stellarConductText) != 4 || radBook.count("Condensed Cluster Beam") != 0 {
		t.Fatalf("radiance cluster %v", radBook.names)
	}
	for i, name := range radBook.names {
		if name == "Condensed Beam"+stellarConductText || name == "Condensed Cluster Beam"+stellarConductText {
			if math.Abs(radBook.elev[i]-c6Elevate) > 1e-9 {
				t.Fatalf("%s elevation %v", name, radBook.elev[i])
			}
		}
	}
}

func TestSwapDecayAndSecondRotation(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 0, 1)
	addProfile(t, c, keys.Kaeya, 0, 1)
	mustInit(t, c)
	s := asSandrone(ch)
	s.power = 100
	advance(c, 30)
	if math.Abs(s.power-(100-decayOnField)) > 1e-9 {
		t.Fatalf("on-field decay %v", s.power)
	}
	if err := c.Player.Exec(action.ActionSwap, keys.Kaeya, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 30)
	if math.Abs(s.power-(100-decayOnField-decayOffField)) > 1e-9 {
		t.Fatalf("off-field decay %v", s.power)
	}
	if err := c.Player.Exec(action.ActionSwap, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 1)
	s.power = 0
	s.mode = modeIdle
	var book logbook
	book.watch(c)
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 192)
	if book.count("Condensed Beam") != 3 {
		t.Fatalf("first rotation %v", book.names)
	}
	if err := c.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	hold(c, action.ActionAttack)
	if err := c.Player.Exec(action.ActionAttack, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	if s.power != 0 {
		t.Fatalf("repair left %v", s.power)
	}
	book.names = nil
	if err := c.Player.Exec(action.ActionCharge, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, 192)
	if book.count("Condensed Beam") != 3 {
		t.Fatalf("second rotation %v", book.names)
	}
}

func TestTeammateCompatibility(t *testing.T) {
	c, trg := makeCore(1)
	sand := addProfile(t, c, keys.Sandrone, 0, 10)
	od := addProfile(t, c, keys.Odette, 0, 10)
	addProfile(t, c, keys.YaeMiko, 0, 10)
	addProfile(t, c, keys.Alyosha, 0, 10)
	// A four-character party is the sim limit. Vodyanitsa's duration hook is the public flag.
	mustInit(t, c)
	_ = addProfile
	if c.Flags.Custom[reactable.StellarConductEnableKey] != 1 || c.Flags.Custom[reactable.StellarSwirlEnableKey] != 1 {
		t.Fatal("enable flags")
	}
	c.Flags.Custom[reactable.VodyanitsaSongUntilKey] = float64(c.F + 9999)
	c.Events.Emit(event.OnStellarSwirl, trg[0], &info.AttackEvent{})
	if sand.StatusDuration(radianceKey) != reactable.RadianceStellarSwirlBase+reactable.RadianceStellarSwirlExtend {
		t.Fatalf("extended radiance %d", sand.StatusDuration(radianceKey))
	}
	sand.DeleteStatus(radianceKey)
	delete(c.Flags.Custom, reactable.VodyanitsaSongUntilKey)
	c.Events.Emit(event.OnStellarSwirl, trg[0], &info.AttackEvent{})
	if sand.StatusDuration(radianceKey) != reactable.RadianceStellarSwirlBase {
		t.Fatalf("base radiance %d", sand.StatusDuration(radianceKey))
	}
	sand.AddStatus(reactable.PolestarFieldKey, 300, true)
	if asSandrone(sand).getRadiance() != radianceStellarConduct {
		t.Fatal("polestar should win")
	}

	atk := &info.AttackEvent{Info: info.AttackInfo{AttackTag: attacks.AttackTagDirectStellarConduct}}
	c.Events.Emit(event.OnEnemyHit, trg[0], atk)
	want := min(sand.TotalAtk()/100*jubileePer100, jubileeCap) + min(od.TotalAtk()/100*0.007, 0.14)
	if math.Abs(atk.Info.BaseDmgBonus-want) > 1e-9 {
		t.Fatalf("jubilee bonus %v want %v", atk.Info.BaseDmgBonus, want)
	}
}

func TestOverdriveExitSchedule(t *testing.T) {
	exitAt := func(t *testing.T, start float64, wantPower float64, wantIdle bool) {
		t.Helper()
		c, _ := makeCore(1)
		ch := addProfile(t, c, keys.Sandrone, 0, 1)
		mustInit(t, c)
		s := asSandrone(ch)
		var book logbook
		book.watch(c)
		s.enterOverdrive()
		s.power = start
		advance(c, overdriveInterval)
		if s.mode != modeOverdrive || book.count("Overdrive") != 2 {
			t.Fatalf("start %v before exit mode %v hits %v", start, s.mode, book.names)
		}
		advance(c, decayInterval-overdriveInterval)
		if wantIdle {
			if s.mode != modeIdle || math.Abs(s.power-wantPower) > 1e-9 {
				t.Fatalf("start %v exit mode %v power %v", start, s.mode, s.power)
			}
			advance(c, overdriveInterval)
			if book.count("Overdrive") != 2 {
				t.Fatalf("start %v shot after exit %v", start, book.names)
			}
			return
		}
		if s.mode != modeOverdrive || math.Abs(s.power-wantPower) > 1e-9 {
			t.Fatalf("start %v still overdrive mode %v power %v", start, s.mode, s.power)
		}
		advance(c, overdriveInterval)
		if book.count("Overdrive") != 3 {
			t.Fatalf("start %v missing follow-up shot %v", start, book.names)
		}
	}
	// 52.5 - 2.5 lands on 50. LessAndEqual exits, and the next 0.4s shot does not fire.
	exitAt(t, 52.5, 50, true)
	// 52 - 2.5 lands at 49.5 and exits on that same tick.
	exitAt(t, 52, 49.5, true)
	// 55 - 2.5 stays above 50, so the following Overdrive shot still fires.
	exitAt(t, 55, 52.5, false)
}

func TestTacticsDecayBoundaryAndRefresh(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 0, 1)
	mustInit(t, c)
	s := asSandrone(ch)
	s.power = 70
	s.notePowerLoss(70)
	if s.frostReduce != 100 {
		t.Fatalf("boundary reduce %v", s.frostReduce)
	}

	s.frostReduce = 0
	s.power = 2.5
	s.mode = modeIdle
	advance(c, decayInterval)
	if s.frostReduce != 5 || s.power != 0 || s.tacticsUntil != decayInterval+tacticsWindow {
		t.Fatalf("decay boundary reduce %v power %v until %d", s.frostReduce, s.power, s.tacticsUntil)
	}

	c2, _ := makeCore(1)
	ch2 := addProfile(t, c2, keys.Sandrone, 0, 1)
	mustInit(t, c2)
	s2 := asSandrone(ch2)
	s2.power = 100
	if err := c2.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	if s2.tacticsUntil != tacticsWindow {
		t.Fatalf("skill timer %d", s2.tacticsUntil)
	}
	hold(c2, action.ActionAttack)
	if s2.tacticsUntil != tacticsWindow {
		t.Fatalf("repair ticks refreshed the timer to %d", s2.tacticsUntil)
	}
	if s2.tactics() != 10 {
		t.Fatalf("tactics %d", s2.tactics())
	}
	advance(c2, s2.tacticsUntil-c2.F)
	if s2.tactics() != 0 {
		t.Fatalf("timer expiry tactics %d", s2.tactics())
	}

	c3, _ := makeCore(1)
	ch3 := addProfile(t, c3, keys.Sandrone, 0, 1)
	mustInit(t, c3)
	s3 := asSandrone(ch3)
	s3.power = 0
	s3.frostReduce = 20
	s3.tacticsUntil = 10
	if err := c3.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	if s3.tactics() != 2 || s3.tacticsUntil != tacticsWindow {
		t.Fatalf("empty skill tactics %d until %d", s3.tactics(), s3.tacticsUntil)
	}
}

func TestBurstClearsTacticsInsideAndOutsideRadiance(t *testing.T) {
	outside, _ := makeCore(1)
	ch := addProfile(t, outside, keys.Sandrone, 0, 1)
	mustInit(t, outside)
	s := asSandrone(ch)
	s.frostReduce = 100
	s.tacticsUntil = outside.F + 100
	var book logbook
	book.watch(outside)
	if err := outside.Player.Exec(action.ActionBurst, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	if s.tactics() != 0 {
		t.Fatalf("outside radiance kept tactics %d", s.tactics())
	}
	advance(outside, burstRayAt)
	if book.count("Convective Inhibition Ray") != 1 || book.count("Convective Inhibition Ray"+stellarConductText) != 0 {
		t.Fatalf("outside ray %v", book.names)
	}
	if math.Abs(book.mult[len(book.mult)-1]-burstRayOrdinary[0]) > 1e-9 {
		t.Fatalf("outside mult %v", book.mult[len(book.mult)-1])
	}

	late, _ := makeCore(1)
	lateCh := addProfile(t, late, keys.Sandrone, 0, 1)
	mustInit(t, late)
	lateS := asSandrone(lateCh)
	lateS.frostReduce = 100
	lateS.tacticsUntil = 100
	var lateBook logbook
	lateBook.watch(late)
	if err := late.Player.Exec(action.ActionBurst, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(late, 10)
	lateCh.AddStatus(reactable.PolestarFieldKey, 300, true)
	advance(late, burstRayAt-10)
	want := burstRayConduct[0] * (1 + a1RayPer*10)
	if lateBook.count("Convective Inhibition Ray"+stellarConductText) != 1 {
		t.Fatalf("late radiance ray %v", lateBook.names)
	}
	if math.Abs(lateBook.mult[len(lateBook.mult)-1]-want) > 1e-6 {
		t.Fatalf("late mult %v want %v", lateBook.mult[len(lateBook.mult)-1], want)
	}

	drop, _ := makeCore(1)
	dropCh := addProfile(t, drop, keys.Sandrone, 0, 1)
	mustInit(t, drop)
	dropS := asSandrone(dropCh)
	dropS.frostReduce = 100
	dropS.tacticsUntil = 100
	dropCh.AddStatus(reactable.PolestarFieldKey, 300, true)
	var dropBook logbook
	dropBook.watch(drop)
	if err := drop.Player.Exec(action.ActionBurst, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	if dropS.tactics() != 0 {
		t.Fatal("cast did not clear stacks")
	}
	advance(drop, 10)
	dropCh.DeleteStatus(reactable.PolestarFieldKey)
	advance(drop, burstRayAt-10)
	if dropBook.count("Convective Inhibition Ray") != 1 || dropBook.count("Convective Inhibition Ray"+stellarConductText) != 0 {
		t.Fatalf("dropped radiance ray %v", dropBook.names)
	}
}

func TestPrismSurvivesSkillCancel(t *testing.T) {
	c, _ := makeCore(1)
	ch := addProfile(t, c, keys.Sandrone, 0, 1)
	mustInit(t, c)
	asSandrone(ch).power = 100
	ch.AddStatus(reactable.PolestarFieldKey, 600, true)
	particles := 0
	c.Events.Subscribe(event.OnParticleReceived, func(...any) { particles++ }, "p")
	var book logbook
	book.watch(c)
	if err := c.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	hold(c, action.ActionDash)
	if err := c.Player.Exec(action.ActionDash, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(c, prismSecond-c.F)
	if book.count("Prism Shot") != 1 || book.count("Prism Shot"+stellarConductText) != 1 {
		t.Fatalf("cancelled skill prisms %v", book.names)
	}
	if math.Abs(book.mult[1]-skillConduct[0]*a1PrismMult) > 1e-9 {
		t.Fatalf("cancelled boost %v", book.mult[1])
	}
	if particles != 1 {
		t.Fatalf("prism particles %d", particles)
	}

	plain, _ := makeCore(1)
	plainCh := addProfile(t, plain, keys.Sandrone, 0, 1)
	mustInit(t, plain)
	asSandrone(plainCh).power = 100
	var plainBook logbook
	plainBook.watch(plain)
	if err := plain.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	advance(plain, prismStarSnap)
	plainCh.AddStatus(reactable.PolestarFieldKey, 600, true)
	advance(plain, prismSecond-prismStarSnap)
	if plainBook.count("Prism Shot 2") != 1 || plainBook.count("Prism Shot"+stellarConductText) != 0 {
		t.Fatalf("late polestar changed the snapped shot %v", plainBook.names)
	}

	full, _ := makeCore(1)
	fullCh := addProfile(t, full, keys.Sandrone, 0, 1)
	mustInit(t, full)
	asSandrone(fullCh).power = 100
	fullCh.AddStatus(reactable.PolestarFieldKey, 600, true)
	var fullBook logbook
	fullBook.watch(full)
	if err := full.Player.Exec(action.ActionSkill, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	hold(full, action.ActionAttack)
	if err := full.Player.Exec(action.ActionAttack, keys.Sandrone, nil); err != nil {
		t.Fatal(err)
	}
	if fullBook.count("Prism Shot") != 1 || fullBook.count("Prism Shot"+stellarConductText) != 1 {
		t.Fatalf("attack cancel prisms %v", fullBook.names)
	}
}

func TestQiqiStillConstructs(t *testing.T) {
	c, _ := makeCore(1)
	addProfile(t, c, keys.Sandrone, 0, 1)
	addProfile(t, c, keys.Qiqi, 0, 1)
	addProfile(t, c, keys.Vodyanitsa, 0, 1)
	mustInit(t, c)
}
