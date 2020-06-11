package manager

import (
	"github.com/securityclippy/gedb/pkg/db"
	"github.com/securityclippy/imagemanager/pkg/dockerhub"
	"github.com/securityclippy/imagemanager/pkg/repositoryreport"
	"strings"
)

// repo -> check for entry in DB.
// if not found, create
//

func (m *Manager) CleanDockerhubRepo(repo *dockerhub.Repository, forceDeprecation bool) error {

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

	err = m.ESC.UpsertInterface(rr, m.Config.ElasticsearchIndex)
	if err != nil {
		m.Log.Error(err)
	}
	err = m.saveReport(repo.Namespace, repo.Name, rr)
	if err != nil {
		return err
	}

	if forceDeprecation {
		rr.MarkForDeprecation()
		m.Bot.Log.Infof("[%s/%s] \n" +
			":warning: marked for deprecation :warning: \n " +
			"Last Updated: %d days ago (%.2f years)",
			repo.Namespace, repo.Name, rr.DaysSinceLastUpdate(), (float64(rr.DaysSinceLastUpdate())/float64(365)))
		err := m.ESC.UpsertInterface(rr, m.Config.ElasticsearchIndex)
		if err != nil {
			m.Log.Error(err)
		}
		err = m.DeprecateDockerhub(repo.Name, 5)
		if err != nil {
			m.Log.Error(err)
			return err
		}
		rr.Deprecated = true
		m.Bot.Log.Infof("[%s/%s] \n:goose-warning: Force deprecated :goose-warning: \n", repo.Namespace, repo.Name)
		m.Bot.Log.Infof("days to deletion: %d\n", rr.DaysToDeletionDate())
		err = m.ESC.UpsertInterface(rr, m.Config.ElasticsearchIndex)
		if err != nil {
			m.Log.Error(err)
		}
		err = m.saveReport(repo.Namespace, repo.Name, rr)
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
			m.Bot.Log.Infof("[%s/%s] \n" +
				":warning: marked for deprecation :warning: \n " +
				"Last Updated: %d days ago (%.2f years)",
				repo.Namespace, repo.Name, rr.DaysSinceLastUpdate(), (float64(rr.DaysSinceLastUpdate())/float64(365)))
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

		if rr.DaysSinceDeprecationMark() < m.Config.DeprecationWarningDays && !rr.Deprecated {
			m.Bot.Log.Infof("[%s/%s] \n " +
				"days till deprecation: %d \n " +
				"Last Updated: %d days ago (%.2f years)",
				repo.Namespace, repo.Name, rr.DaysToDeprecationDate(), rr.DaysSinceLastUpdate(), (float64(rr.DaysSinceLastUpdate())/float64(365)))
			return nil
		}

		//deprecate()
		if !rr.Deprecated && !(strings.Contains(repo.Name, "auth-service")) {
			err = m.DeprecateDockerhub(repo.Name, 10)
			if err != nil {
				return err
			}
			rr.Deprecated = true
			m.Bot.Log.Infof("[%s/%s] :goose-warning: deprecated :goose-warning: ", repo.Namespace, repo.Name)
			m.Bot.Log.Infof("days to deletion: %d", rr.DaysToDeletionDate())
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
			m.Bot.Log.Infof("[%s/%s] \n " +
				"days to deletion: %d", repo.Namespace, repo.Name, rr.DaysToDeletionDate())
			return nil
		}

		//delete
		m.Bot.Log.Infof(":alert: [%s/%s] was deleted! :alert:")

		//err = m.Hub.DeleteRepository(repo)
		//if err != nil {
			//return err
		//}


	}

	return nil

}

