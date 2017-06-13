package cflib

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// modified from here
//https://github.com/cloudfoundry-incubator/cf-networking-release/blob/develop/src/cf-pusher/cf_cli_adapter/adapter.go

type Adapter struct {
	CfCliPath string
}

func (a *Adapter) Push(name, directory string) error {
	fmt.Printf("running: %s push %s -p %s \n", a.CfCliPath, name, directory)
	bytes, err := exec.Command(a.CfCliPath,
		"push", name,
		"-p", directory).CombinedOutput()
	if err != nil {
		fmt.Printf("output: %s\n", string(bytes))
	}
	return err
}

func (a *Adapter) Delete(appName string) error {
	fmt.Printf("running: %s delete -f %s\n", a.CfCliPath, appName)
	cmd := exec.Command(a.CfCliPath, "delete", "-f", appName)
	return runCommandWithTimeout(cmd)
}

func (a *Adapter) AppGuid(name string) (string, error) {
	fmt.Printf("running: %s app %s --guid\n", a.CfCliPath, name)
	bytes, err := exec.Command(a.CfCliPath, "app", name, "--guid").CombinedOutput()
	return strings.TrimSpace(string(bytes)), err
}

type CmdErr struct {
	Out     string
	Err     string
	Message string
}

func (e *CmdErr) Error() string {
	return fmt.Sprintf("%s:\n\nOut:\n%s\n\nErr:%s\n", e.Message, e.Out, e.Err)
}

func runCommandWithTimeout(cmd *exec.Cmd) error {
	outBuffer := &bytes.Buffer{}
	errBuffer := &bytes.Buffer{}
	wrapErr := func(msg string) error {
		return &CmdErr{
			Out:     outBuffer.String(),
			Err:     errBuffer.String(),
			Message: msg,
		}
	}
	cmd.Stdout = outBuffer
	cmd.Stderr = errBuffer
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(2 * time.Minute):
		if err := cmd.Process.Kill(); err != nil {
			return wrapErr(fmt.Sprintf("command timed out and could not be killed: %s", err))
		}
		return wrapErr("command timed out")

	case err := <-done:
		if err != nil {
			return wrapErr(err.Error())
		}
	}
	return nil
}
