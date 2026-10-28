package main

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/shizukayuki/excel-hk4e"
)

// pinnedCharacterSlugs is every character const except NoChar and InvalidChar.
// The production order is unchanged. Nefer, Vesna, Sandrone, and Zibai are
// appended so their iota values sit after Zhongli. Zibai is not inserted
// alphabetically ahead of Zhongli.
var pinnedCharacterSlugs = []string{
	"AetherAnemo", "AetherCryo", "AetherDendro", "AetherElectro",
	"AetherGeo", "AetherHydro", "AetherPyro", "Aino",
	"Albedo", "Alhaitham", "Aloy", "Alyosha",
	"Amber", "AratakiItto", "Arlecchino", "Baizhu",
	"Barbara", "Beidou", "Bennett", "Candace",
	"Charlotte", "Chasca", "Chevreuse", "Chiori",
	"Chongyun", "Citlali", "Clorinde", "Collei",
	"Columbina", "Cyno", "Dahlia", "Dehya",
	"Diluc", "Diona", "Dori", "Durin",
	"Emilie", "Escoffier", "Eula", "Faruzan",
	"Fischl", "Flins", "Freminet", "Furina",
	"Gaming", "Ganyu", "Gorou", "HuTao",
	"Iansan", "Ifa", "Illuga", "Ineffa",
	"Jahoda", "Jean", "Kachina", "KaedeharaKazuha",
	"Kaeya", "KamisatoAyaka", "KamisatoAyato", "Kaveh",
	"Keqing", "Kinich", "Kirara", "Klee",
	"KujouSara", "KukiShinobu", "LanYan", "Lauma",
	"Layla", "Linnea", "Lisa", "Lohen",
	"LumineAnemo", "LumineCryo", "LumineDendro", "LumineElectro",
	"LumineGeo", "LumineHydro", "LuminePyro", "Lynette",
	"Lyney", "Mavuika", "Mika", "Mona",
	"Mualani", "Nahida", "Navia", "Neuvillette",
	"Nicole", "Nilou", "Ningguang", "Noelle",
	"Odette", "Ororon", "Prune", "Qiqi",
	"RaidenShogun", "Razor", "Rosaria", "SangonomiyaKokomi",
	"Sayu", "Sethos", "Shenhe", "ShikanoinHeizou",
	"Sigewinne", "Skirk", "Sucrose", "Tartaglia",
	"Thoma", "Tighnari", "Varesa", "Varka",
	"Venti", "Vodyanitsa", "Wanderer", "Wriothesley",
	"Xiangling", "Xianyun", "Xiao", "Xilonen",
	"Xingqiu", "Xinyan", "YaeMiko", "Yanfei",
	"Yaoyao", "Yelan", "Yoimiya", "YumemizukiMizuki",
	"YunJin", "Zhongli", "Nefer", "Vesna",
	"Sandrone", "Zibai",
}

// pinnedICDGroups is the ICD const suffix. Groups through LohenSkill keep the
// production indexes. The hand-maintained tail is Alyosha, Lohen, Vesna's
// entry group, Vesna's 2s group, then Sandrone's provisional machine-gun group.
var pinnedICDGroups = []string{
	"Default", "PoleExtraAttack", "ReactionA", "ReactionB",
	"Burning", "AinoBurstMoonHit", "AlhaithamExtraAttack", "AlhaithamProjectionAttack",
	"Amber", "ArlecchinoElementalArt", "AyakaExtraAttack", "BaizhuC2",
	"CharlotteKamera", "CharlotteMark", "ChascaBurst", "ChascaShadowhunt",
	"ChascaShining", "ChascaTap", "ChevreuseBurstMines", "ChioriSkill",
	"CitlaliFrostfallStorm", "ClorindeElementalArt", "ColleiBurst", "CynoBolt",
	"Diluc", "DoriBurst", "DurinBurstBlack", "DurinBurstWhite",
	"DurinSkill", "EmilieLumidouce", "EscoffierSkill", "Fischl",
	"FurinaSalonSolitaire", "IfaSedationMark", "IfaSkillHit", "JahodaBurst",
	"JahodaCons", "KinichLoopShot", "KinichScalespikerCannon", "LanyanRingAttack",
	"LanyanRingAttackMix", "Layla", "LyneyExtra", "MizukiSkill",
	"NahidaSkill", "NaviaBurst", "Nilou", "OdetteDanceDuo",
	"OroronElementalBurst", "Sigewinne", "SigewinneBurst", "Tighnari",
	"TravelerBurst", "TravelerCryoIcicle", "TravelerDewdrop", "Venti",
	"WandererA4", "WandererC6", "XiaoDash", "YaeCharged",
	"YaoyaoRadishBurst", "YaoyaoRadishSkill", "YelanBreakthrough", "YelanBurst",
	"AlyoshaElementalBurst", "LohenSkill", "VesnaSkillStart", "VesnaSkill",
	"SandroneMachineGun",
}

type icdSeqPin struct {
	timer float64
	dur   int
	dmg   int
}

// Fully specified rows. A generator that shortens one of these must refuse to write.
var pinnedHandICDSeq = map[string]icdSeqPin{
	"AlyoshaElementalBurst": {timer: 1.6, dur: 2, dmg: 4},
	"LohenSkill":            {timer: 5, dur: 12, dmg: 12},
	"VesnaSkillStart":       {timer: 2.5, dur: 24, dmg: 15},
	"VesnaSkill":            {timer: 2, dur: 2, dmg: 1},
	"SandroneMachineGun":    {timer: 2.5, dur: 24, dmg: 15},
}

func validateCharacterKeyNames(specNames []string) error {
	slugs := make([]string, len(specNames))
	for i, name := range specNames {
		slugs[i] = excel.Slug(name)
	}
	if slices.Equal(slugs, pinnedCharacterSlugs) {
		return nil
	}
	return fmt.Errorf("character keys would drop a character or change an existing ID (got %d, pin %d)", len(slugs), len(pinnedCharacterSlugs))
}

func validateICDOrder(sorted []string) error {
	if slices.Equal(sorted, pinnedICDGroups) {
		return nil
	}
	return fmt.Errorf("ICD groups would be missing, reordered, or inserted ahead of a production group (got %d, pin %d)", len(sorted), len(pinnedICDGroups))
}

func validateHandICDSequences(groups map[string]AttackAttenuation) error {
	for name, want := range pinnedHandICDSeq {
		att, ok := groups[name]
		if !ok {
			return fmt.Errorf("missing hand-maintained ICD group %s", name)
		}
		if att.Timer != want.timer || len(att.Durability) != want.dur || len(att.Damage) != want.dmg {
			return fmt.Errorf("ICD group %s sequence changed: timer=%v durability=%d damage=%d", name, att.Timer, len(att.Durability), len(att.Damage))
		}
	}
	return nil
}

// fullRegenGate keeps unattended regeneration off until released character
// data and a stable generated-table strategy exist. It does not replace the
// character or ICD guards below.
func fullRegenGate() error {
	if os.Getenv("GCSIM_ALLOW_FULL_REGEN") == "1" {
		return nil
	}
	return errors.New("refusing full pipeline regeneration: set GCSIM_ALLOW_FULL_REGEN=1 only when released character data and a stable generated-table strategy exist; checked-in generated files were not modified")
}

// assertHandMaintainedRegen refuses a pipeline run whose character or ICD
// output would renumber production IDs or drop Alyosha, Lohen, Vesna, Sandrone, or Zibai.
// Nefer, Vesna, Sandrone, and Zibai are not excel-backed in this module, so a full run
// fails closed before any generated file is written. Zibai's config.yml is hand-maintained;
// a generator that sorted it ahead of Zhongli would still miss this tail pin.
func (c *Compiled) assertHandMaintainedRegen() error {
	var names []string
	for _, cfg := range c.Configuration {
		if cfg.Kind == KindCharacter && cfg.Character != nil {
			names = append(names, cfg.Character.Name)
		}
	}
	if err := validateCharacterKeyNames(names); err != nil {
		return fmt.Errorf("refusing to regenerate: %w", err)
	}
	sorted := slices.SortedFunc(maps.Keys(c.ICDGroup), compareICDGroup)
	if err := validateICDOrder(sorted); err != nil {
		return fmt.Errorf("refusing to regenerate: %w", err)
	}
	if err := validateHandICDSequences(c.ICDGroup); err != nil {
		return fmt.Errorf("refusing to regenerate: %w", err)
	}
	return nil
}
