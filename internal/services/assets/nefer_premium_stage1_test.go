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

// The counts below are the current engine's execution of the Stage 1 configs,
// including placeholder Nefer and Lauma frames. They are not a KQM timing
// validation and they are not a damage target.
func TestNeferPremiumStage1ConfigsExecute(t *testing.T) {
	root := repoRoot(t)
	cases := []struct {
		file      string
		actions   map[int]map[string]float64
		hits      map[int]map[string]float64
		reactions map[int]map[string]float64
	}{
		{
			file: "nefer-premium-deepwood-first.txt",
			actions: map[int]map[string]float64{
				0: {"skill": 3, "charge": 6, "dash": 6},
				1: {"skill": 1, "burst": 1},
				2: {"attack": 1, "skill": 1, "burst": 1},
				3: {"skill": 2, "dash": 2},
			},
			hits: map[int]map[string]float64{
				0: {"Charged Attack": 1, "Phantasm Performance (Nefer 1)": 5, "Phantasm Performance (Nefer 2)": 5},
				1: {"Gravity Interference (Lunar-Bloom)": 15},
				2: {"Hymn of Eternal Rest (Hold)": 1, "Hymn of Eternal Rest (Hold) (Lunar-Bloom)": 1},
			},
		},
		{
			file: "nefer-premium-deepwood-sustain.txt",
			actions: map[int]map[string]float64{
				0: {"skill": 5, "charge": 12, "dash": 12},
				1: {"skill": 2, "burst": 2},
				2: {"attack": 1, "skill": 2, "burst": 2},
				3: {"skill": 4, "dash": 4},
			},
			hits: map[int]map[string]float64{
				0: {"Charged Attack": 1, "Phantasm Performance (Nefer 1)": 11, "Phantasm Performance (Nefer 2)": 11},
				2: {"Hymn of Eternal Rest (Hold)": 2, "Hymn of Eternal Rest (Hold) (Lunar-Bloom)": 2},
			},
		},
		{
			file: "nefer-premium-instructor-first.txt",
			actions: map[int]map[string]float64{
				0: {"skill": 3, "charge": 6, "dash": 6},
				1: {"attack": 2, "skill": 1, "burst": 1},
				2: {"attack": 1, "skill": 1, "burst": 1},
				3: {"skill": 2, "dash": 2},
			},
			hits: map[int]map[string]float64{
				0: {"Charged Attack": 0, "Phantasm Performance (Nefer 1)": 6, "Phantasm Performance (Nefer 2)": 6},
				1: {"Normal 0": 2, "Gravity Interference (Lunar-Bloom)": 15},
				2: {"Hymn of Eternal Rest (Hold)": 1, "Hymn of Eternal Rest (Hold) (Lunar-Bloom)": 1},
			},
			reactions: map[int]map[string]float64{
				3: {"swirl-hydro": 0},
			},
		},
		{
			file: "nefer-premium-instructor-sustain.txt",
			actions: map[int]map[string]float64{
				0: {"skill": 5, "charge": 12, "dash": 12},
				1: {"attack": 4, "skill": 2, "burst": 2},
				2: {"attack": 1, "skill": 2, "burst": 2},
				3: {"skill": 4, "dash": 4},
			},
			hits: map[int]map[string]float64{
				0: {"Charged Attack": 0, "Phantasm Performance (Nefer 1)": 12, "Phantasm Performance (Nefer 2)": 12},
				2: {"Hymn of Eternal Rest (Hold)": 2, "Hymn of Eternal Rest (Hold) (Lunar-Bloom)": 2},
			},
			reactions: map[int]map[string]float64{
				3: {"swirl-hydro": 1},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			cfg, err := os.ReadFile(filepath.Join(root, "docs", "yuheng", "configs", tc.file))
			if err != nil {
				t.Fatal(err)
			}
			file := ast.NewFile()
			simcfg, gcsl, err := simulator.Parse(file, string(cfg))
			if err != nil {
				t.Fatal(err)
			}
			if simcfg.Settings.Iterations != 3 {
				t.Fatalf("iterations = %d, want 3", simcfg.Settings.Iterations)
			}
			res, err := simulator.RunWithSeededConfig(context.Background(), file, string(cfg), simcfg, gcsl, []int64{1, 2, 3})
			if err != nil {
				t.Fatal(err)
			}
			for char, want := range tc.actions {
				for name, n := range want {
					if got := actionMean(res, char, name); got != n {
						t.Errorf("char %d action %s = %v, want %v", char, name, got, n)
					}
				}
			}
			for char, want := range tc.hits {
				for name, n := range want {
					if got := sourceMean(res.GetStatistics().GetSourceDamageInstances(), char, name); got != n {
						t.Errorf("char %d hit %s = %v, want %v", char, name, got, n)
					}
				}
			}
			for char, want := range tc.reactions {
				for name, n := range want {
					if got := sourceMean(res.GetStatistics().GetSourceReactions(), char, name); got != n {
						t.Errorf("char %d reaction %s = %v, want %v", char, name, got, n)
					}
				}
			}
		})
	}
}

func actionMean(res *model.SimulationResult, char int, name string) float64 {
	return sourceMean(res.GetStatistics().GetCharacterActions(), char, name)
}

func sourceMean(rows []*model.SourceStats, char int, name string) float64 {
	if char >= len(rows) || rows[char] == nil {
		return 0
	}
	stat := rows[char].GetSources()[name]
	if stat == nil || stat.Mean == nil {
		return 0
	}
	return stat.GetMean()
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
