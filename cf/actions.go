package cf

import (
	"log"
	"os/exec"
	"strings"
)

func Push(name, directory, stack, buildpack, startCommand string) error {
	pushArgs := []string{"push", name, "-p", directory, "-s", stack}
	if buildpack != "" {
		pushArgs = append(pushArgs, "-b", buildpack)
	}
	if startCommand != "" {
		pushArgs = append(pushArgs, "-c", startCommand)
	}

	_, err := runCF(pushArgs...)
	return err
}

func Delete(appName string) error {
	_, err := runCF("delete", "-r", "-f", appName)
	return err
}

func PurgeRoutes() error {
	_, err := runCF("delete-orphaned-routes", "-f")
	return err
}

func AppGuid(name string) (string, error) {
	output, err := runCF("app", name, "--guid")
	return strings.TrimSpace(string(output)), err
}

func runCF(args ...string) (string, error) {
	log.Printf("running: cf %s\n", strings.Join(args, " "))
	output, err := exec.Command("cf", args...).CombinedOutput()
	if err != nil {
		log.Printf("error running above command. Output: '%s', error: '%s'\n", string(output), err)
		return "", err
	}
	return string(output), nil
}
