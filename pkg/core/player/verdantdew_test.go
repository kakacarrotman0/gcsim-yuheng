package player_test

import (
	"strings"
	"testing"

	"github.com/genshinsim/gcsim/pkg/core"
)

func TestSetVerdantDewLogsVerdantDew(t *testing.T) {
	c, err := core.New(core.Opt{Seed: 1, Debug: true})
	if err != nil {
		t.Fatal(err)
	}
	c.Player.SetMoonridgeDew(1)
	c.Player.SetVerdantDew(2)
	b, err := c.Log.Dump()
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, "verdant dew set to 2") {
		t.Fatalf("log %s", s)
	}
	if strings.Contains(s, "verdant dew set to 1") {
		t.Fatalf("verdant dew log used the moonridge value: %s", s)
	}
	if !strings.Contains(s, "moonridge dew set to 1") {
		t.Fatalf("log %s", s)
	}
}
