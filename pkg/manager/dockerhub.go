package manager

import (
	"fmt"
	"github.com/schollz/progressbar/v2"
	"github.com/securityclippy/imagemanager/pkg/dockerhub"
	"strings"
	"sync"
)


// DeleteDockerhubRepository delete a repository from dockerhub and snyk
func (m *Manager) DeleteDockerhubRepository(repoName string) error {
	repo, err := m.Hub.GetRepository(repoName)
	if err != nil {
		return err
	}

	//err = m.Hub.DeleteRepository(repo)
	//if err != nil {
		//return err
	//}
	projName := fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)
	snyProj, err := m.Snk.GetProject(projName)
	if err != nil {
		return err
	}

	return m.Snk.DeleteProjectByID(snyProj.ID)
}

func (m Manager) DeprecateDockerhub(imageName string, threads int) error {

	tags, err := m.Hub.ListTags(imageName)

	rlChan := make(chan int, threads)
	wg := sync.WaitGroup{}
	//tagStream := make(chan *dockerhub.Tag, len(tags))
	//close(tagStream)

	tagBar := progressbar.New(len(tags))

	fmt.Printf("Deprecating: %s\n", imageName)
	deprecateTag := func(t *dockerhub.Tag, rlChan chan int, wg *sync.WaitGroup) {
		defer wg.Done()
		if t == nil {
			<- rlChan
			return
		}

		if strings.HasSuffix(t.Name, "__deprecated") {
			<- rlChan
			return
		}

		//fmt.Printf("deprecating: %s\n", t.Name)
		taggedImage := fmt.Sprintf("%s:%s", imageName, t.Name)
		orgImage := fmt.Sprintf("segment/%s", taggedImage)
		err = m.DockerPull(orgImage, false)
		if err != nil {
			m.Log.Error(err)
		}
		//fmt.Printf("deprecating: %s\n", orgImage)
		_, err = m.DockerRename(orgImage, false)
		if err != nil {
			m.Log.Error(err)
		} else {
			s := strings.Split(taggedImage, ":")
			//fmt.Printf("deleting tag: %+v\n", s)
			err = m.Hub.DeleteTag(s[0], s[1])
			if err != nil {
				m.Log.Error(err)
			}
			err = m.DockerRMI(orgImage)
			if err != nil {
				m.Log.Error(err)
			}
			depped := fmt.Sprintf("%s__deprecated", orgImage)
			err = m.DockerRMI(depped)
			if err != nil {
				m.Log.Error(err)
			}
			//fmt.Printf("removed: %s\n", orgImage)
		}
		tagBar.Add(1)
		<- rlChan
	}
	for _, t := range tags {
		wg.Add(1)
		rlChan <- 1
		go deprecateTag(t, rlChan, &wg)
		//tagStream <- t
	}
	wg.Wait()


	project, err := m.Snk.GetProject(fmt.Sprintf("%s/%s", "segment", imageName))
	if err != nil {
		m.Log.Error(err)
	}

	if project != nil {
		err =  m.Snk.DeleteProjectByID(project.ID)
		if err != nil {
			m.Log.Error(err)
		}
	}
	/*for _, i := range oldImages {
		s := strings.Split(i, ":")
		err = m.Hub.DeleteTag(s[0], s[1])
		if err != nil {
			return err
		}
	}*/
	return nil
}

func (m *Manager) UnDeprecateDockerhubRepo(repoName string, threads int) error {
	tags, err := m.Hub.ListTags(repoName)

	if err != nil {
		return err
	}

	rlChan := make(chan int, threads)
	wg := sync.WaitGroup{}
	//tagStream := make(chan *dockerhub.Tag, len(tags))
	//close(tagStream)

	tagBar := progressbar.New(len(tags))

	fmt.Printf("UnDeprecating: %s\n", repoName)
	tagBar.Add(0)
	for _, t := range tags {
		wg.Add(1)
		rlChan <- 1
		go m.UnDeprecateDockhubTag(repoName, t.Name, rlChan, &wg, tagBar)
	}

	wg.Wait()

	return nil
}

func (m *Manager) UnDeprecateDockhubTag(imageName, tag string, rlchan chan int, wg *sync.WaitGroup, pb *progressbar.ProgressBar) error {


	// bail if its not a deprecated tag
	if !strings.Contains(tag, "__deprecated") {
		pb.Add(1)
		<- rlchan
		wg.Done()
		return nil
	}

	repoName := fmt.Sprintf("segment/%s", imageName)

	_, err := m.DockerUnDeprecate(repoName, tag, false)
	if err != nil {
		return err
	}

	if err := m.DockerhubDeleteTag(imageName, tag); err != nil {
		pb.Add(1)
		wg.Done()
		<- rlchan
		return err
	}

	pb.Add(1)
	<- rlchan
	wg.Done()
	return nil
}
