package manager

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/securityclippy/imagemanager/pkg/dockerhub"
	"github.com/securityclippy/snyker/pkg/snykclient"
	log "github.com/sirupsen/logrus"
	"sync"
)

type Manager struct {
	ECR *ecr.ECR
	Hub *dockerhub.Client
	Snk *snykclient.SnykClient
}

func NewManager(dhUsername, dhPassword, snykToken string) *Manager {
	cfg := aws.NewConfig()
	sess, err := session.NewSession(cfg)
	if err != nil {
		log.Fatal(err)
	}

	client := ecr.New(sess)

	hub := dockerhub.NewClient(dhUsername, dhPassword, "", "")
	_, err = hub.GetAuthToken()
	if err != nil {
		log.Fatal(err)
	}

	snyk := snykclient.NewSnykClient(snykToken)
	m := &Manager{
		ECR:client,
		Hub: hub,
		Snk: snyk,
	}

	return m
}

func (m Manager) DescribeRepositories() (repos []*ecr.Repository, err error) {
	maxResults := 500
	repos = []*ecr.Repository{}
	result, err := m.ECR.DescribeRepositories(&ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int64(int64(maxResults)),
	})
	if err != nil {
		return nil, err
	}

	repos = append(repos, result.Repositories...)
	fmt.Printf("Got: %d repos\n", len(result.Repositories))
	for {
		if len(result.Repositories) < maxResults {
			repos = append(repos, result.Repositories...)
			return repos, nil
		}
		result, err = m.ECR.DescribeRepositories(&ecr.DescribeRepositoriesInput{
			MaxResults: aws.Int64(int64(maxResults)),
			NextToken: result.NextToken,
		})
		if err != nil {
			return nil, err
		}
		repos = append(repos, result.Repositories...)
		fmt.Printf("Got: %d more repos\n", len(result.Repositories))
	}
}

func (m Manager) ListImageIds(repo *ecr.Repository) (imageIds []*ecr.ImageIdentifier, err error) {
	imageIds = []*ecr.ImageIdentifier{}
	input := &ecr.ListImagesInput{
		MaxResults: aws.Int64(1000),
		RegistryId: repo.RegistryId,
		RepositoryName: repo.RepositoryName,
	}
	result, err := m.ECR.ListImages(input)
	if err != nil {
		return nil, err
	}
	imageIds = append(imageIds, result.ImageIds...)
	next := *result.NextToken
	if next != "" {
		for {
			input := &ecr.ListImagesInput{
				MaxResults: aws.Int64(1000),
				RegistryId: repo.RegistryId,
				RepositoryName: repo.RepositoryName,
				NextToken: aws.String(next),
			}
			result, err := m.ECR.ListImages(input)
			if err != nil {
				return nil, err
			}
			imageIds = append(imageIds, result.ImageIds...)
			next = *result.NextToken
		}
	}
	return imageIds, nil
}

func (m Manager) DescribeImages(imageIds []*ecr.ImageIdentifier, repo *ecr.Repository) (images []*ecr.ImageDetail, err error) {
	maxResults := 500
	images = []*ecr.ImageDetail{}
	input := &ecr.DescribeImagesInput{
		MaxResults: aws.Int64(int64(maxResults)),
		RegistryId: repo.RegistryId,
		RepositoryName:repo.RepositoryName,
	}
	result, err := m.ECR.DescribeImages(input)
	if err != nil {
		return nil, err
	}

	images = append(images, result.ImageDetails...)

	for {
		if len(result.ImageDetails) < maxResults {
			//images = append(images, result.ImageDetails...)
			return images, nil
		}
		input := &ecr.DescribeImagesInput{
			MaxResults: aws.Int64(int64(maxResults)),
			RegistryId: repo.RegistryId,
			RepositoryName:repo.RepositoryName,
			NextToken: result.NextToken,
		}
		result, err = m.ECR.DescribeImages(input)
		if err != nil {
			return nil, err
		}
		images = append(images, result.ImageDetails...)
	}
}

func (m Manager) LatestImage(images []*ecr.ImageDetail) (latest *ecr.ImageDetail, err error) {
	latest = images[0]
	for _, i := range images {
		if latest == nil {
			latest = i
		}
		if latest.ImagePushedAt.Before(*i.ImagePushedAt) {
			latest = i
		}
	}
	return latest, nil
}


func (m Manager) UpdateECR(start, finish, threads int) error {
	repos, err := m.DescribeRepositories()
	if err != nil {
		log.Fatal(err)
	}
	if finish == 0 {
		finish = len(repos)
	}
	log.Infof("Found: %d repos", len(repos))
	rlChan := make(chan int, threads)
	wg := sync.WaitGroup{}
	repoStream := make(chan *ecr.Repository, threads)
	for i, r := range repos[start:finish] {
		rlChan <- 1
		wg.Add(1)
		repoStream <- r
		go func() {
			defer wg.Done()
			repo := <- repoStream
			images, err := m.DescribeImages(nil, repo)
			if err != nil {
				log.Error(err)
			}
			log.Infof("%s: has %d images", *repo.RepositoryName, len(images))
			if len(images) > 0 {
				latest, err := m.LatestImage(images)
				if err != nil {
					log.Error(err)
				}
				//log.Infof("latest: %+v", latest)
				imageWithTag := ""
				if len(latest.ImageTags) > 0 {
					imageWithTag = fmt.Sprintf("%s:%s", *repo.RepositoryUri, *latest.ImageTags[0])
				} else {
					imageWithTag = *repo.RepositoryUri
				}
				//log.Infof("latest w/tags: %s", imageWithTag)
				err = m.DockerPull(imageWithTag, false)
				if err != nil {
					log.Error(err)
				}

				err = m.SnykMonitorDocker(imageWithTag)
				if err != nil {
					log.Error(err)
				}

				err = m.DockerRMI(imageWithTag)
				if err != nil {
					log.Error(err)
				}
				log.Infof("Finished: [%d]", i)
			}
			<- rlChan
		}()

	}
	close(rlChan)
	wg.Wait()
	return nil
}

func (m *Manager) UpdateDockerhub(threads int) error {

	repos, err := m.Hub.ListRepositories()
	log.Infof("Got: %d repos", len(repos))
	if err != nil {
		return err
	}
	rlChan := make(chan int, threads)
	wg := sync.WaitGroup{}
	repoStream := make(chan *dockerhub.Repository, threads)
	for i, r := range repos {
		rlChan <- i
		wg.Add(1)
		repoStream <- r
		go func() {
			defer wg.Done()
			rp := <- repoStream
			tags, err := m.Hub.ListTags(rp.Name)
			if err != nil {
				log.Error(err)
			}

			latest := m.Hub.MostRecentTag(tags)
			if latest != nil {
				image := ""
				if len(latest.Name) > 0 {
					image = fmt.Sprintf("segment/%s:%s", rp.Name, latest.Name)
				} else {
					image = fmt.Sprintf("segment/%s", rp.Name)
				}


				err = m.DockerPull(image, false)
				if err != nil {
					log.Error(err)
				}

				err = m.SnykMonitorDocker(image)
				if err != nil {
					log.Error(err)
				}

				err = m.DockerRMI(image)
				if err != nil {
					log.Error(err)
				}
				fmt.Printf("finished: %d", <- rlChan)

			}
		}()
	}
	close(rlChan)
	wg.Wait()
	return nil
}