package wriothesley

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

const (
	stellarConductText = " (Stellar-Conduct)"
	a1StellarKey       = "wriothesley-a1-stellar"
	a1StellarHealICD   = "wriothesley-a1-stellar-icd"

	// Live talent text is 20% (Honey Hunter, KQM TCL). gcsim fcc117b still uses 60%.
	c6StellarIcicleRatio = 0.2
)

func (c *char) inConductRadiance() bool {
	return c.StatusIsActive(reactable.PolestarFieldKey)
}

func (c *char) revelationConduct() bool {
	return c.revelation && c.inConductRadiance()
}

func (c *char) stellarInit() {
	if !c.revelation {
		return
	}
	c.AddReactBonusMod(character.ReactBonusMod{
		Base: modifier.NewBase(a1StellarKey, -1),
		Amount: func(ai info.AttackInfo) float64 {
			if ai.AttackTag != attacks.AttackTagDirectStellarConduct {
				return 0
			}
			return 0.3
		},
	})
}

func (c *char) lusterReady() bool {
	if c.Base.Ascension < 1 || !c.revelationConduct() {
		return false
	}
	return c.CurrentHPRatio() < 0.6
}

func (c *char) lusterAttack() action.Info {
	ai := info.AttackInfo{
		ActorIndex:       c.Index(),
		Abil:             "Luster: Vaulting Fist" + stellarConductText,
		AttackTag:        attacks.AttackTagDirectStellarConduct,
		AdditionalTags:   []attacks.AttackTag{attacks.AttackTagWrioSSCCA},
		StrikeType:       attacks.StrikeTypeBlunt,
		PoiseDMG:         110,
		Element:          attributes.Cryo,
		IgnoreDefPercent: 1,
		Mult:             charge[c.TalentLvlAttack()] * (1 + c.c2ConductFraction()),
		HitlagFactor:     0.01,
		HitlagHaltFrames: 0.12 * 60,
	}
	snap := c.Snapshot(&ai)
	c.addC6Buff(&snap)
	ap := combat.NewBoxHitOnTarget(c.Core.Combat.Player(), info.Point{Y: -0.8}, 4, 5)
	c.QueueCharTask(func() {
		c.Core.QueueAttackWithSnap(ai, snap, ap, 0, c.lusterHealCB, c.c1LusterCB)
		c.queueConductIcicle(ai, snap, ap, "Luster: Vaulting Fist (C6)"+stellarConductText, nil)
	}, chargeHitmark)
	return action.Info{
		Frames:          frames.NewAbilFunc(chargeFrames),
		AnimationLength: chargeFrames[action.InvalidAction],
		CanQueueAfter:   chargeHitmark,
		State:           action.ChargeAttackState,
	}
}

func (c *char) lusterHealCB(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy || c.StatusIsActive(a1StellarHealICD) {
		return
	}
	c.AddStatus(a1StellarHealICD, 2*60, true)
	c.Core.Player.Heal(info.HealInfo{
		Caller:  c.Index(),
		Target:  c.Index(),
		Message: a1HealName,
		Src:     0.3 * c.MaxHP(),
		Bonus:   c.Stat(attributes.Heal),
	})
}

func (c *char) queueConductIcicle(ai info.AttackInfo, snap info.Snapshot, ap info.AttackPattern, abil string, cb info.AttackCBFunc) {
	if c.Base.Cons < 6 || !c.revelationConduct() {
		return
	}
	ai.Abil = abil
	ai.Mult *= c6StellarIcicleRatio
	ai.AdditionalTags = nil
	ai.IsDeployable = true
	c.Core.QueueAttackWithSnap(ai, snap, ap, 0, cb)
}
