module github.com/securityclippy/imagemanager

go 1.13

require (
	github.com/apex/log v1.1.2
	github.com/aquasecurity/fanal v0.0.0-20200112144021-9a35ce3bd793
	github.com/aquasecurity/trivy v0.4.4
	github.com/aws/aws-sdk-go v1.29.3
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/schollz/progressbar/v2 v2.15.0
	github.com/securityclippy/esc v0.0.0-20191210235700-ca822a1f8aaf
	github.com/securityclippy/snyker v0.0.0-20190726212104-7e0560e41992
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.6.2
	github.com/urfave/cli v1.20.0
)

replace github.com/aquasecurity/trivy => /home/clippy/go/src/github.com/aquasecurity/trivy

replace github.com/genuinetools/reg => github.com/tomoyamachi/reg v0.16.1-0.20190706172545-2a2250fd7c00
