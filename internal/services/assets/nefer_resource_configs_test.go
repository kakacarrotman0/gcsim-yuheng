package assets

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/genshinsim/gcsim/pkg/gcs/ast"
	"github.com/genshinsim/gcsim/pkg/model"
	_ "github.com/genshinsim/gcsim/pkg/simulation"
	"github.com/genshinsim/gcsim/pkg/simulator"
)

// The hit counts are the documented natural-resource and seeded-Dew checks.
// They are re-run after the Shadow Dance swap fix. A change here is a
// behavior change, not a retarget.
//
// nefer-lunar.txt records a Lauma swap-CD mean of 0.4. That count is the
// same with and without the swap fix: Xingqiu's skill and burst do not
// always cover the swap cooldown before Lauma acts. It is not a failed
// Nefer action.
func TestNeferResourceConfigsExecute(t *testing.T) {
	root := repoRoot(t)
	cases := []struct {
		file      string
		hits      map[string]float64
		laumaSwap float64
	}{
		{
			file: "nefer-natural.txt",
			hits: map[string]float64{
				"Charged Attack":                 0,
				"Phantasm Performance (Nefer 1)": 1,
				"Phantasm Performance (Nefer 2)": 1,
				"Phantasm Performance (Shade 1)": 1,
				"Phantasm Performance (Shade 2)": 1,
				"Phantasm Performance (Shade 3)": 1,
			},
		},
		{
			file: "nefer-lunar.txt",
			hits: map[string]float64{
				"Charged Attack":                 0,
				"Phantasm Performance (Nefer 1)": 3,
				"Phantasm Performance (Nefer 2)": 3,
				"Phantasm Performance (Shade 1)": 3,
				"Phantasm Performance (Shade 2)": 3,
				"Phantasm Performance (Shade 3)": 3,
			},
			laumaSwap: 0.4,
		},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(root, "docs", "yuheng", "configs", tc.file))
			if err != nil {
				t.Fatal(err)
			}
			cfg := string(raw)
			file := ast.NewFile()
			simcfg, gcsl, err := simulator.Parse(file, cfg)
			if err != nil {
				t.Fatal(err)
			}
			seeds := make([]int64, simcfg.Settings.Iterations)
			for i := range seeds {
				seeds[i] = int64(i + 1)
			}
			res, err := simulator.RunWithSeededConfig(context.Background(), file, cfg, simcfg, gcsl, seeds)
			if err != nil {
				t.Fatal(err)
			}
			for i, row := range res.GetStatistics().GetFailedActions() {
				if row == nil {
					continue
				}
				wantSwap := 0.0
				if i == 1 {
					wantSwap = tc.laumaSwap
				}
				if got := descMean(row.GetSwapCd()); got != wantSwap {
					t.Errorf("char %d swap cd %v, want %v", i, got, wantSwap)
				}
				for _, got := range []float64{
					descMean(row.GetInsufficientEnergy()),
					descMean(row.GetInsufficientStamina()),
					descMean(row.GetSkillCd()),
					descMean(row.GetDashCd()),
					descMean(row.GetBurstCd()),
				} {
					if got != 0 {
						t.Errorf("char %d unexpected failed-action mean %v", i, got)
					}
				}
			}
			for name, want := range tc.hits {
				if got := sourceMean(res.GetStatistics().GetSourceDamageInstances(), 0, name); got != want {
					t.Errorf("hit %s = %v, want %v", name, got, want)
				}
			}
		})
	}
}

func descMean(stat *model.DescriptiveStats) float64 {
	if stat == nil || stat.Mean == nil {
		return 0
	}
	return stat.GetMean()
}
