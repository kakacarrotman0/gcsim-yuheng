package assets

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/genshinsim/gcsim/pkg/catalog"
	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/gcs/ast"
	"github.com/genshinsim/gcsim/pkg/model"
	_ "github.com/genshinsim/gcsim/pkg/simulation"
	"github.com/genshinsim/gcsim/pkg/simulator"
)

func TestEveryCharacterNameIsRegisteredOnce(t *testing.T) {
	seen := map[keys.Char]string{}
	for _, name := range keys.CharStrings() {
		if name == "" || name == "invalidchar" {
			continue
		}
		key, err := keys.CharString(name)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		if prev, ok := seen[key]; ok {
			t.Fatalf("duplicate key %d for %s and %s", key, prev, name)
		}
		seen[key] = name
		if _, ok := core.NewCharFuncMap[key]; !ok {
			t.Fatalf("%s is not registered", name)
		}
		if _, ok := catalog.CharacterMap[key]; !ok {
			t.Fatalf("%s has no catalog entry", name)
		}
	}
	if len(core.NewCharFuncMap) != len(seen) {
		t.Fatalf("registered %d names %d", len(core.NewCharFuncMap), len(seen))
	}
}

func TestBatch4CatalogAndAssets(t *testing.T) {
	checks := []struct {
		key     keys.Char
		name    string
		element model.ElementType
		weapon  model.WeaponType
		icon    string
		attack  uint32
		skill   uint32
		burst   uint32
	}{
		{keys.Nefer, "nefer", model.ElementType_Grass, model.WeaponType_WEAPON_CATALYST, "UI_AvatarIcon_Nefer", 11221, 11222, 11225},
		{keys.Vesna, "vesna", model.ElementType_Wind, model.WeaponType_WEAPON_SWORD_ONE_HAND, "UI_AvatarIcon_Vesna", 11431, 11432, 11435},
		{keys.Sandrone, "sandrone", model.ElementType_Ice, model.WeaponType_WEAPON_CLAYMORE, "UI_AvatarIcon_MarionetteNew", 11331, 11332, 11335},
		{keys.Zibai, "zibai", model.ElementType_Rock, model.WeaponType_WEAPON_SWORD_ONE_HAND, "UI_AvatarIcon_Zibai", 11261, 11262, 11265},
	}
	for _, want := range checks {
		data := catalog.CharacterMap[want.key]
		if data.Key != want.name || data.Rarity != model.QualityType_QUALITY_ORANGE || data.Element != want.element || data.WeaponClass != want.weapon {
			t.Fatalf("%s catalog %+v", want.name, data)
		}
		if data.IconName != want.icon || avatarMap[want.name] != want.icon {
			t.Fatalf("%s icon catalog %s map %s", want.name, data.IconName, avatarMap[want.name])
		}
		if data.SkillDetails.Attack != want.attack || data.SkillDetails.Skill != want.skill || data.SkillDetails.Burst != want.burst || data.SkillDetails.BurstEnergyCost != 60 {
			t.Fatalf("%s talents %+v", want.name, data.SkillDetails)
		}
	}
}

func TestBatch4ConfigParsesAndRuns(t *testing.T) {
	cfg := `options iteration=1 duration=10;
target lvl=100 resist=0.1;
nefer char lvl=90/90 cons=0 talent=1,1,1;
nefer add weapon="sacrificialfragments" refine=1 lvl=90/90;
vesna char lvl=90/90 cons=0 talent=1,1,1;
vesna add weapon="favoniussword" refine=1 lvl=90/90;
sandrone char lvl=90/90 cons=6 talent=1,1,1;
sandrone add weapon="tidalshadow" refine=1 lvl=90/90;
zhongli char lvl=90/90 cons=0 talent=1,1,1;
zhongli add weapon="blacktassel" refine=1 lvl=90/90;
active nefer;
nefer attack;
vesna skill;
sandrone skill;
`
	file := ast.NewFile()
	simcfg, gcsl, err := simulator.Parse(file, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := simulator.RunWithSeededConfig(context.Background(), file, cfg, simcfg, gcsl, []int64{1}); err != nil {
		t.Fatal(err)
	}
}

func TestZibaiConfigParsesAndRuns(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), "docs/yuheng/configs/zibai-c0.txt"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := string(raw)
	file := ast.NewFile()
	simcfg, gcsl, err := simulator.Parse(file, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := simulator.RunWithSeededConfig(context.Background(), file, cfg, simcfg, gcsl, []int64{1}); err != nil {
		t.Fatal(err)
	}
}
