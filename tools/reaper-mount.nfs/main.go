package main

import (
	"time"

	ps "github.com/mitchellh/go-ps"
	"github.com/prometheus/common/log"
	"github.com/shirou/gopsutil/process"
)

const (
	targetExecutable = "mount.nfs"
	targetState      = "D"
)

func main() {
	log.Info("Starting NFS fixer")

	limiter := time.Tick(time.Second * 5)

	for {
		<-limiter

		log.Info("Starting new check")

		list, err := ps.Processes()
		if err != nil {
			log.Error("failed to list processes:", err)
			continue
		}

		for _, item := range list {
			if item.Executable() != targetExecutable {
				continue
			}

			// Query this process for morning information about it.
			proc, err := process.NewProcess(int32(item.Pid()))
			if err != nil {
				log.Error("failed to inspect process:", err)
				continue
			}

			status, err := proc.Status()
			if err != nil {
				log.Error("failed to inspect process status:", err)
				continue
			}

			if status != targetState {
				log.Error("process is not ready to be killed:", status)
				continue
			}

			log.Info("killing process:", item.Pid())

			err = proc.Kill()
			if err != nil {
				log.Error("failed to kill the process:", err)
			}
		}
	}
}
