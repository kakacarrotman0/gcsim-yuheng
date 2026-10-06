package alyosha

import (
	"github.com/genshinsim/gcsim/pkg/core/event"
)

func (c *char) initC1() {
	if c.Base.Cons < 1 {
		return
	}
	restore := func(_ ...any) {
		if c.StatusIsActive(c1ICDKey) {
			return
		}
		c.AddStatus(c1ICDKey, c1Interval, true)
		c.AddEnergy(c1ICDKey, c1Energy)
	}
	for _, ev := range []event.Event{
		event.OnOverload,
		event.OnElectroCharged,
		event.OnLunarCharged,
		event.OnSuperconduct,
		event.OnSwirlElectro,
		event.OnCrystallizeElectro,
		event.OnStellarConduct,
	} {
		c.Core.Events.Subscribe(ev, restore, "alyosha-c1")
	}
}
