package manager

import (
	"github.com/securityclippy/gedb/pkg/db"
	"github.com/securityclippy/imagemanager/pkg/dockerhub"
	"github.com/securityclippy/imagemanager/pkg/repositoryreport"
)

// repo -> check for entry in DB.
// if not found, create
//

func (m *Manager) CleanDockerhubRepo(repo *dockerhub.Repository) error {

	// load existing repo report or create new one
	rr, err := m.loadReport(repo.Namespace, repo.Name)
	if err == db.ErrKeyNotFound {
		rr = repositoryreport.NewDockerHub(repo, m.Config)
		err = m.saveReport(repo.Namespace, repo.Name, rr)
		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	rr.LastPushed = repo.LastUpdated

	if rr.DaysSinceLastUpdate() <= 30 {
		err := m.saveReport(repo.Namespace, repo.Name, rr)
		if err != nil {
			return err
		}
		return nil
	}

	//check deprecations
	// should be deprecated or warned of dep
	if rr.DaysToDeprecationDate() <= m.Config.DeprecationWarningDays {//} && rr.DaysToDeletionDate() > 0 {
		// check if already marked for dep
		// if zero, not marked for dep yet
		if rr.DaysSinceDeprecationMark() == -1 {
			rr.MarkForDeprecation()
			//TODO add better slack integration
			//m.Bot.Log.Infof("[%s/%s] marked for deprecation \n " +
				//"Last Updated: %d days ago (%.2f years)",
				//repo.Namespace, repo.Name, rr.DaysSinceLastUpdate(), (float64(rr.DaysSinceLastUpdate())/float64(365)))
			err := m.ESC.UpsertInterface(rr, m.Config.ElasticsearchIndex)
			if err != nil {
				m.Log.Error(err)
			}
			err = m.saveReport(repo.Namespace, repo.Name, rr)
			if err != nil {
				return err
			}
			return nil
		}

		if rr.DaysSinceDeprecationMark() <= m.Config.DeprecationWarningDays {
			//TODO better slack integration
			//m.Bot.Log.Infof("[%s/%s] %d days till deprecation \n " +
				//"Last Updated: %d days ago (%.2f years)",
				//repo.Namespace, repo.Name, rr.DaysToDeprecationDate(), rr.DaysSinceLastUpdate(), (float64(rr.DaysSinceLastUpdate())/float64(365)))
			return nil
		}

		//deprecate()
		if !rr.Deprecated {
			err = m.DeprecateDockerhub(repo.Name, 5)
			if err != nil {
				return err
			}
			rr.Deprecated = true
			//TODO
			//m.Bot.Log.Infof("[%s/%s] deprecated", repo.Namespace, repo.Name)
			//m.Bot.Log.Infof("[%s/%s] days to deletion: %d", rr.DaysToDeletionDate())
			err := m.ESC.UpsertInterface(rr, m.Config.ElasticsearchIndex)
			if err != nil {
				m.Log.Error(err)
			}
			err = m.saveReport(repo.Namespace, repo.Name, rr)
			if err != nil {
				return err
			}
			return nil
		}

		if rr.DaysToDeletionDate() > 0 {
			//TODO
			//m.Bot.Log.Infof("[%s/%s] days to deletion: %d", rr.DaysToDeletionDate())
			return nil
		}

		//delete
		//m.Bot.Log.Infof("[%s/%s] days to deleted")

		err = m.Hub.DeleteRepository(repo)
		if err != nil {
			return err
		}


	}

	return nil

}

