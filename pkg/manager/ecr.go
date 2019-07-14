package manager

import (
	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/service/ecr"
)

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
			log.Errorf("Delete failed: %+v\n", fail)
		}
	}

	return nil
}
