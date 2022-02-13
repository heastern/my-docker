package namespace

import (
	"fmt"
	"github.com/spf13/cobra"
	"linux/util"
	"log"
	"os"
	"os/exec"
	"syscall"
)

type CGroups struct {
	cmd *cobra.Command
	sh  *exec.Cmd
}

func CGroupsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cgroups",
		Short: "cgroups 隔离",
		RunE: func(cmd *cobra.Command, args []string) error {
			r := CGroups{cmd: cmd}
			return r.run()
		},
	}
	return cmd
}

func (c *CGroups) run() error {
	c.createSh()
	if err := c.sh.Start(); err != nil {
		return err
	}

	return c.limit()
}

func (c *CGroups) createSh() {
	cmd := exec.Command("sh", "-c", "while : ; do : ; done &")
	c.sh = cmd
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
}

func (c *CGroups) limit() error {
	err := c.cpuLimit()
	if err != nil {
		return err
	}
	return nil
}

func (c *CGroups) cpuLimit() error {
	path := fmt.Sprintf("/sys/fs/cgroup/cpu/%s", "test")
	err := util.CreateDir(path)
	if err != nil {
		return err
	}

	// 写入CPU限制
	err = exec.Command("sh", "-c", fmt.Sprintf("echo 20000 > %v/%v", path, "cpu.cfs_quota_us")).Run()
	if err != nil {
		log.Println("写入CPU限制失败")
		return err
	}
	// 写入PID
	log.Println("pid=", c.sh.Process.Pid+1)

	err = exec.Command("sh", "-c", fmt.Sprintf("echo %v > %v/%v", c.sh.Process.Pid+1, path, "tasks")).Run()
	if err != nil {
		log.Println("写入PID失败")
		return err
	}
	return nil

}

func (c *CGroups) memLimit() {

}
