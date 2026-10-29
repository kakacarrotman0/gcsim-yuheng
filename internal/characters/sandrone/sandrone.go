package sandrone

import (
	tmpl "github.com/genshinsim/gcsim/internal/template/character"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

const (
	powerCap = 100.0
	// Marionette_FrostValue_Overload removes itself with compareType LessAndEqual 50.
	// The talent sentence says "under 50". The graph wins: 50 itself exits.
	overdriveExit = 50.0

	// MarionetteTickAddFrostValue, think 0.2s, +3.2 * FrostValueAddRatio.
	gainInterval = 12
	gainPerTick  = 3.2
	// Avatar_Marionette_TriggerBullet, think 1.0s. The +17 is FrostValue_Driver onAbilityStart with opt 0.
	beamInterval = 60
	beamBonus    = 17.0
	// Button-hold charge mixin is 0.18s. 12 frames is the 0.2s machine-gun attach used as the sweeping placeholder.
	sweepDelay = 12
	// TriggerBulletWeak think 0.4s.
	overdriveInterval = 24
	// FrostValue_Driver think 0.5s. -2.5 on-field, -7.5 off-field (BackgroundRatio 1 vs 3).
	decayInterval = 30
	decayOnField  = 2.5
	decayOffField = 7.5
	// UNIQUE_MarionetteForceSet100 removes itself on its 0.25s think.
	forceDelay = 15

	skillCDFrames  = 4 * 60
	burstCDFrames  = 15 * 60
	burstStarSnap  = 17 // SetStarModeDelay 0.28s, rounded
	burstRayAt     = 51 // MarionetteNewBurstAttackerLaser think 0.85s
	burstEnergy    = 60
	particleICD    = 150 // MarionetteNew_Elemball_Proc, 2.5s
	skillTick      = 6   // ReduceFrost think 0.1s
	skillInstant   = 70.0
	skillTickDrain = 10.0
	// Hover state is 6s. Cancel windows below are placeholders, not a frame sheet.
	hoverFrames   = 360
	skillDash     = 12
	skillEarly    = 36
	prismFirst    = 15 // FunnelShooterReal think 0.25s, then the modifier removes itself
	prismStarSnap = 18 // FunnelEff think 0.3s writes FunnelStar before the second bullet
	prismSecond   = 33 // 0.3s + FunnelShooterRealEnhanced think 0.25s
	tacticsWindow = 60 * 60
	a1Threshold   = 50.0
	// Talent text is 400% of the original shot. The graph multiplies 1+P1AttackUpRatio, and that special is 0 in the dump.
	a1PrismMult = 4.0
	a1RayPer    = 0.1

	c1Ratio    = 0.5
	c1Bonus    = 0.30
	c1Duration = 2 * 60
	c1Refresh  = 90 // T1 driver think 1.5s, team buff duration 2s
	c2Base     = 0.40
	c2Per      = 0.20
	c2Max      = 3
	c4CD       = 4 * 60
	c4Delay    = 30 // T4 attack modifier duration 0.5s
	c4Conduct  = 1.25
	c4Swirl    = 1.875
	c6Hits     = 4
	c6Space    = 10 // spacing of bullet 41133008 is not in the avatar file
	c6Ordinary = 1.0
	c6Conduct  = 0.80
	c6Swirl    = 1.20
	c6Elevate  = 0.20

	a4Per100 = 8.0
	a4Cap    = 160.0

	jubileePer100 = 0.007
	jubileeCap    = 0.14

	c1Key       = "sandrone-c1"
	c2Key       = "sandrone-c2"
	c4Key       = "sandrone-c4"
	c4ICDKey    = "sandrone-c4-icd"
	c6Key       = "sandrone-c6"
	stellarKey  = "sandrone-stellar"
	radianceKey = "radiance-stellar-swirl"

	stellarConductText = " (Stellar-Conduct)"
	stellarSwirlText   = " (Stellar Swirl)"

	c4Abil = "Prismatic Resonance Cannon"
)

type fagioMode int

const (
	modeIdle fagioMode = iota
	modeDecoding
	modeOverdrive
)

type radianceState int

const (
	radianceNone radianceState = iota
	radianceStellarConduct
	radianceStellarSwirl
)

type char struct {
	*tmpl.Character
	power        float64
	mode         fagioMode
	beamGen      int
	repairGen    int
	laserCount   int
	c2Stacks     int
	c2Armed      bool
	c2Snap       bool
	frostReduce  float64
	tacticsUntil int
	particleAt   int
	chargeUntil  int
}

func NewChar(s *core.Core, w *character.CharWrapper, _ info.CharacterProfile) error {
	c := char{}
	c.Character = tmpl.NewWithWrapper(s, w)
	c.EnergyMax = burstEnergy
	c.NormalHitNum = normalHitNum
	c.NormalCon = 3
	c.BurstCon = 5
	w.Character = &c
	return nil
}

func (c *char) Init() error {
	c.stellarInit()
	c.a4Init()
	c.c1Init()
	c.c2Init()
	c.c4Init()
	c.c6Init()
	c.Core.Tasks.Add(func() { c.decayTick() }, decayInterval)
	return nil
}

func (c *char) Condition(fields []string) (any, error) {
	switch fields[0] {
	case "power":
		return c.power, nil
	case "mode":
		switch c.mode {
		case modeDecoding:
			return "decoding", nil
		case modeOverdrive:
			return "overdrive", nil
		default:
			return "idle", nil
		}
	case "tactics":
		return c.tactics(), nil
	case "lasers":
		return c.laserCount, nil
	case "c2":
		return c.c2Stacks, nil
	case "radiance":
		switch c.getRadiance() {
		case radianceStellarConduct:
			return "conduct", nil
		case radianceStellarSwirl:
			return "swirl", nil
		default:
			return "none", nil
		}
	default:
		return c.Character.Condition(fields)
	}
}

func (c *char) getRadiance() radianceState {
	if c.StatusIsActive(reactable.PolestarFieldKey) {
		return radianceStellarConduct
	}
	if c.StatusIsActive(radianceKey) {
		return radianceStellarSwirl
	}
	return radianceNone
}

func (c *char) gainRatio() float64 {
	if c.Base.Cons >= 1 {
		return c1Ratio
	}
	return 1
}

func (c *char) addPower(d float64) {
	c.power += d
	if c.power > powerCap {
		c.power = powerCap
	}
	if c.power < 0 {
		c.power = 0
	}
	if c.mode == modeOverdrive && c.power <= overdriveExit {
		c.stopShooting()
	}
}

func (c *char) stopShooting() {
	c.beamGen++
	c.mode = modeIdle
	c.c2Stacks = 0
	c.c2Armed = false
}

func (c *char) onField() bool {
	return c.Core.Player.Active() == c.Index()
}

func (c *char) tactics() int {
	if c.tacticsUntil <= c.Core.F {
		c.frostReduce = 0
		return 0
	}
	n := int(c.frostReduce / 10)
	if n > 10 {
		return 10
	}
	return n
}

// notePowerLoss records Decoding Power that the skill repair actually removes.
// The graph adds the full chunk when power is at least the chunk, and also adds
// the current power when power is at most the chunk. Both actions run at the boundary.
func (c *char) notePowerLoss(chunk float64) {
	if c.power <= 0 || chunk <= 0 {
		return
	}
	if c.power >= chunk {
		c.frostReduce += chunk
	}
	if c.power <= chunk {
		c.frostReduce += c.power
	}
	if c.frostReduce > powerCap {
		c.frostReduce = powerCap
	}
}

// refreshTactics restarts the 60s unique timer.
// Skill does this once, from FrostValue_Driver opt type 1. The 0.1s repair ticks do not.
// Each decay think that is allowed to run applies the timer again.
func (c *char) refreshTactics() {
	c.tacticsUntil = c.Core.F + tacticsWindow
}

func (c *char) particleCB(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy {
		return
	}
	if c.Core.F < c.particleAt {
		return
	}
	c.particleAt = c.Core.F + particleICD
	c.Core.QueueParticle("sandrone", 1, attributes.Cryo, 0)
}

// decodingHorizon is the frame, counted from now, when continuous gain plus condensed bonuses reach 100.
// It follows the 0.2s and 1.0s ability intervals. It is not a measured cancel window.
func (c *char) decodingHorizon() int {
	power := c.power
	ratio := c.gainRatio()
	for f := gainInterval; f <= 20*60; f += gainInterval {
		power += gainPerTick * ratio
		if f%beamInterval == 0 {
			power += beamBonus * ratio
		}
		if power >= powerCap {
			return f
		}
	}
	return 20 * 60
}
