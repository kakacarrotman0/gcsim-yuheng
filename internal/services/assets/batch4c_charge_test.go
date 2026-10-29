package assets

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/genshinsim/gcsim/pkg/gcs/ast"
	_ "github.com/genshinsim/gcsim/pkg/simulation"
	"github.com/genshinsim/gcsim/pkg/simulator"
)

func TestClaymoreChargeConfigsBatch4C(t *testing.T) {
	for _, name := range []string{"dori-charge.txt", "sayu-charge.txt", "eula-charge.txt", "kaveh-charge.txt"} {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(repoRoot(t), "docs/yuheng/configs", name))
			if err != nil {
				t.Fatal(err)
			}
			cfg := string(raw)
			file := ast.NewFile()
			simcfg, gcsl, err := simulator.Parse(file, cfg)
			if err != nil {
				t.Fatal(err)
			}
			res, err := simulator.RunWithSeededConfig(context.Background(), file, cfg, simcfg, gcsl, []int64{1})
			if err != nil {
				t.Fatal(err)
			}
			if sourceMean(res.GetStatistics().GetSourceDamageInstances(), 0, "Charge Attack") <= 0 {
				t.Fatal("no cyclic charged-attack hit")
			}
			if sourceMean(res.GetStatistics().GetSourceDamageInstances(), 0, "Charge Attack (Finisher)") <= 0 {
				t.Fatal("no charged-attack finisher")
			}
		})
	}
}
