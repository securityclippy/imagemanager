package manager

import (
	"errors"
	"github.com/aws/aws-sdk-go/service/ecr"
	"sync"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
)


//func (m *Manager) ECRLogin() error {
	//m.Img.
//}

var errNoImages = errors.New("no images found")

func (m *Manager) DeleteECRImages(images []*ecr.ImageDetail, repo *ecr.Repository) error {
	imageIds := []*ecr.ImageIdentifier{}
	for _, image := range images {
		iID := &ecr.ImageIdentifier{
			ImageDigest: image.ImageDigest,
			ImageTag: image.ImageTags[0],

		}
		imageIds = append(imageIds, iID)
	}
	input := &ecr.BatchDeleteImageInput{
		ImageIds: imageIds,
		RepositoryName: repo.RepositoryName,
		RegistryId: repo.RegistryId,
	}
	result, err := m.ECR.BatchDeleteImage(input)

	if err != nil {
		return err
	}

	if result.Failures != nil {
		for _, fail := range result.Failures {
			m.Log.Errorf("Delete failed: %+v\n", fail)
		}
	}

	return nil
}


func (m Manager) UpdateECR(start, finish, threads int, verbose bool) error {
	repos, err := m.DescribeRepositories()
	if err != nil {
		m.Log.Fatal(err)
	}
	if finish == 0 {
		finish = len(repos)
	}
	m.Log.Infof("Found: %d repos", len(repos))
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

			err := m.UpdateECRImage(repo, verbose)
			if err == errNoImages {
				//m.Bot.Log.Infof("[%s ] No images found. Repo should be deleted\n", *repo.RepositoryName)
			}
			if err != nil {
				m.Log.Error(err)
			}
			fmt.Printf("finished: %d/%d\n", i, len(repos))
			<- rlChan
		}()

	}
	close(rlChan)
	wg.Wait()
	return nil
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

func (m Manager) UpdateECRImage(repo *ecr.Repository, verbose bool) error {
	images, err := m.DescribeImages(nil, repo)
	if err != nil {
		return err
	}
	if len(images) == 0 {
		return errNoImages
	}

	latest, err := m.LatestImage(images)
	if err != nil {
		return err
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
		return err
	}

	err = m.SnykMonitorDocker(imageWithTag, verbose)
	if err != nil {
		return err
	}

	err = m.DockerRMI(imageWithTag)
	if err != nil {
		return err
	}
	m.Log.Infof("Finished: [%s]\n", *repo.RepositoryName)
	return nil
}
