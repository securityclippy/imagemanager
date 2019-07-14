package manager

import (
	"fmt"
	"github.com/apex/log"
)

const (
	dockerCmd = "docker"
)

func (m Manager) DockerPull(imageName string, verbose bool) error {
	args := []string{
		"pull",
		imageName,
	}
	if verbose {
		return runCMD("docker", args, []string{"digest", "status"})
	}
	return runCMD("docker", args, []string{"done"})
}

func (m Manager) DockerPush(image string, verbose bool) error {
	pushArgs := []string{
		"push",
		image,
	}
	if verbose {
		return runCMD("docker", pushArgs, nil)
	}
	return runCMD(dockerCmd, pushArgs, []string{"done"})

}

func (m Manager) DockerTag(oldTag, newTag string, verbose bool) error {

	args := []string{
		"tag",
		oldTag,
		newTag,
	}
	if verbose {
		return runCMD(dockerCmd, args, nil)
	}
	return runCMD(dockerCmd, args, []string{"done"})
}

func (m Manager) DockerRMI(imageName string) error {
	args := []string {
		"rmi",
		imageName,
	}
	return runCMD("docker", args, []string{"status"})
}


func (m Manager) DockerRename(image string, verbose bool) (renamed string, err error) {
	deprecated := fmt.Sprintf("%s__deprecated", image)

	err = m.DockerTag(image, deprecated, verbose)
	if err != nil {
		return "", err
	}

	err = m.DockerPush(deprecated, verbose)
	if err != nil {
		return "", err
	}

	return deprecated, nil
}

func (m Manager) GetManifest(image, tag string) {
	_, err := m.Hub.GetManifest(image, tag)
	if err != nil {
		log.Error(err.Error())
	}
}

func (m Manager) ListTags(image string) {
	tags, err := m.Hub.ListTags(image)
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, t := range tags {
		fmt.Printf("%+v\n", t)
	}
}

func (m Manager) DockerhubDeleteTag(image, tag string) error {
	err := m.Hub.DeleteTag(image, tag)
	if err != nil {
		return err
	}
	return nil
}