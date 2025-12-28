package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/nomad/helper/uuid"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/shoenig/test/must"
	"github.com/stackshadow/nspawn2/nspawndriver"
	testutils "github.com/stackshadow/nspawn2/pkg/TestUtils"
)

func TestService(t *testing.T) {

	basePath := testutils.GetProjectPath("./")

	task := &drivers.TaskConfig{
		ID:      uuid.Generate(),
		Name:    "stattest",
		AllocID: uuid.Generate(),
	}
	taskCfg := nspawndriver.TaskConfig{
		MachineName: "test",
		Image:       basePath + "/debian.raw",
		Boot:        true,
		Ephemeral:   true,
	}
	must.NoError(t, task.EncodeConcreteDriverConfig(&taskCfg))

	d := testutils.DriverHarness(t)
	cleanup := d.MkAllocDir(task, false)
	defer cleanup()

	_, _, err := d.StartTask(task)
	must.NoError(t, err)
	defer func() {
		err = d.StopTask(task.ID, time.Second*30, "")
		must.NoError(t, err)

		result, err := d.WaitTask(t.Context(), task.ID)
		<-result
		must.NoError(t, err)

		err = d.DestroyTask(task.ID, true)
		must.NoError(t, err)
	}()

	// get machine Name
	stats, err := d.TaskStats(context.Background(), task.ID, time.Second)
	<-stats
	<-stats
	<-stats

}
