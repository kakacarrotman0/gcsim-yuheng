package vodyanitsa

import (
	"fmt"
	"slices"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/gcs/validation"
)

// Registration matches the pipeline's zz_*.dm.go shape.
// Talent numbers live in talents.go because excel-hk4e vendored in go.mod predates Vodyanitsa,
// so `go run ./pipeline` cannot emit this character without rewriting every other character.

func init() {
	core.RegisterCharFunc(keys.Vodyanitsa, NewChar)
	paramsFor := map[action.Action][]string{
		action.ActionLowPlunge:  {"collision"},
		action.ActionHighPlunge: {"collision"},
	}
	validation.RegisterCharParamValidationFunc(keys.Vodyanitsa, func(a action.Action, keys []string) error {
		valid, ok := paramsFor[a]
		if !ok {
			return nil
		}
		for _, v := range keys {
			if !slices.Contains(valid, v) {
				return fmt.Errorf("key %v is invalid for action %v", v, a)
			}
		}
		return nil
	})
}
