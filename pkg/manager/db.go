package manager

import (
	"fmt"
	"github.com/securityclippy/imagemanager/pkg/repositoryreport"
)

func (m *Manager) saveReport(org, imageName string, rr *repositoryreport.RepositoryReport) error {
	key := fmt.Sprintf("%s/%s", org, imageName)
	return m.DB.Save(key, rr)
}

func (m *Manager) loadReport(org, imageName string) (*repositoryreport.RepositoryReport, error) {
	key := fmt.Sprintf("%s/%s", org, imageName)

	rr := &repositoryreport.RepositoryReport{}
	err := m.DB.Load(key, rr)
	if err != nil {
		return nil, err
	}

	return rr, nil
}
