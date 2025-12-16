//go:build linux

package nspawndriver_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shoenig/test/must"
	"github.com/stackshadow/nspawn2/nspawndriver"
	"golang.org/x/sys/unix"
)

var stdErrFifo string
var stdOutFifo string
var testPid int

func TestMain(m *testing.M) {
	stdErrFifo = filepath.Join(os.TempDir(), "test-fifo-stderr")
	if err := CreateFIFO(stdErrFifo, 0600); err != nil {
		panic(err)
	}
	defer os.Remove(stdErrFifo)

	stdOutFifo = filepath.Join(os.TempDir(), "test-fifo-stdout")
	if err := CreateFIFO(stdOutFifo, 0600); err != nil {
		panic(err)
	}
	defer os.Remove(stdOutFifo)

	code := m.Run()
	os.Exit(code)
}

func TestStartBackgroundWithFIFO(t *testing.T) {
	cmds := []string{"sleep", "50"}

	go readFifoOnce(t, stdErrFifo)
	go readFifoOnce(t, stdOutFifo)

	cmder, err := nspawndriver.NewCommander(nspawndriver.ExecBackgroundWithFIFOOpts{
		Commands:       cmds,
		StdErrFifoPath: stdErrFifo,
		StdOuFifooPath: stdOutFifo,
	})
	defer cmder.Destroy()

	if err != nil {
		t.Fatalf("StartBackgroundWithFIFO failed: %v", err)
	}

	// Prüfen, ob Prozess noch läuft
	if !cmder.IsAlive() {
		t.Fatalf("Prozess läuft nicht")
	}

	// Kurz warten, bis sleep etwas schreiben könnte (sleep schreibt nichts, aber Test prinft)
	time.Sleep(500 * time.Millisecond)

	// Prozess nach Test killen
	err = cmder.Stop()
	must.NoError(t, err)
}

// CreateFIFO legt eine FIFO mit den gegebenen Rechten an,
// falls sie noch nicht existiert.
func CreateFIFO(path string, mode os.FileMode) error {
	// Falls bereits da: nichts tun, aber prüfen, ob es wirklich eine FIFO ist
	if info, err := os.Stat(path); err == nil {
		if info.Mode()&os.ModeNamedPipe == 0 {
			return fmt.Errorf("%s existiert, ist aber keine FIFO", path)
		}
		return nil
	}

	// Mode nach Unix (nur Permissions‑Bits)
	perm := uint32(mode.Perm())
	if err := unix.Mkfifo(path, perm); err != nil {
		return fmt.Errorf("mkfifo %s: %w", path, err)
	}
	return nil
}

func readFifoOnce(t *testing.T, fifoPath string) string {
	f, err := os.OpenFile(fifoPath, os.O_RDONLY, 0600)
	if err != nil {
		return ""
	}
	defer f.Close()

	buf := make([]byte, 1024)
	n, _ := f.Read(buf)

	t.Log(string(buf[:n]))

	return string(buf[:n])
}
