package keys

import (
	"slices"
	"testing"
)

// productionCharacterNames is the yuheng-production _CharNames list, with
// nefer, vesna, and sandrone inserted immediately before invalidchar.
var productionCharacterNames = []string{
	"", "aetheranemo", "aethercryo", "aetherdendro",
	"aetherelectro", "aethergeo", "aetherhydro", "aetherpyro",
	"aino", "albedo", "alhaitham", "aloy",
	"alyosha", "amber", "aratakiitto", "arlecchino",
	"baizhu", "barbara", "beidou", "bennett",
	"candace", "charlotte", "chasca", "chevreuse",
	"chiori", "chongyun", "citlali", "clorinde",
	"collei", "columbina", "cyno", "dahlia",
	"dehya", "diluc", "diona", "dori",
	"durin", "emilie", "escoffier", "eula",
	"faruzan", "fischl", "flins", "freminet",
	"furina", "gaming", "ganyu", "gorou",
	"hutao", "iansan", "ifa", "illuga",
	"ineffa", "jahoda", "jean", "kachina",
	"kaedeharakazuha", "kaeya", "kamisatoayaka", "kamisatoayato",
	"kaveh", "keqing", "kinich", "kirara",
	"klee", "kujousara", "kukishinobu", "lanyan",
	"lauma", "layla", "linnea", "lisa",
	"lohen", "lumineanemo", "luminecryo", "luminedendro",
	"lumineelectro", "luminegeo", "luminehydro", "luminepyro",
	"lynette", "lyney", "mavuika", "mika",
	"mona", "mualani", "nahida", "navia",
	"neuvillette", "nicole", "nilou", "ningguang",
	"noelle", "odette", "ororon", "prune",
	"qiqi", "raidenshogun", "razor", "rosaria",
	"sangonomiyakokomi", "sayu", "sethos", "shenhe",
	"shikanoinheizou", "sigewinne", "skirk", "sucrose",
	"tartaglia", "thoma", "tighnari", "varesa",
	"varka", "venti", "vodyanitsa", "wanderer",
	"wriothesley", "xiangling", "xianyun", "xiao",
	"xilonen", "xingqiu", "xinyan", "yaemiko",
	"yanfei", "yaoyao", "yelan", "yoimiya",
	"yumemizukimizuki", "yunjin", "zhongli", "nefer",
	"vesna", "sandrone", "invalidchar",
}

func TestCharacterIDsMatchProductionPin(t *testing.T) {
	got := CharStrings()
	if !slices.Equal(got, productionCharacterNames) {
		t.Fatalf("character key IDs shifted: len got %d pin %d", len(got), len(productionCharacterNames))
	}
	if Zhongli != 130 || Neuvillette != 88 || Vodyanitsa != 114 || Nilou != 90 {
		t.Fatalf("landmark IDs zhongli=%d neuvillette=%d vodyanitsa=%d nilou=%d", Zhongli, Neuvillette, Vodyanitsa, Nilou)
	}
	if Nefer != Zhongli+1 || Vesna != Zhongli+2 || Sandrone != Zhongli+3 || InvalidChar != Zhongli+4 {
		t.Fatalf("batch4 tail nefer=%d vesna=%d sandrone=%d invalid=%d", Nefer, Vesna, Sandrone, InvalidChar)
	}
}
