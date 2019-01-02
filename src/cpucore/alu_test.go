package cpucore

import (
	"instruction"
	"testing"

	"github.com/romana/rlog"
)

func TestADD(t *testing.T) {
	SetupLogger()
	rlog.Info("TestADD")
	core := Core{}
	core.Init()
	syncSeen, _ := waitForSync(&core)
	if !syncSeen {
		t.Fatal("Sync was not seen")
	}
	// 4 bits max
	accumVal := uint64(0xc)
	register := uint64(0x3)
	// Load the first value into the accumulator
	runOneCycle(&core, uint64(instruction.LDM|accumVal), t)
	// Now put it into register 3
	runOneCycle(&core, uint64(instruction.XCH|register), t)
	// Load the second value into the accumulator
	accumVal = 0x3
	runOneCycle(&core, uint64(instruction.LDM|accumVal), t)
	// Finally, run the alu instruction
	runOneCycle(&core, uint64(instruction.ADD|register), t)
	// The carry bit should NOT be set. Test it with a jump
	verifyJump(&core, uint64(instruction.JCN_CARRY_SET), false, t)

	// Run it again, which will cause the carry bit to be set
	// Finally, run the alu instruction
	runOneCycle(&core, uint64(instruction.ADD|register), t)
	// The carry bit should NOT be set. Test it with a jump
	verifyJump(&core, uint64(instruction.JCN_CARRY_SET), true, t)

}
