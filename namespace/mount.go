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

type Mount struct {
	cmd *cobra.Command
	sh  *exec.Cmd
}

func MountCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mount",
		Short: "mount 隔离",
		RunE: func(cmd *cobra.Command, args []string) error {
			r := Mount{cmd: cmd}
			return r.run()
		},
	}
	var chroot bool
	cmd.Flags().BoolVarP(&chroot, "chroot", "c", false, "使用chroot")

	return cmd
}

func (m *Mount) run() error {
	cmd := exec.Command("sh")
	m.sh = cmd
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if m.cmd.Flags().Lookup("chroot").Value.String() == "true" {
		err := m.chroot()
		if err != nil {
			return err
		}
	} else {
		// 这里会拿一个假的盖上来
		syscall.Mount("none", "/root/namespace/test", "tmpfs", 0, "")
	}

	if err := m.sh.Run(); err != nil {
		return err
	}
	return nil
}

func (m *Mount) chroot() error {
	return m.createDir()
}

func (m *Mount) createDir() error {
	home, err := os.Getwd()
	if err != nil {
		log.Println("获取home错误", err.Error())
		return err
	}

	dir := fmt.Sprintf("%v/%v", home, "container")
	bin := fmt.Sprintf("%v/%v", dir, "bin")
	lib64 := fmt.Sprintf("%v/%v", dir, "lib64")
	lib := fmt.Sprintf("%v/%v", dir, "lib")

	dirs := []string{dir, bin, lib64, lib}
	for _, d := range dirs {
		err = util.CreateDir(d)
		if err != nil {
			log.Println("创建文件夹错误：", d, err.Error())
			return err
		}
	}

	// copy bash ls
	err = exec.Command("sh", "-c", fmt.Sprintf("cp -v /bin/{bash,ls} %v", bin)).Run()
	if err != nil {
		log.Println("copy bash ls", err.Error())
		return err
	}

	// copy so
	err = exec.Command("sh", "-c", fmt.Sprintf("for i in $(ldd /bin/ls | egrep -o '/lib.*\\.[0-9]'); do cp -v \"$i\" \"%v${i}\"; done", dir)).Run()
	if err != nil {
		log.Println("copy so", err.Error())
		return err
	}

	//// chroot
	//err = exec.Command("sh", "-c", fmt.Sprintf("chroot %v /bin/bash", dir)).Run()
	//if err != nil {
	//	log.Println("chroot", err.Error())
	//	return err
	//}
	return nil
}
