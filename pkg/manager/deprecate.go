package manager

import (
	"fmt"
	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/securityclippy/imagemanager/pkg/dockerhub"
	"strings"
	"sync"
)

func (m Manager) Deprecate(imageName string) error {
	input := &ecr.DescribeRepositoriesInput{
		RepositoryNames: aws.StringSlice([]string{imageName}),
	}
	result, err := m.ECR.DescribeRepositories(input)
	if err != nil {
		return err
	}
	if result.Repositories == nil {
		return fmt.Errorf("no repository named: %s found", imageName)
	}
	repo := result.Repositories[0]

	images, err := m.DescribeImages(nil, repo)
	if err != nil {
		return err
	}

	oldImages := []*ecr.ImageDetail{}

	// if we have any images
	if len(images) > 0 {
		imageWithTag := ""
		for _, i := range images {
			if len(i.ImageTags) > 0 {
				imageWithTag = fmt.Sprintf("%s:%s", *repo.RepositoryUri, *i.ImageTags[0])
			} else {
				imageWithTag = *repo.RepositoryUri
			}
			err = m.DockerPull(imageWithTag, false)
			if err != nil {
				return err
			}

			_, err := m.DockerRename(imageWithTag, false)
			if err != nil {
				return err
			}

			oldImages = append(oldImages, i)
		}
	}

	err = m.DeleteECRImages(oldImages, repo)

	if err != nil {
		return err
	}

	return nil
}

func (m Manager) DeprecateDockerhub(imageName string, threads int) error {

	tags, err := m.Hub.ListTags(imageName)

	rlChan := make(chan int, threads)
	wg := sync.WaitGroup{}
	tagStream := make(chan *dockerhub.Tag, len(tags))
	for _, t := range tags {
		tagStream <- t
	}
	close(tagStream)

	//oldImages := []string{}
	if len(tags) > 0 {
		for i := 0;  i <= len(tagStream); i++  {
			rlChan <- i
			wg.Add(1)
			go func() {
				defer wg.Done()
				t := <- tagStream
				if t == nil {
					<- rlChan
					return
				}
				if strings.ContainsAny(t.Name, "__deprecated") {
					<- rlChan
					return
				}
				taggedImage := fmt.Sprintf("%s:%s", imageName, t.Name)
				orgImage := fmt.Sprintf("segment/%s", taggedImage)
				err = m.DockerPull(orgImage, false)
				if err != nil {
					log.Error(err.Error())
				}
				_, err = m.DockerRename(orgImage, false)
				if err != nil {
					log.Error(err.Error())
				} else {
					s := strings.Split(taggedImage, ":")
					fmt.Printf("deleting tag: %+v\n", s)
					err = m.Hub.DeleteTag(s[0], s[1])
					if err != nil {
						log.Error(err.Error())
					}
				}

				//oldImages = append(oldImages, taggedImage)
				fmt.Printf("deprected: %d/%d tags\r", <-rlChan, len(tags))
			}()
		}
		wg.Wait()
	}

	project, err := m.Snk.GetProject(fmt.Sprintf("%s/%s", "segment", imageName))
	if err != nil {
		return err
	}

	err =  m.Snk.DeleteProjectByID(project.ID)
	if err != nil {
		return err
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

//TODO
//Get Repository
//Get repository images
// for images
// pull image
// rename image -> deprecated
// push to ecr
// if success
// add to list of batch delete
// else
// add to failure list
// print out results
// if success, delete from snyk