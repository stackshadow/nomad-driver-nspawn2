package commander

import (
	"fmt"
	"os"
)

func (cmder *Commander) IsAlive() bool {
	_, err := os.Stat(fmt.Sprintf("/proc/%d", cmder.pid))
	return err == nil
}
