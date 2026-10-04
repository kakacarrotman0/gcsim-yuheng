package vodyanitsa

import (
	"github.com/genshinsim/gcsim/internal/frames"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/attacks"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/combat"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/enemy"
	"github.com/genshinsim/gcsim/pkg/modifier"
	"github.com/genshinsim/gcsim/pkg/reactable"
)

var skillFrames []int

const (
	// APPROXIMATE cast hitmark. Horn, heal, RES, and particle timings below are from the ability graph.
	skillHitmark = 22

	skillCD          = 16 * 60
	songDur          = 16 * 60
	c2SongExtend     = 9 * 60
	hornInterval     = 3 * 60
	hornDamageDelay  = 12 // bullet modifier lasts 0.2s, then the hit is fired
	healPreDelay     = 12 // 0.2s
	healInterval     = 90 // 1.5s
	particleICD      = 12 // DropBall unique 0.2s
	particleCount    = 3
	resDur           = 6 * 60
	concertoDur      = 30 * 60
	leadMax          = 25
	chorusMax        = 10
	a1AnemoShred     = 0.35
	a1AnemoDur       = 6 * 60
	a1AlterDur       = 5 * 60
	a4MinHP          = 40000.0
	a4HydroPerKilo   = 140.0
	a4HydroCap       = 3500.0
	a4StellarPerKilo = 260.0
	a4StellarCap     = 6500.0
	skillRadius      = 5.5
	hornRadius       = 3.0
	burstEnergy      = 60
	burstCD          = 15 * 60

	songKey       = "vodyanitsa-song"
	particleKey   = "vodyanitsa-particle-icd"
	concertoKey   = "vodyanitsa-concerto"
	hydroShredKey = "vodyanitsa-hydro-res"
	cryoShredKey  = "vodyanitsa-cryo-res"
	anemoShredKey = "vodyanitsa-anemo-res"
)

func init() {
	skillFrames = frames.InitAbilSlice(48)
	skillFrames[action.ActionAttack] = 36
	skillFrames[action.ActionBurst] = 38
	skillFrames[action.ActionDash] = 32
	skillFrames[action.ActionJump] = 32
	skillFrames[action.ActionSwap] = 34
}

func (c *char) Skill(p map[string]int) (action.Info, error) {
	c.skillSrc = c.Core.F
	src := c.skillSrc
	dur := c.songDuration()

	c.AddStatus(songKey, dur, false)
	c.Core.Flags.Custom[reactable.VodyanitsaSongUntilKey] = float64(c.Core.F + dur)

	ai := info.AttackInfo{
		ActorIndex: c.Index(),
		Abil:       "Rechitativ: Sonorous Dawn",
		AttackTag:  attacks.AttackTagElementalArt,
		ICDTag:     attacks.ICDTagElementalArt,
		ICDGroup:   attacks.ICDGroupDefault,
		StrikeType: attacks.StrikeTypeDefault,
		Element:    attributes.Hydro,
		Durability: 25,
		PoiseDMG:   50,
	}
	c.QueueCharTask(func() {
		ai.FlatDmg = skill[c.TalentLvlSkill()] * c.MaxHP()
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(c.Core.Combat.Player(), nil, skillRadius),
			0,
			0,
			c.resCB,
		)
	}, skillHitmark)

	// Horns spawn on the fixed interval and deal damage 0.2s later. First spawn is after one interval.
	for t := hornInterval; t < dur; t += hornInterval {
		c.Core.Tasks.Add(c.hornAttack(src), t+hornDamageDelay)
	}
	// Heal handler is attached 0.2s after the song starts, then ticks every 1.5s with no immediate trigger.
	for t := healPreDelay + healInterval; t < dur; t += healInterval {
		c.Core.Tasks.Add(c.songHeal(src), t)
	}

	c.onSkillCast()
	c.SetCD(action.ActionSkill, skillCD)

	return action.Info{
		Frames:          frames.NewAbilFunc(skillFrames),
		AnimationLength: skillFrames[action.InvalidAction],
		CanQueueAfter:   skillFrames[action.ActionDash],
		State:           action.SkillState,
	}, nil
}

func (c *char) hornAttack(src int) func() {
	return func() {
		if c.skillSrc != src || !c.songActive() {
			return
		}
		ai := info.AttackInfo{
			ActorIndex: c.Index(),
			Abil:       "Horn of Spring's Call",
			AttackTag:  attacks.AttackTagElementalArt,
			ICDTag:     attacks.ICDTagElementalArt,
			ICDGroup:   attacks.ICDGroupDefault,
			StrikeType: attacks.StrikeTypeDefault,
			Element:    attributes.Hydro,
			Durability: 25,
			FlatDmg:    horn[c.TalentLvlSkill()] * c.MaxHP(),
		}
		center := c.Core.Combat.PrimaryTarget()
		if center == nil {
			center = c.Core.Combat.Player()
		}
		c.Core.QueueAttack(
			ai,
			combat.NewCircleHitOnTarget(center, nil, hornRadius),
			0,
			0,
			c.resCB,
			c.particleCB,
			c.c2OnHorn,
		)
	}
}

func (c *char) resCB(a info.AttackCB) {
	e, ok := a.Target.(*enemy.Enemy)
	if !ok {
		return
	}
	shred := resShred[c.TalentLvlSkill()]
	e.AddResistMod(info.ResistMod{
		Base:  modifier.NewBase(hydroShredKey, resDur),
		Ele:   attributes.Hydro,
		Value: -shred,
	})
	e.AddResistMod(info.ResistMod{
		Base:  modifier.NewBase(cryoShredKey, resDur),
		Ele:   attributes.Cryo,
		Value: -shred,
	})
}

func (c *char) particleCB(a info.AttackCB) {
	if a.Target.Type() != info.TargettableEnemy {
		return
	}
	if c.StatusIsActive(particleKey) {
		return
	}
	c.AddStatus(particleKey, particleICD, false)
	c.Core.QueueParticle(c.Base.Key.String(), particleCount, attributes.Hydro, c.ParticleDelay)
}

func (c *char) songHeal(src int) func() {
	return func() {
		if c.skillSrc != src || !c.songActive() {
			return
		}
		target := c.Core.Player.Active()
		active := c.Core.Player.ByIndex(target)
		// The heal reads Max HP before the C4 stack is applied. The ability heals, then applies the Max HP modifier.
		flat := healFlat[c.TalentLvlSkill()]
		pct := healPct[c.TalentLvlSkill()]
		hp := c.MaxHP()
		low := c.Base.Cons >= 4 && active.CurrentHPRatio() < c4HPThreshold
		if low {
			flat *= 1 + c4ExtraHeal
			pct *= 1 + c4ExtraHeal
		}
		c.Core.Player.Heal(info.HealInfo{
			Caller:  c.Index(),
			Target:  target,
			Message: "Song of Ages Past",
			Src:     flat + pct*hp,
			Bonus:   c.Stat(attributes.Heal),
		})
		if c.Base.Cons >= 4 && !low {
			c.addC4Stack()
		}
		c.c1OnHeal()
	}
}
