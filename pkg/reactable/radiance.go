package reactable

import "github.com/genshinsim/gcsim/pkg/core"

const (
	// Radiance: Stellar Swirl lasts 8s. Vodyanitsa A1 extends the duration to 12s
	// only when a character enters the state while Song of Ages Past is active.
	RadianceStellarSwirlBase   = 8 * 60
	RadianceStellarSwirlExtend = 4 * 60
	VodyanitsaSongUntilKey     = "vodyanitsa-song-until"
)

// RadianceSwirlDuration returns the status duration to apply on a Stellar Swirl.
// alreadyActive means the character is already in Radiance: Stellar Swirl, so this
// application is a refresh and does not gain the enter-only extension.
func RadianceSwirlDuration(c *core.Core, alreadyActive bool) int {
	if !alreadyActive && c.Flags.Custom[VodyanitsaSongUntilKey] > float64(c.F) {
		return RadianceStellarSwirlBase + RadianceStellarSwirlExtend
	}
	return RadianceStellarSwirlBase
}
