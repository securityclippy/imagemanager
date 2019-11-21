package manager


func (m Manager) SnykMonitorDocker(imageName string, verbose bool) error {
	args := []string{
		"monitor",
		"--docker",
		imageName,
	}
	if verbose {
		return runCMD("snyk", args, nil)
	}
	return runCMD("snyk", args, []string{"alongstringthatshouldnevermatch"})

}


