package sandrone

import (
	"errors"

	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// ChargeAttack holds the field until Decoding Power would reach 100.
// The horizon is derived from the 0.2s gain tick and the 1.0s condensed beam.
// It is not a measured animation cancel. From 0 power that is frame 192 at C0 and frame 372 at C1.
// The sim ticks the frame before it runs the next action, so the capping tick enters Overdrive
// and then Skill can repair on that same frame.
func (c *char) ChargeAttack(_ map[string]int) (action.Info, error) {
	if c.mode == modeOverdrive {
		return action.Info{}, errors.New("fagio is in power overdrive and cannot enter decoding")
	}
	horizon := c.decodingHorizon()
	c.chargeUntil = c.Core.F + horizon
	c.enterDecoding()
	return action.Info{
		Frames: func(action.Action) int {
			return horizon
		},
		AnimationLength: horizon + gainInterval,
		CanQueueAfter:   horizon,
		State:           action.ChargeAttackState,
		OnRemoved: func(action.AnimationState) {
			// A replacement before the cap, such as a dash, ends Decoding.
			// Overdrive is left running. Skill is what repairs it.
			if c.mode == modeDecoding {
				c.stopShooting()
			}
		},
	}, nil
}

func (c *char) enterDecoding() {
	c.stopShooting()
	c.mode = modeDecoding
	c.laserCount = 0
	gen := c.beamGen
	if c.Base.Cons >= 1 {
		c.AddStatus(c1Key, c1Duration, true)
		c.Core.Tasks.Add(func() { c.c1Loop(gen) }, c1Refresh)
	}
	c.Core.Tasks.Add(func() { c.gainTick(gen) }, gainInterval)
	c.Core.Tasks.Add(func() { c.sweeping(gen) }, sweepDelay)
	c.Core.Tasks.Add(func() { c.condensed(gen) }, beamInterval)
}

func (c *char) c1Loop(gen int) {
	if c.Base.Cons < 1 || c.beamGen != gen || c.mode != modeDecoding {
		return
	}
	c.AddStatus(c1Key, c1Duration, true)
	c.Core.Tasks.Add(func() { c.c1Loop(gen) }, c1Refresh)
}

func (c *char) gainTick(gen int) {
	if c.beamGen != gen || c.mode != modeDecoding {
		return
	}
	c.addPower(gainPerTick * c.gainRatio())
	if c.power >= powerCap {
		c.enterOverdrive()
		return
	}
	c.Core.Tasks.Add(func() { c.gainTick(gen) }, gainInterval)
}

func (c *char) sweeping(gen int) {
	if c.beamGen != gen || c.mode != modeDecoding {
		return
	}
	ai := c.cryoInfo("Sweeping Fire", attacks.AttackTagExtra, attacks.ICDTagExtraAttack, attacks.ICDGroupSandroneMachineGun, sweeping[c.TalentLvlAttack()], 25)
	c.queue(ai, 3.5, 0, false)
}

func (c *char) condensed(gen int) {
	if c.beamGen != gen || c.mode != modeDecoding {
		return
	}
	c.laserCount++
	count := c.laserCount
	c.fireCondensed(count)
	c.addPower(beamBonus * c.gainRatio())
	if c.power >= powerCap {
		c.enterOverdrive()
		return
	}
	// Backup cap. The 0.25s force modifier adds 100 when the count is exactly 3, or 6 at C1.
	// Natural gain reaches 100 first when the beam starts from empty.
	need := 3
	if c.Base.Cons >= 1 {
		need = 6
	}
	if count == need {
		c.Core.Tasks.Add(func() {
			if c.beamGen != gen || c.mode != modeDecoding || c.laserCount != need {
				return
			}
			c.power = powerCap
			c.enterOverdrive()
		}, forceDelay)
	}
	c.Core.Tasks.Add(func() { c.condensed(gen) }, beamInterval)
}

func (c *char) fireCondensed(count int) {
	lvl := c.TalentLvlAttack()
	rad := c.getRadiance()
	c.armC2(rad)
	defer func() { c.c2Snap = false }()

	switch rad {
	case radianceStellarConduct:
		ai := c.stellarInfo("Condensed Beam"+stellarConductText, attacks.AttackTagDirectStellarConduct, condensedConduct[lvl])
		c.queue(ai, 4, 0, true)
	case radianceStellarSwirl:
		ai := c.stellarInfo("Condensed Beam"+stellarSwirlText, attacks.AttackTagDirectStellarSwirl, condensedSwirl[lvl])
		c.queue(ai, 4, 0, true)
	default:
		ai := c.cryoInfo("Condensed Beam", attacks.AttackTagExtra, attacks.ICDTagExtraAttack, attacks.ICDGroupDefault, condensedOrdinary[lvl], 25)
		c.queue(ai, 4, 0, true)
	}
	if c.Base.Cons >= 6 && count >= 3 {
		c.cluster(rad)
	}
}

func (c *char) cluster(rad radianceState) {
	for i := 1; i <= c6Hits; i++ {
		delay := i * c6Space
		switch rad {
		case radianceStellarConduct:
			ai := c.stellarInfo("Condensed Cluster Beam"+stellarConductText, attacks.AttackTagDirectStellarConduct, c6Conduct)
			c.queue(ai, 4, delay, true)
		case radianceStellarSwirl:
			ai := c.stellarInfo("Condensed Cluster Beam"+stellarSwirlText, attacks.AttackTagDirectStellarSwirl, c6Swirl)
			c.queue(ai, 4, delay, true)
		default:
			// Ordinary cluster uses MarionetteExtrdaAttackLaserNormalT6, durability 25.
			// That tag is not in MarionetteNew_Elemball_Driver, so it does not drop a particle.
			ai := c.cryoInfo("Condensed Cluster Beam", attacks.AttackTagExtra, attacks.ICDTagExtraAttack, attacks.ICDGroupDefault, c6Ordinary, 25)
			c.queue(ai, 4, delay, false)
		}
	}
}

func (c *char) armC2(rad radianceState) {
	if c.Base.Cons < 2 || rad == radianceNone {
		c.c2Armed = false
		c.c2Stacks = 0
		return
	}
	if !c.c2Armed {
		c.c2Stacks = 0
		c.c2Armed = true
	}
	if c.c2Stacks < c2Max {
		c.c2Stacks++
	}
	c.c2Snap = true
}

func (c *char) enterOverdrive() {
	if c.mode == modeOverdrive {
		return
	}
	c.power = powerCap
	c.mode = modeOverdrive
	c.beamGen++
	c.c2Stacks = 0
	c.c2Armed = false
	gen := c.beamGen
	// TriggerBulletWeak fires once on attach, then every 0.4s.
	c.overdriveHit()
	c.Core.Tasks.Add(func() { c.overdriveLoop(gen) }, overdriveInterval)
}

func (c *char) overdriveLoop(gen int) {
	if c.beamGen != gen || c.mode != modeOverdrive {
		return
	}
	c.overdriveHit()
	c.Core.Tasks.Add(func() { c.overdriveLoop(gen) }, overdriveInterval)
}

func (c *char) overdriveHit() {
	ai := c.cryoInfo("Overdrive", attacks.AttackTagExtra, attacks.ICDTagExtraAttack, attacks.ICDGroupSandroneMachineGun, overdrive[c.TalentLvlAttack()], 25)
	c.queue(ai, 3.5, 0, false)
}

func (c *char) decayTick() {
	c.Core.Tasks.Add(func() { c.decayTick() }, decayInterval)
	// The driver does not tick during Decoding or during the Skill's own repair.
	if c.mode == modeDecoding {
		return
	}
	if c.onField() && c.Core.Player.CurrentState() == action.SkillState {
		return
	}
	if c.power <= 0 {
		return
	}
	loss := decayOnField
	if !c.onField() {
		loss = decayOffField
	}
	c.addPower(-loss)
}

func (c *char) cryoInfo(abil string, tag attacks.AttackTag, icd attacks.ICDTag, group attacks.ICDGroup, mult float64, dur float64) info.AttackInfo {
	return info.AttackInfo{
		ActorIndex:     c.Index(),
		Abil:           abil,
		AttackTag:      tag,
		ICDTag:         icd,
		ICDGroup:       group,
		StrikeType:     attacks.StrikeTypeBlunt,
		Element:        attributes.Cryo,
		Durability:     info.Durability(dur),
		Mult:           mult,
		IgnoreInfusion: true,
	}
}

func (c *char) stellarInfo(abil string, tag attacks.AttackTag, mult float64) info.AttackInfo {
	return info.AttackInfo{
		ActorIndex:       c.Index(),
		Abil:             abil,
		AttackTag:        tag,
		ICDTag:           attacks.ICDTagNone,
		ICDGroup:         attacks.ICDGroupDefault,
		StrikeType:       attacks.StrikeTypeBlunt,
		Element:          attributes.Cryo,
		Durability:       0,
		Mult:             mult,
		IgnoreDefPercent: 1,
		IgnoreInfusion:   true,
	}
}

func (c *char) queue(ai info.AttackInfo, radius float64, delay int, particle bool) {
	ap := combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, radius)
	if particle {
		c.Core.QueueAttack(ai, ap, 0, delay, c.particleCB)
		return
	}
	c.Core.QueueAttack(ai, ap, 0, delay)
}
