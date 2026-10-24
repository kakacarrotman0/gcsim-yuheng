package main

import (
	"slices"
	"testing"
)

func TestCompareICDGroupOrder(t *testing.T) {
	names := []string{
		"Venti",
		"Default",
		"VesnaSkill",
		"Burning",
		"AlyoshaElementalBurst",
		"LohenSkill",
		"VesnaSkillStart",
		"SandroneMachineGun",
		"Amber",
		"YelanBurst",
	}
	slices.SortFunc(names, compareICDGroup)
	want := []string{
		"Default",
		"Burning",
		"Amber",
		"Venti",
		"YelanBurst",
		"AlyoshaElementalBurst",
		"LohenSkill",
		"VesnaSkillStart",
		"VesnaSkill",
		"SandroneMachineGun",
	}
	if !slices.Equal(names, want) {
		t.Fatalf("order %v", names)
	}
}

func TestCompareICDGroupMatchesFullPin(t *testing.T) {
	names := append([]string(nil), pinnedICDGroups...)
	slices.Reverse(names)
	slices.SortFunc(names, compareICDGroup)
	if !slices.Equal(names, pinnedICDGroups) {
		t.Fatalf("comparator drifted from the pin")
	}
}

func TestRegenGuardsRejectIDShifts(t *testing.T) {
	if err := validateICDOrder(pinnedICDGroups); err != nil {
		t.Fatal(err)
	}
	dropped := pinnedICDGroups[:len(pinnedICDGroups)-1]
	if err := validateICDOrder(dropped); err == nil {
		t.Fatal("dropped SandroneMachineGun was accepted")
	}
	swapped := append([]string(nil), pinnedICDGroups...)
	swapped[len(swapped)-1], swapped[len(swapped)-2] = swapped[len(swapped)-2], swapped[len(swapped)-1]
	if err := validateICDOrder(swapped); err == nil {
		t.Fatal("swapped tail was accepted")
	}
	inserted := append([]string{"Default"}, pinnedICDGroups...)
	if err := validateICDOrder(inserted); err == nil {
		t.Fatal("inserted leading group was accepted")
	}
	if err := validateCharacterKeyNames(pinnedCharacterSlugs); err != nil {
		t.Fatal(err)
	}
	alpha := append([]string(nil), pinnedCharacterSlugs...)
	alpha = append(alpha[:88], append([]string{"Nefer"}, alpha[88:]...)...)
	alpha = alpha[:len(alpha)-1]
	if err := validateCharacterKeyNames(alpha); err == nil {
		t.Fatal("alphabetical Nefer insert was accepted")
	}
	shortSeq := map[string]AttackAttenuation{
		"AlyoshaElementalBurst": {Timer: 1.6, Durability: []float64{1, 0}, Damage: []float64{1, 1, 1, 1}},
		"LohenSkill":            {Timer: 5, Durability: []float64{1}, Damage: []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}},
		"VesnaSkillStart":       {Timer: 2.5, Durability: []float64{1, 0, 0}, Damage: []float64{1}},
		"VesnaSkill":            {Timer: 2, Durability: []float64{1, 0}, Damage: []float64{1}},
		"SandroneMachineGun":    {Timer: 2.5, Durability: []float64{1}, Damage: []float64{1}},
	}
	if err := validateHandICDSequences(shortSeq); err == nil {
		t.Fatal("truncated sequences were accepted")
	}
}
