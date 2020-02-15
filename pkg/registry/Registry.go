package registry

import (
	"github.com/aquasecurity/trivy/pkg/report"
)

type RegistryPlugin interface {
	ListRepositories() (interface{}, error)
	ListRepositoryTags(repoName string) (tags []string, err error)
	DeprecateRepository(repoName string) (error)
	DeprecateTag(repoName, tag string) (error)
	DeleteRepository(repoName string) (error)
	Scan() ([]*report.Results, error)
	ScanRepository(repoName string) (*report.Results, error)
	Name() string
}


type RegistryConfig struct {
	//Name
	Name string `json:"name"`
	// ecr/dockerhub/tbd
	Type string `json:"type"`
	//Days to keep old tags
	MaxTagAge int `json:"max_tag_age"`
	//Days to keep images. Images more than X days old will be deleted
	MaxImageAge int `json:"max_image_age"`
	//Days to keep repository
	MaxRepoAge int `json:"max_repo_age"`
	// repositories in this list (partial or full match) will not be subjet to lifecycle rules
	RepositoryIgnoreList []string `json:"repository_ignore_list"`
	//Organization name (if needed, eg dockerhub/myorg)
	Organization string `json:"organization"`
	//Enable or disable registry
	Enabled bool `json:"enabled"`

}