package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	tree, err := getProcessTree(4)
	if err != nil {
		fmt.Printf("cannot get process tree: %v\n", err)
		os.Exit(2)
	}

	current := tree[0]
	parents := tree[1:]

	fmt.Println("Process tree:")
	fmt.Print(formatProcessTree(tree))
	//fmt.Println("----------------")
	//
	//fmt.Println("Testing wait for parent processes:")
	//fmt.Println("")
	//
	//for _, p := range parents {
	//	fmt.Printf("Testing attach to %s(%v) from %s(%v)...\n", p.name, p.pid, current.name, current.pid)
	//	testWait(p.pid)
	//	fmt.Println("")
	//}
	//fmt.Println("----------------")
	//
	//fmt.Println("Testing ptrace parent processes:")
	//for _, p := range parents {
	//	fmt.Printf("Testing ptrace %s(%v) from %s(%v)...\n", p.name, p.pid, current.name, current.pid)
	//	testPtrace(p.pid)
	//	fmt.Println("")
	//}
	//fmt.Println("----------------")
	//
	//p := sleepChild(10)
	//
	//fmt.Println("Testing wait for child process:")
	//fmt.Println("")
	//
	//fmt.Printf("Testing attach to %s(%v) from %s(%v)...\n", p.name, p.pid, current.name, current.pid)
	//testWait(p.pid)
	//fmt.Println("")
	//fmt.Println("----------------")
	//
	//fmt.Println("Testing ptrace child process:")
	//fmt.Printf("Testing ptrace %s(%v) from %s(%v)...\n", p.name, p.pid, current.name, current.pid)
	//testPtrace(p.pid)
	//fmt.Println("")
	//fmt.Println("----------------")
	//
	//fmt.Println("----------------")

	time.Sleep(5 * time.Second)
	fmt.Println("Testing ptrace parent process (requires root):")
	grandpa := parents[2]
	fmt.Printf("Testing ptrace %s(%v) from %s(%v)...\n", grandpa.name, grandpa.pid, current.name, current.pid)
	testPtraceSeize(grandpa.pid)
	fmt.Printf("Testing wait %s(%v) from %s(%v)...\n", grandpa.name, grandpa.pid, current.name, current.pid)

	time.AfterFunc(30*time.Second, func() { os.Exit(10) })
	for {
		status, err := testWait4(grandpa.pid)
		if err != nil {
			fmt.Printf("\twait4 failed: %v\n", err)
		}

		fmt.Printf("%v exited? %t\n", grandpa.pid, status.Exited())
		fmt.Printf("%v exitstatus? %v\n", grandpa.pid, status.ExitStatus())
		fmt.Printf("%v signaled? %t\n", grandpa.pid, status.Signaled())
		fmt.Printf("%v stop signal? %s\n", grandpa.pid, status.StopSignal().String())
		fmt.Printf("%v stopped? %t\n", grandpa.pid, status.Stopped())
		fmt.Printf("%v signal? %s\n", grandpa.pid, status.Signal().String())
		fmt.Printf("%v continued? %t\n", grandpa.pid, status.Continued())
		fmt.Printf("%v core dump? %t\n", grandpa.pid, status.CoreDump())
		fmt.Printf("%v trap cause? %v\n", grandpa.pid, status.TrapCause())

		if status.StopSignal() == syscall.SIGTRAP && status.TrapCause() == syscall.PTRACE_EVENT_EXEC {
			fmt.Println("execve event detected")
		} else {
			fmt.Printf("execve event not detected: %s, %v\n", status.StopSignal().String(), status.TrapCause())
		}
		err = syscall.PtraceSyscall(pid, 0)
		if err != nil {
			fmt.Printf("\tinitial trace failed: %v\n", err)
			return
		} else {
			fmt.Println("\tinitial trace succeeded")
		}
	}
}

func testWait(pid int) {
	// wait4
	if _, err := testWait4(pid); err != nil {
		fmt.Printf("\twait4 failed: %v\n", err)
	} else {
		fmt.Println("\twait4 succeeded")
		return
	}

	// waitpid
	if err := testWaitPID(pid); err != nil {
		fmt.Printf("\twaitpid failed: %v\n", err)
	} else {
		fmt.Println("\twaitpid succeeded")
	}
}

func testPtraceSeize(pid int) {
	// In soviet russia, ptrace seizes you

	// also there are some differences in behavior here, but unsure if they are relevant?
	// note that ptrace’s behavious changes in some cases when switching to the SEIZE method.
	// Unlike PTRACE_ATTACH, detecting a group stop is possible without calling
	// PTRACE_GETSIGINFO under PTRACE_SEIZE.
	err := unix.PtraceAttach(pid)
	if err != nil {
		fmt.Printf("\tseize trace failed: %v\n", err)
		return
	} else {
		fmt.Println("\tseize trace succeeded")
	}

	err = checkProcessRunning(pid)
	if err != nil {
		fmt.Println("\tprocess not running")
	}
	time.Sleep(time.Millisecond)

	// Reference: https://man7.org/linux/man-pages/man2/ptrace.2.html
	opts := unix.PTRACE_O_TRACEEXEC
	err = unix.PtraceSetOptions(pid, opts)
	if err != nil {
		fmt.Printf("\tfailed to set ptrace options: %v\n", err)
		return
	} else {
		fmt.Println("\tset ptrace options succeeded")
	}

	//err = unix.PtraceInterrupt(pid)
	//if err != nil {
	//	fmt.Printf("\tseize interrupt failed: %v\n", err)
	//	return
	//} else {
	//	fmt.Println("\tseize interrupt succeeded")
	//}

	err = syscall.PtraceSyscall(pid, 0)
	if err != nil {
		fmt.Printf("\tinitial trace failed: %v\n", err)
		return
	} else {
		fmt.Println("\tinitial trace succeeded")
	}

	return
}

func checkProcessRunning(pid int) error {
	// Send signal 0 to check if the process is still running
	err := syscall.Kill(pid, 0)
	if err != nil {
		return fmt.Errorf("process not running: %w", err)
	}
	return nil
}

func testIntercept(pid int) {
	return
}

func testPtrace(pid int) {
	// Reference: https://man7.org/linux/man-pages/man2/ptrace.2.html
	opts := unix.PTRACE_O_TRACESYSGOOD |
		unix.PTRACE_O_TRACEEXIT |
		unix.PTRACE_O_EXITKILL |
		unix.PTRACE_O_TRACECLONE |
		unix.PTRACE_O_TRACEEXEC |
		unix.PTRACE_O_TRACEFORK |
		unix.PTRACE_O_TRACEVFORK |
		unix.PTRACE_O_TRACEVFORKDONE

	err := unix.PtraceSetOptions(pid, opts)
	if err != nil {
		fmt.Printf("\tfailed to set ptrace options: %v\n", err)
		return
	} else {
		fmt.Println("\tset ptrace options succeeded")
	}

	err = syscall.PtraceSyscall(pid, 0)
	if err != nil {
		fmt.Printf("\tinitial trace failed: %v\n", err)
		return
	} else {
		fmt.Println("\tinitial trace succeeded")
	}
}

func getProcessName(pid int) (string, error) {
	cmdlinePath := filepath.Join("/proc", fmt.Sprint(pid), "cmdline")
	cmdline, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return "", fmt.Errorf("cannot read cmdline: %w", err)
	}
	// The cmdline file contains the command line arguments separated by null bytes.
	// The first argument is the name of the executable.
	name := strings.SplitN(string(cmdline), "\x00", 2)[0]
	return name, nil
}

type process struct {
	pid  int
	name string
}

func getProcessTree(depth int) ([]process, error) {
	tree := []process{}
	currentPid := os.Getpid()
	for _ = range depth {
		op, err := exec.Command("ps", "-o", "ppid=", "-p", fmt.Sprint(currentPid)).CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("cannot get parent pid: %w", err)
		}
		name, err := getProcessName(currentPid)
		if err != nil {
			return nil, fmt.Errorf("cannot get pid name: %w", err)
		}

		tree = append(tree, process{pid: currentPid, name: name})
		currentPid, _ = strconv.Atoi(strings.TrimSpace(string(op)))
	}

	return tree, nil
}

func formatProcessTree(tree []process) string {
	var sb strings.Builder
	for _, p := range tree {
		sb.WriteString(fmt.Sprintf("%8v: %v\n", p.pid, p.name))
	}
	return sb.String()
}

func sleepChild(seconds int) process {
	sleepCmd := exec.Command("sleep", fmt.Sprint(seconds))
	sleepCmd.SysProcAttr = &syscall.SysProcAttr{Ptrace: true}

	sleepCmd.Start()

	return process{
		pid:  sleepCmd.Process.Pid,
		name: "sleep",
	}
}

func testWait4(pid int) (*syscall.WaitStatus, error) {
	var status syscall.WaitStatus
	_, err := syscall.Wait4(pid, &status, syscall.WALL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot wait4: %w", err)
	}
	return &status, nil
}

func testWaitPID(pid int) error {
	var info unix.Siginfo
	err := unix.Waitid(unix.P_PID, pid, &info, unix.WEXITED|unix.WCONTINUED|unix.WSTOPPED, &unix.Rusage{})
	if err != nil {
		return fmt.Errorf("cannot waitpid: %w", err)
	}
	return nil
}
