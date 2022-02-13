package namespace

import (
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"syscall"
)

type PID struct {
	cmd *cobra.Command
}

func PIDCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pid",
		Short: "pid 隔离",
		RunE: func(cmd *cobra.Command, args []string) error {
			r := &PID{cmd: cmd}
			return r.run()
		},
	}

	var tty bool
	cmd.Flags().BoolVarP(&tty, "shell", "s", false, "enable shell")
	return cmd
}

func (r *PID) run() error {
	cmd := r.getOSCmd()
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (r *PID) getOSCmd() *exec.Cmd {
	f := r.cmd.Flags().Lookup("shell").Value

	switch f.String() {
	case "true":
		return exec.Command("sh")
	}
	return exec.Command("sh", "-c", "echo pid=$$")
}
