package nspawndriver_test

import (
	"testing"
	"time"

	"github.com/stackshadow/nspawn2/nspawndriver"
	"github.com/stretchr/testify/assert"
)

func TestListMachines(t *testing.T) {
	machines, err := nspawndriver.Machines()
	assert.NoError(t, err)
	_ = machines

	// nspawndriver.MachineStateFromName("3fe6132a-8152-077a-6270-0248fbc8cbfc")

	_, err = nspawndriver.MachineWaitForRunning("notexist", time.Second*10)
	assert.Error(t, err)

	// err = nspawndriver.WaitForMachineRunning("3fe6132a-8152-077a-6270-0248fbc8cbfc", time.Second*10)
	// assert.NoError(t, err)

}
