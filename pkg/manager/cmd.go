package manager

import (
	"bufio"
	"github.com/pkg/errors"
	"os/exec"
	"strings"
	log "github.com/sirupsen/logrus"

)


func runCMD(command string, cmdArgs, outputFilters []string) error {
	cmd := exec.Command(command, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "could not open stdout")
	}
	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			text := scanner.Text()
			lowerText := strings.ToLower(text)
			if func() bool {
				if outputFilters == nil {
					return true
				}
				for _, of := range outputFilters {
					if strings.Contains(lowerText, of) {
						return true
					}
				}
				return false
			}() {
				log.Infof("[%s] %s\n", command, text)
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		return errors.Wrap(err, "could not start cmd")
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}


