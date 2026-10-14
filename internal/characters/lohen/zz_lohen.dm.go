// Hand-maintained talent tables from released live data.
// Yatta avatar 10000129. Do not treat this file as pipeline output.

package lohen

import (
	"fmt"
	"slices"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/action"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/gcs/validation"
)

func init() {
	core.RegisterCharFunc(keys.Lohen, NewChar)
	paramsFor := map[action.Action][]string{
		action.ActionSkill: {"etched"},
	}
	validation.RegisterCharParamValidationFunc(keys.Lohen, func(a action.Action, keys []string) error {
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

var (
	// normal 1-Hit
	attack1 = []float64{0.539917, 0.583863, 0.62781, 0.690591, 0.734538, 0.784763, 0.853822, 0.922881, 0.99194, 1.067277, 1.142614, 1.217951, 1.293289, 1.368626, 1.443963}
	// normal 2-Hit
	attack2 = []float64{0.564427, 0.610368, 0.65631, 0.721941, 0.767883, 0.820387, 0.892582, 0.964776, 1.03697, 1.115727, 1.194484, 1.273241, 1.351999, 1.430756, 1.509513}
	// normal 3-Hit
	attack3 = []float64{0.254199, 0.274889, 0.29558, 0.325138, 0.345829, 0.369475, 0.401989, 0.434503, 0.467016, 0.502486, 0.537956, 0.573425, 0.608895, 0.644364, 0.679834}
	// normal 4-Hit
	attack4 = []float64{0.752259, 0.81349, 0.87472, 0.962192, 1.023422, 1.0934, 1.189619, 1.285838, 1.382058, 1.487024, 1.59199, 1.696957, 1.801923, 1.90689, 2.011856}
	// normal 5-Hit A
	attack5 = []float64{0.368579, 0.398579, 0.42858, 0.471438, 0.501439, 0.535725, 0.582869, 0.630013, 0.677156, 0.728586, 0.780016, 0.831445, 0.882875, 0.934304, 0.985734}
	// normal 5-Hit B
	attack6 = []float64{0.552868, 0.597869, 0.64287, 0.707157, 0.752158, 0.803588, 0.874303, 0.945019, 1.015735, 1.092879, 1.170023, 1.247168, 1.324312, 1.401457, 1.478601}
	// normal Charged Attack
	attack7 = []float64{0.65876, 0.71238, 0.766, 0.8426, 0.89622, 0.9575, 1.04176, 1.12602, 1.21028, 1.3022, 1.39412, 1.48604, 1.57796, 1.66988, 1.7618}
	// charged attack stamina
	chargeStamina = []float64{25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25, 25}
	// plunge collision
	collision = []float64{0.639324, 0.691362, 0.7434, 0.81774, 0.869778, 0.92925, 1.011024, 1.092798, 1.174572, 1.26378, 1.352988, 1.442196, 1.531404, 1.620612, 1.70982}
	// low plunge
	lowPlunge = []float64{1.278377, 1.382431, 1.486485, 1.635134, 1.739187, 1.858106, 2.02162, 2.185133, 2.348646, 2.527025, 2.705403, 2.883781, 3.062159, 3.240537, 3.418915}
	// high plunge
	highPlunge = []float64{1.596762, 1.726731, 1.8567, 2.04237, 2.172339, 2.320875, 2.525112, 2.729349, 2.933586, 3.15639, 3.379194, 3.601998, 3.824802, 4.047606, 4.27041}
	// masterstroke 1-Hit
	enhanced1 = []float64{0.809875, 0.875795, 0.941715, 1.035887, 1.101807, 1.177144, 1.280732, 1.384321, 1.48791, 1.600915, 1.713921, 1.826927, 1.939933, 2.052939, 2.165944}
	// masterstroke 2-Hit
	enhanced2 = []float64{0.84664, 0.915552, 0.984465, 1.082912, 1.151824, 1.230581, 1.338872, 1.447164, 1.555455, 1.67359, 1.791726, 1.909862, 2.027998, 2.146134, 2.264269}
	// masterstroke 3-Hit
	enhanced3 = []float64{0.381298, 0.412334, 0.44337, 0.487707, 0.518743, 0.554212, 0.602983, 0.651754, 0.700525, 0.753729, 0.806933, 0.860138, 0.913342, 0.966547, 1.019751}
	// masterstroke 4-Hit
	enhanced4 = []float64{1.128389, 1.220234, 1.31208, 1.443288, 1.535134, 1.6401, 1.784429, 1.928758, 2.073086, 2.230536, 2.387986, 2.545435, 2.702885, 2.860334, 3.017784}
	// masterstroke 5-Hit A
	enhanced5 = []float64{0.552868, 0.597869, 0.64287, 0.707157, 0.752158, 0.803588, 0.874303, 0.945019, 1.015735, 1.092879, 1.170023, 1.247168, 1.324312, 1.401457, 1.478601}
	// masterstroke 5-Hit B
	enhanced6 = []float64{0.829302, 0.896804, 0.964305, 1.060736, 1.128237, 1.205381, 1.311455, 1.417528, 1.523602, 1.639319, 1.755035, 1.870752, 1.986468, 2.102185, 2.217901}
	// masterstroke Charged Attack
	enhanced7 = []float64{0.98814, 1.06857, 1.149, 1.2639, 1.34433, 1.43625, 1.56264, 1.68903, 1.81542, 1.9533, 2.09118, 2.22906, 2.36694, 2.50482, 2.6427}
	// masterstroke charged stamina
	enhancedStamina = []float64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
	// masterstroke seconds
	masterDuration = []float64{13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13, 13}
	// joy per normal hit
	joyOnNormal = []float64{17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17}
	// joy per charged hit
	joyOnCharge = []float64{17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 17}
	// will success threshold, multiple of base ATK
	willAtkRatio = []float64{10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}
	// will gained on a qualifying hit
	willSuccess = []float64{20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20, 20}
	// etched hit, four times
	raid = []float64{0.6, 0.645, 0.69, 0.75, 0.795, 0.84, 0.9, 0.96, 1.02, 1.08, 1.14, 1.2, 1.275, 1.35, 1.425}
	// damage bonus per point of will
	willRatio = []float64{0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004}
	// burst hit, six times
	burstDMG = []float64{1.188, 1.2771, 1.3662, 1.485, 1.5741, 1.6632, 1.782, 1.9008, 2.0196, 2.1384, 2.2572, 2.376, 2.5245, 2.673, 2.8215}
	// burst damage bonus per point of will
	burstWillRatio = []float64{0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004, 0.004}
)
