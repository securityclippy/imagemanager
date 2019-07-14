package manager


func (m Manager) SnykMonitorDocker(imageName string) error {
	args := []string{
		"monitor",
		"--docker",
		imageName,
	}
	return runCMD("snyk", args, nil)
}


