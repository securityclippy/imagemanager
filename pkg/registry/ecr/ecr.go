package ecr

import "github.com/aws/aws-sdk-go/service/ecr"

type ECRRegistryPlugin struct {
	*ecr.ECR
}
