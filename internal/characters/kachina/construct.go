package kachina

import (
	"github.com/genshinsim/gcsim/pkg/core/construct"
	"github.com/genshinsim/gcsim/pkg/core/info"
)

// twirlyConstruct registers Turbo Twirly with the Geo construct limit.
// Its lifetime is the Nightsoul blessing, so Expiry stays unset and the
// character destroys it when the gadget ends.
type twirlyConstruct struct {
	src  int
	char *char
	pos  info.Point
	dir  info.Point
}

func (c *char) spawnConstruct() {
	c.despawnConstruct()
	c.constructSeq++
	player := c.Core.Combat.Player()
	con := &twirlyConstruct{
		src:  c.constructSeq,
		char: c,
		pos:  player.Pos(),
		dir:  player.Direction(),
	}
	c.twirlyCon = con
	c.Core.Constructs.New(con, true)
}

func (c *char) despawnConstruct() {
	if c.twirlyCon == nil {
		return
	}
	con := c.twirlyCon
	c.twirlyCon = nil
	c.Core.Constructs.Destroy(con.Key())
}

func (t *twirlyConstruct) OnDestruct() {
	if t.char.twirlyCon != t {
		return
	}
	t.char.twirlyCon = nil
	if t.char.StatusIsActive(twirlyKey) {
		t.char.endTwirly()
	}
}

func (t *twirlyConstruct) Key() int { return t.src }
func (t *twirlyConstruct) Type() construct.GeoConstructType {
	return construct.GeoConstructKachinaSkill
}
func (t *twirlyConstruct) Expiry() int           { return -1 }
func (t *twirlyConstruct) IsLimited() bool       { return true }
func (t *twirlyConstruct) Count() int            { return 1 }
func (t *twirlyConstruct) Direction() info.Point { return t.dir }
func (t *twirlyConstruct) Pos() info.Point       { return t.pos }
