package nspawndriver

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
	"time"

	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/hashicorp/nomad/plugins/shared/structs"
)

// Fingerprint returns a channel that will be used to send health information
// and other driver specific node attributes.
func (d *NSpawnDriverPlugin) Fingerprint(ctx context.Context) (<-chan *drivers.Fingerprint, error) {
	ch := make(chan *drivers.Fingerprint)
	go d.handleFingerprint(ctx, ch)
	return ch, nil
}

// handleFingerprint manages the channel and the flow of fingerprint data.
func (d *NSpawnDriverPlugin) handleFingerprint(ctx context.Context, ch chan<- *drivers.Fingerprint) {
	defer close(ch)

	// Nomad expects the initial fingerprint to be sent immediately
	ticker := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			// after the initial fingerprint we can set the proper fingerprint
			// period
			ticker.Reset(fingerprintPeriod)
			ch <- d.buildFingerprint()
		}
	}
}

// buildFingerprint returns the driver's fingerprint data
func (d *NSpawnDriverPlugin) buildFingerprint() *drivers.Fingerprint {
	fp := &drivers.Fingerprint{
		Attributes:        map[string]*structs.Attribute{},
		Health:            drivers.HealthStateHealthy,
		HealthDescription: drivers.DriverHealthy,
	}

	// TODO: implement fingerprinting logic to populate health and driver
	// attributes.
	//
	// Fingerprinting is used by the plugin to relay two important information
	// to Nomad: health state and node attributes.
	//
	// If the plugin reports to be unhealthy, or doesn't send any fingerprint
	// data in the expected interval of time, Nomad will restart it.
	//
	// Node attributes can be used to report any relevant information about
	// the node in which the plugin is running (specific library availability,
	// installed versions of a software etc.). These attributes can then be
	// used by an operator to set job constrains.
	//
	// In the example below we check if the shell specified by the user exists
	// in the node.

	cmd := exec.Command("whereis", "--version")
	if err := cmd.Run(); err != nil {
		return &drivers.Fingerprint{
			Health:            drivers.HealthStateUndetected,
			HealthDescription: "whereis not found",
			Err:               errors.New("whereis not found"),
		}
	}

	cmd = exec.Command("whereis", "systemd-nspawn")
	if err := cmd.Run(); err != nil {
		return &drivers.Fingerprint{
			Health:            drivers.HealthStateUndetected,
			HealthDescription: "systemd-nspawn not found",
			Err:               errors.New("systemd-nspawn not found"),
		}
	}

	cmd = exec.Command("whereis", "machinectl")
	if err := cmd.Run(); err != nil {
		return &drivers.Fingerprint{
			Health:            drivers.HealthStateUndetected,
			HealthDescription: "machinectl not found",
			Err:               errors.New("machinectl not found"),
		}
	}

	// We also set the shell and its version as attributes
	cmd = exec.Command("systemd-nspawn", "--version")
	if out, err := cmd.Output(); err != nil {
		d.logger.Warn("failed to find systemd-nspawn", err)
	} else {
		re := regexp.MustCompile("[0-9]{3}")
		version := re.FindString(string(out))

		fp.Attributes["driver.nspawn2.version"] = structs.NewStringAttribute(pluginVersion)
		fp.Attributes["driver.nspawn2.systemd_version"] = structs.NewStringAttribute(version)
	}

	return fp
}
