package attacks

import (
	"slices"
	"testing"
)

type icdPinRow struct {
	timer int
	dur   int
	dmg   int
}

// icdPin is the production timer and sequence length of every group, plus the
// three groups appended by this integration. Index is the iota value.
var icdPin = []icdPinRow{
	{timer: 150, dur: 24, dmg: 15},  // Default
	{timer: 30, dur: 24, dmg: 15},   // PoleExtraAttack
	{timer: 30, dur: 10, dmg: 10},   // ReactionA
	{timer: 30, dur: 10, dmg: 10},   // ReactionB
	{timer: 120, dur: 8, dmg: 14},   // Burning
	{timer: 108, dur: 4, dmg: 4},    // AinoBurstMoonHit
	{timer: 120, dur: 15, dmg: 15},  // AlhaithamExtraAttack
	{timer: 720, dur: 24, dmg: 15},  // AlhaithamProjectionAttack
	{timer: 60, dur: 24, dmg: 15},   // Amber
	{timer: 600, dur: 20, dmg: 5},   // ArlecchinoElementalArt
	{timer: 30, dur: 27, dmg: 27},   // AyakaExtraAttack
	{timer: 240, dur: 11, dmg: 15},  // BaizhuC2
	{timer: 240, dur: 8, dmg: 12},   // CharlotteKamera
	{timer: 720, dur: 10, dmg: 12},  // CharlotteMark
	{timer: 90, dur: 6, dmg: 6},     // ChascaBurst
	{timer: 90, dur: 2, dmg: 2},     // ChascaShadowhunt
	{timer: 90, dur: 6, dmg: 6},     // ChascaShining
	{timer: 90, dur: 2, dmg: 2},     // ChascaTap
	{timer: 180, dur: 8, dmg: 8},    // ChevreuseBurstMines
	{timer: 114, dur: 5, dmg: 5},    // ChioriSkill
	{timer: 90, dur: 2, dmg: 2},     // CitlaliFrostfallStorm
	{timer: 60, dur: 24, dmg: 13},   // ClorindeElementalArt
	{timer: 180, dur: 12, dmg: 12},  // ColleiBurst
	{timer: 150, dur: 11, dmg: 11},  // CynoBolt
	{timer: 300, dur: 30, dmg: 15},  // Diluc
	{timer: 180, dur: 12, dmg: 12},  // DoriBurst
	{timer: 120, dur: 2, dmg: 2},    // DurinBurstBlack
	{timer: 90, dur: 2, dmg: 2},     // DurinBurstWhite
	{timer: 18, dur: 3, dmg: 3},     // DurinSkill
	{timer: 120, dur: 2, dmg: 2},    // EmilieLumidouce
	{timer: 90, dur: 6, dmg: 5},     // EscoffierSkill
	{timer: 300, dur: 32, dmg: 15},  // Fischl
	{timer: 1800, dur: 24, dmg: 12}, // FurinaSalonSolitaire
	{timer: 120, dur: 2, dmg: 2},    // IfaSedationMark
	{timer: 720, dur: 18, dmg: 18},  // IfaSkillHit
	{timer: 900, dur: 20, dmg: 20},  // JahodaBurst
	{timer: 900, dur: 10, dmg: 10},  // JahodaCons
	{timer: 120, dur: 4, dmg: 8},    // KinichLoopShot
	{timer: 72, dur: 4, dmg: 4},     // KinichScalespikerCannon
	{timer: 150, dur: 24, dmg: 15},  // LanyanRingAttack
	{timer: 150, dur: 24, dmg: 15},  // LanyanRingAttackMix
	{timer: 180, dur: 12, dmg: 12},  // Layla
	{timer: 60, dur: 6, dmg: 6},     // LyneyExtra
	{timer: 72, dur: 4, dmg: 4},     // MizukiSkill
	{timer: 60, dur: 11, dmg: 21},   // NahidaSkill
	{timer: 720, dur: 24, dmg: 24},  // NaviaBurst
	{timer: 114, dur: 12, dmg: 12},  // Nilou
	{timer: 180, dur: 4, dmg: 4},    // OdetteDanceDuo
	{timer: 180, dur: 7, dmg: 12},   // OroronElementalBurst
	{timer: 120, dur: 13, dmg: 13},  // Sigewinne
	{timer: 114, dur: 13, dmg: 13},  // SigewinneBurst
	{timer: 150, dur: 12, dmg: 12},  // Tighnari
	{timer: 480, dur: 12, dmg: 12},  // TravelerBurst
	{timer: 150, dur: 4, dmg: 4},    // TravelerCryoIcicle
	{timer: 90, dur: 8, dmg: 8},     // TravelerDewdrop
	{timer: 60, dur: 24, dmg: 15},   // Venti
	{timer: 60, dur: 20, dmg: 15},   // WandererA4
	{timer: 120, dur: 20, dmg: 15},  // WandererC6
	{timer: 6, dur: 7, dmg: 7},      // XiaoDash
	{timer: 30, dur: 7, dmg: 7},     // YaeCharged
	{timer: 90, dur: 6, dmg: 6},     // YaoyaoRadishBurst
	{timer: 150, dur: 6, dmg: 6},    // YaoyaoRadishSkill
	{timer: 18, dur: 4, dmg: 4},     // YelanBreakthrough
	{timer: 120, dur: 24, dmg: 15},  // YelanBurst
	{timer: 96, dur: 2, dmg: 4},     // AlyoshaElementalBurst
	{timer: 300, dur: 12, dmg: 12},  // LohenSkill
	{timer: 150, dur: 24, dmg: 15},  // VesnaSkillStart
	{timer: 120, dur: 2, dmg: 1},    // VesnaSkill
	{timer: 150, dur: 24, dmg: 15},  // SandroneMachineGun
}

func TestICDGroupIDsAndSequenceLengths(t *testing.T) {
	if len(ICDGroupResetTimer) != len(icdPin) || len(ICDGroupEleApplicationSequence) != len(icdPin) || len(ICDGroupDamageSequence) != len(icdPin) {
		t.Fatalf("table len timer=%d ele=%d dmg=%d pin=%d", len(ICDGroupResetTimer), len(ICDGroupEleApplicationSequence), len(ICDGroupDamageSequence), len(icdPin))
	}
	for i, want := range icdPin {
		if ICDGroupResetTimer[i] != want.timer {
			t.Fatalf("group %d timer %d want %d", i, ICDGroupResetTimer[i], want.timer)
		}
		if len(ICDGroupEleApplicationSequence[i]) != want.dur || len(ICDGroupDamageSequence[i]) != want.dmg {
			t.Fatalf("group %d truncated ele=%d dmg=%d", i, len(ICDGroupEleApplicationSequence[i]), len(ICDGroupDamageSequence[i]))
		}
		if want.dur == 0 || want.dmg == 0 || want.timer == 0 {
			t.Fatalf("group %d empty sequence", i)
		}
	}
	if int(ICDGroupAlyoshaElementalBurst) != int(ICDGroupLohenSkill)-1 {
		t.Fatal("Alyosha is no longer the group immediately before Lohen")
	}
	if int(ICDGroupVesnaSkillStart) != int(ICDGroupLohenSkill)+1 ||
		int(ICDGroupVesnaSkill) != int(ICDGroupLohenSkill)+2 ||
		int(ICDGroupSandroneMachineGun) != int(ICDGroupLohenSkill)+3 {
		t.Fatalf("tail ids start=%d skill=%d gun=%d lohen=%d", ICDGroupVesnaSkillStart, ICDGroupVesnaSkill, ICDGroupSandroneMachineGun, ICDGroupLohenSkill)
	}
	if int(ICDGroupLohenSkill) != 65 {
		t.Fatalf("LohenSkill production id changed: %d", ICDGroupLohenSkill)
	}
}

func TestHandMaintainedICDSequences(t *testing.T) {
	eq := func(g ICDGroup, ele, dmg []float64) {
		t.Helper()
		if !slices.Equal(ICDGroupEleApplicationSequence[g], ele) || !slices.Equal(ICDGroupDamageSequence[g], dmg) {
			t.Fatalf("%d sequence changed", g)
		}
	}
	eq(ICDGroupAlyoshaElementalBurst, []float64{1.0, 0.0}, []float64{1.0, 1.0, 1.0, 1.0})
	eq(ICDGroupLohenSkill, []float64{1.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0, 1.0, 0.0}, []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0})
	eq(ICDGroupVesnaSkillStart, []float64{1.0, 0.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0}, []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0})
	eq(ICDGroupVesnaSkill, []float64{1.0, 0.0}, []float64{1.0})
	eq(ICDGroupSandroneMachineGun, []float64{1.0, 0.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0, 1.0, 0.0, 0.0}, []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0})

	if ICDGroupResetTimer[ICDGroupAlyoshaElementalBurst] != 96 ||
		ICDGroupResetTimer[ICDGroupLohenSkill] != 300 ||
		ICDGroupResetTimer[ICDGroupVesnaSkillStart] != 150 ||
		ICDGroupResetTimer[ICDGroupVesnaSkill] != 120 ||
		ICDGroupResetTimer[ICDGroupSandroneMachineGun] != 150 {
		t.Fatal("hand-maintained reset timer changed")
	}
}
