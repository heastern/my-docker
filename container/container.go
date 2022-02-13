package container

import (
	"fmt"
	"linux/util"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
)

type R struct {
	cmd  *cobra.Command
	sh   *exec.Cmd
	name string
	home string
}

func ContainerCommand() []*cobra.Command {
	r := &R{}
	cmd := &cobra.Command{
		Use:   "container",
		Short: "run a container",
		RunE: func(cmd *cobra.Command, args []string) error {
			r.cmd = cmd
			return r.run()
		},
	}
	cmd.Flags().StringVarP(&r.name, "name", "n", "container_test", "mount name")
	cmd.Flags().StringVar(&r.home, "home", "", "mount home")

	cmd2 := &cobra.Command{
		Use:   "run",
		Short: "run a container and mount",
		RunE: func(cmd *cobra.Command, args []string) error {
			r.cmd = cmd
			return r.mount()
		},
	}
	return []*cobra.Command{cmd, cmd2}
}

func (r *R) run() error {

	// /proc/self/exe 会使用后面的命来重新跑一遍当前程序
	cmd := exec.Command("/proc/self/exe", "run")
	// cmd := exec.Command("/bin/bash")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// 设置一些系统进程属性，下面这行代码负责创建一个新的独立进程
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}
	r.sh = cmd

	// 运行命令并捕获错误
	if err := cmd.Run(); err != nil {
		return err

	}

	// return nil
	return r.mount()
}

func (r *R) mount() error {
	must(r.initDir())

	must(syscall.Chroot(filepath.Join(r.home, r.name)))

	must(syscall.Chdir(r.name))

	must(syscall.Mount("proc", "proc", "proc", 0, ""))
	must(syscall.Mount("thing", "mytemp", "tmpfs", 0, ""))

	cmd := exec.Command("/bin/bash")
	// attatching os-std process to our cmd-std process
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// running the command and catching error
	if err := cmd.Run(); err != nil {
		log.Fatal("Error: ", err)
	}
	// unmount the proc after command is finished
	must(syscall.Unmount("proc", 0))
	must(syscall.Unmount("thing", 0))

	return nil
}

func (r *R) copy() error {

	dir := filepath.Join(r.home, r.name)
	must(util.CreateDir(dir))

	// if exist, _ := util.PathExists(dir); exist {
	// 	return nil
	// }

	must(exec.Command("/bin/bash", "-c", fmt.Sprintf("cp /bin %v/ -R", dir)).Run())

	must(exec.Command("/bin/bash", "-c", fmt.Sprintf("cp /lib %v/ -R", dir)).Run())

	must(exec.Command("/bin/bash", "-c", fmt.Sprintf("cp /lib64 %v/ -R", dir)).Run())

	must(util.CreateDir(filepath.Join(dir, "proc")))

	must(util.CreateDir(filepath.Join(dir, "mytemp")))

	return nil
}

func (r *R) initDir() error {
	if r.home == "" {
		home, err := os.Getwd()
		if err != nil {
			return err
		}
		r.home = home
	}
	must(syscall.Sethostname([]byte(r.name)))
	return r.copy()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
