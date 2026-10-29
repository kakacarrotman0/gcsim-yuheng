package assets

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/genshinsim/gcsim/pkg/gcs/ast"
	_ "github.com/genshinsim/gcsim/pkg/simulation"
	"github.com/genshinsim/gcsim/pkg/simulator"
)

func TestZibaiStrideFromTicker(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), "docs/yuheng/configs/zibai-stride.txt"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := string(raw)
	if strings.Contains(cfg, "set_starting") || strings.Contains(cfg, "radiance=") {
		t.Fatal("stride config seeds a resource")
	}
	file := ast.NewFile()
	simcfg, gcsl, err := simulator.Parse(file, cfg)
	if err != nil {
		t.Fatal(err)
	}
	res, err := simulator.RunWithSeededConfig(context.Background(), file, cfg, simcfg, gcsl, []int64{1})
	if err != nil {
		t.Fatal(err)
	}
	if got := actionMean(res, 0, "skill"); got != 2 {
		t.Fatalf("skill actions %v", got)
	}
	stats := res.GetStatistics()
	if got := sourceMean(stats.GetSourceDamageInstances(), 0, "Spirit Steed's Stride"); got < 1 {
		t.Fatalf("stride hit 1 %v", got)
	}
	if got := sourceMean(stats.GetSourceDamageInstances(), 0, "Spirit Steed's Stride (Lunar-Crystallize)"); got < 1 {
		t.Fatalf("direct lunar-crystallize hit %v", got)
	}

	sample, err := simulator.GenerateSampleWithSeed(cfg, 1, simulator.Options{})
	if err != nil {
		t.Fatal(err)
	}
	spent := false
	for _, ev := range sample.GetLogs() {
		msg := ev.GetFields()["msg"].GetStringValue()
		if strings.Contains(msg, "Consumed 70 radiance") {
			spent = true
			break
		}
	}
	if !spent {
		t.Fatal("debug log did not record a 70 radiance spend")
	}
}
