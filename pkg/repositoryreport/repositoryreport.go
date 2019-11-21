package repositoryreport

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/securityclippy/imagemanager/pkg/config"
	"github.com/securityclippy/imagemanager/pkg/dockerhub"
	"time"
)

type RepositoryReport struct {
	Registry string `json:"registry"`
	Repository string `json:"repository"`
	LastPushed time.Time `json:"last_pushed"`
	LastPulled string `json:"last_pulled"`
	Deprecated bool `json:"deprecated"`
	DeprecationDate time.Time `json:"deprecation_date"`
	DeprecationMarkedOn time.Time `json:"deprecation_marked_on"`

	DeletionDate time.Time `json:"deletion_date"`
	Public bool `json:"public"`

}

func (rr *RepositoryReport) JsonString() string {
	js, _ := json.MarshalIndent(rr, "", "  ")
	return string(js)
}


func (rr *RepositoryReport) DaysToDeprecationDate() int {
	if rr.DeprecationDate.IsZero() {
		return 365
	}
	return int(rr.DeprecationDate.Sub(time.Now()).Hours()/24)
}

func (rr *RepositoryReport) DaysToDeprecation() int {
	if rr.DaysToDeprecationDate() <= 0 {

	}
	return rr.DaysSinceDeprecationMark()
}

func (rr *RepositoryReport) DaysToDeletionDate() int {
	if rr.DeletionDate.IsZero() {
		return 365
	}
	return int(rr.DeletionDate.Sub(time.Now()).Hours()/24)
}

func (rr *RepositoryReport) DaysSinceLastUpdate() int {

	return int(time.Now().Sub(rr.LastPushed).Hours()/24)
	//return int(rr.LastPushed.Sub(time.Now()).Hours()/24)
}

func (rr *RepositoryReport) DaysSinceDeprecationMark() int {
	if rr.DeprecationMarkedOn.IsZero() {
		return -1
	}
	return int(time.Now().Sub(rr.DeprecationMarkedOn).Hours()/24)
}

func (rr *RepositoryReport) SetDeprecationDate(cfg *config.Config) {
	//t := time.Duration(cfg.DeprecationThresholdDays * 24) * time.Hour
	rr.DeprecationDate = time.Now().Add(time.Duration(cfg.DeprecationWarningDays * 24) *time.Hour)
	//rr.DeprecationDate = rr.LastPushed.Add(t)
}

func (rr *RepositoryReport) SetDeletionDate(cfg *config.Config) {
	//t := time.Duration(cfg.DeletionThresholdDays * 24) * time.Hour
	rr.DeletionDate = time.Now().Add(time.Duration(cfg.DeletionWarningDays * 24) * time.Hour)
}

func (rr *RepositoryReport) MarkForDeprecation() {
	rr.DeprecationMarkedOn = time.Now()
}

func New(registry, repository string) *RepositoryReport {
	return &RepositoryReport{
		Registry:registry,
		Repository:repository,
	}
}

func NewDockerHub(repo *dockerhub.Repository, conf *config.Config) *RepositoryReport {
	rr := New(repo.Namespace, repo.Name)
	rr.LastPushed = repo.LastUpdated
	rr.Public = !repo.IsPrivate
	rr.SetDeprecationDate(conf)
	rr.SetDeletionDate(conf)
	return rr
}

func NewECR(repo *ecr.Repository, latestImageDetail *ecr.ImageDetail) *RepositoryReport {
	rr := New(*repo.RegistryId, *repo.RepositoryName)
	rr.LastPushed = *latestImageDetail.ImagePushedAt
	rr.Public = false
	return rr
}
