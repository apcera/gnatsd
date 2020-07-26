// Copyright 2012-2019 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !windows

package server

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/shirou/gopsutil/process"
)

var processName = "nats-server"

// SetProcessName allows to change the expected name of the process.
func SetProcessName(name string) {
	processName = name
}

// Signal Handling
func (s *Server) handleSignals() {
	if s.getOpts().NoSigs {
		return
	}
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGHUP)

	go func() {
		for {
			select {
			case sig := <-c:
				s.Debugf("Trapped %q signal", sig)
				switch sig {
				case syscall.SIGINT:
					s.Shutdown()
					os.Exit(0)
				case syscall.SIGUSR1:
					// File log re-open for rotating file logs.
					s.ReOpenLogFile()
				case syscall.SIGUSR2:
					go s.lameDuckMode()
				case syscall.SIGHUP:
					// Config reload.
					if err := s.Reload(); err != nil {
						s.Errorf("Failed to reload server configuration: %s", err)
					}
				}
			case <-s.quitCh:
				return
			}
		}
	}()
}

// ProcessSignal sends the given signal command to the given process. If pidStr
// is empty, this will send the signal to the single running instance of
// nats-server. If multiple instances are running, it returns an error. This returns
// an error if the given process is not running or the command is invalid.
func ProcessSignal(command Command, pidStr string) error {
	var pid int
	if pidStr == "" {
		pids, err := resolvePids()
		if err != nil {
			return err
		}
		if len(pids) == 0 {
			return fmt.Errorf("no %s processes running", processName)
		}
		if len(pids) > 1 {
			errStr := fmt.Sprintf("multiple %s processes running:\n", processName)
			prefix := ""
			for _, p := range pids {
				errStr += fmt.Sprintf("%s%d", prefix, p)
				prefix = "\n"
			}
			return errors.New(errStr)
		}
		pid = pids[0]
	} else {
		p, err := strconv.Atoi(pidStr)
		if err != nil {
			return fmt.Errorf("invalid pid: %s", pidStr)
		}
		pid = p
	}

	var err error
	switch command {
	case CommandStop:
		err = kill(pid, syscall.SIGKILL)
	case CommandQuit:
		err = kill(pid, syscall.SIGINT)
	case CommandReopen:
		err = kill(pid, syscall.SIGUSR1)
	case CommandReload:
		err = kill(pid, syscall.SIGHUP)
	case commandLDMode:
		err = kill(pid, syscall.SIGUSR2)
	default:
		err = fmt.Errorf("unknown signal %q", command)
	}
	return err
}

// resolvePids returns the pids for all running nats-server processes.
func resolvePids() ([]int, error) {
	// myPid
	mp := int32(os.Getpid())
	// nats Pids
	np := []int{}
	// every Pid on system
	ep, err := process.Pids()
	if err != nil {
		return np, err
	}
	for _, pid := range ep {
		prc, err := process.NewProcess(int32(pid))
		if err == process.ErrorProcessNotRunning {
			// process went away since we gathered pids, this is not a crisis.
			continue
		}
		if err != nil {
			return np, err
		}
		name, err := prc.Name()
		if err != nil {
			return np, err
		}
		if name == processName && pid != mp {
			np = append(np, int(pid))
		}
	}
	return np, nil
}

var kill = func(pid int, signal syscall.Signal) error {
	return syscall.Kill(pid, signal)
}
