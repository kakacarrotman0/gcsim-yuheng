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
	}
	if !slices.Equal(names, want) {
		t.Fatalf("order %v", names)
	}
}
