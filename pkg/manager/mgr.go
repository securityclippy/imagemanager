package manager

import (
	"fmt"
	"github.com/securityclippy/imagemanager/pkg/config"
	"github.com/securityclippy/imagemanager/pkg/dockerhub"
	"github.com/aws/aws-sdk-go/aws"
	//"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/securityclippy/imagemanager/pkg/registry"
	"github.com/aquasecurity/fanal/cache"
	"github.com/securityclippy/imagemanager/pkg/storage"
	"os"
	ospkgScanner "github.com/aquasecurity/trivy/pkg/scanner/ospkg"

	ospkgDetector "github.com/aquasecurity/trivy/pkg/detector/ospkg"

	libScanner "github.com/aquasecurity/trivy/pkg/scanner/library"
	libDetector "github.com/aquasecurity/trivy/pkg/detector/library"





	//"github.com/segmentio/cloudsec-bot/pkg/bot"
	"github.com/sirupsen/logrus"
	"github.com/securityclippy/snyker/pkg/snykclient"
	"github.com/securityclippy/esc"
	"sync"
	"github.com/aquasecurity/trivy/pkg/scanner"
)

var trivyScanner scanner.Scanner

func init() {
	log := logrus.New()
	// init cache
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	cacheClient := cache.Initialize(wd)

	if err != nil {
		log.Fatal(err)
	}

	//osPkgScanner init
	osPkgscan := ospkgScanner.NewScanner(ospkgDetector.Detector{})


	//libScan init
	libScan := libScanner.NewScanner(libDetector.Detector{})


	//
	trivyScanner = scanner.NewScanner(cacheClient, osPkgscan, libScan)
}

type Manager struct {
	TrivyScanner scanner.Scanner
	Plugins []registry.RegistryPlugin
	Log *logrus.Logger
	ECR *ecr.ECR
	Hub *dockerhub.Client
	//Bot *bot.Bot
	Config *config.Config
	Storage storage.Storage
	Snk *snykclient.SnykClient
	ESC *esc.ESC
}


func NewManager(dhUsername, dhPassword, snykToken string, conf *config.Config, store storage.Storage) *Manager {
	log := logrus.New()
	//cfg := aws.NewConfig()
	//sess, err := session.NewSession(cfg)
	//if err != nil {
		//log.Fatal(err)
	//}

	//ecrClient := ecr.New(sess)

	hub := dockerhub.NewClient(dhUsername, dhPassword, "", "")
	_, err := hub.GetAuthToken()
	if err != nil {
		log.Fatal(err)
	}

	//snyk, err := snykclient.NewSnykClient(snykToken)
	if err != nil {
		log.Error("could not create snyk client")
	}

	//esClient := esc.NewAWS(conf.ElasticsearchEndpoint)
	m := &Manager{
		Log:log,
		//ECR:ecrClient,
		Hub: hub,
		//Bot: bot.NewBetaBot(),
		Config:conf,
		Storage: store,
		//Snk: snyk,
		//ESC:esClient,
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



func (m *Manager) UpdateDockerhub(threads int) error {

	repos, err := m.Hub.ListRepositories()
	m.Log.Infof("Got: %d repos", len(repos))
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
				m.Log.Error(err)
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
					m.Log.Error(err)
				}

				err = m.SnykMonitorDocker(image, false)
				if err != nil {
					m.Log.Error(err)
				}

				err = m.DockerRMI(image)
				if err != nil {
					m.Log.Error(err)
				}
				fmt.Printf("finished: %d", <- rlChan)

			}
		}()
	}
	close(rlChan)
	wg.Wait()
	return nil
}
