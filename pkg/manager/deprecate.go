package manager

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
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

/*
func (m *Manager) DockerhubDeleteDeprecated() error {
	repos, err := m.Hub.ListRepositories()
	if err != nil {
		return err
	}
	for _, repo := range repos {
		tags :=
		tagStream := make(chan *dockerhub.Tag, )
	}
}
 */



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