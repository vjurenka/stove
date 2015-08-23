package pegasus

import (
	"github.com/HearthSim/hs-proto-go/pegasus/util"
	"testing"
)

func TestUtilRegisterPacket(t *testing.T) {
	for _, x := range []struct {
		Enum   interface{}
		ID     int32
		System int32
	}{
		{util.BuySellCard_ID, 257, 0},
		{util.GuardianVars_ID, 264, 0},
		{util.GetPurchaseMethod_ID, 250, 1},
	} {
		packetId := packetIDFromProto(x.Enum)
		if packetId.ID != x.ID {
			t.Errorf("bad packet id: %d != %d", packetId.ID, x.ID)
		}
		if packetId.System != x.System {
			t.Errorf("bad system id: %d != %d", packetId.System, x.System)
		}
	}
}
