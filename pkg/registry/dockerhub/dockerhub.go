package dockerhub

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"fmt"
	"io"
)

type DockerhubRegistryPlugin struct {
	*http.Client
	header http.Header
	authToken string
	url string
	org string
	username string
	password string
	enabled bool
	regConf *RegistryConfig
}

func NewDockerhubRegistryPlugin(regConfig *RegistryConfig) *DockerhubRegistryPlugin {
	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	h := http.Header{}
	h.Add("Content-Type", "application/json")
	p := &DockerhubRegistryPlugin{
		Client: httpClient,
		url:    "https://hub.docker.com",
		org:    regConfig.Organization,
		username: os.Getenv("DH_USERNAME"),
		password: os.Getenv("DH_PASSWORD"),
		regConf:regConfig,
	}

	if p.username == "" {
		logrus.Fatalf("DH_USERNAME environment variable cannot be empty")
	}
	if p.password == "" {
		logrus.Fatalf("DH_PASSWORD environment variable cannot be empty")
	}

	return p

}

func (p *DockerhubRegistryPlugin) getAuthToken() (token string, err error) {
		login := struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: p.username,
			Password: p.password,
		}
		js, err := json.Marshal(login)
		if err != nil {
			return "", err

		}
		req, err := http.NewRequest("POST", "https://hub.docker.com/v2/users/login/", bytes.NewReader(js))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")

		req.Header = p.header
		resp, err := p.Client.Do(req)
		if err != nil {
			return "", err
		}

		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return "", err
		}

		tokenResp := struct {
			Token string `json:"token"`
		}{}

		err = json.Unmarshal(body, &tokenResp)
		if err != nil {
			return "", err
		}
		p.authToken = tokenResp.Token
		if tokenResp.Token == "" {
			fmt.Println("authentication failed")
		}

		return tokenResp.Token, nil
}

func (p *DockerhubRegistryPlugin) auth() error {
	if p.authToken == "" {
		token, err := p.getAuthToken()
		if err != nil {
			return err
		}

		p.authToken = token
	}
	return nil
}

func (p *DockerhubRegistryPlugin) NewRequest(method, url string, payload io.Reader) (*http.Request, error) {
	if err := p.auth(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, err
	}
	header := http.Header{}
	header.Set("Authorization", fmt.Sprintf("JWT %s", p.authToken))
	req.Header = header
	//req.Header.Set("Authorization", fmt.Sprintf("JWT %s", c.AuthToken))
	return req, nil
}

func (c *DockerhubRegistryPlugin) doRequest(method, url string, payload io.Reader) (data []byte, status int, err error) {
	req, err := c.NewRequest(method, url, payload)
	if err != nil {
		return nil, 0, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil,0,  err
	}

	return body, resp.StatusCode, nil
}

/*
func (c *Client) Catalog() error {
	url := fmt.Sprintf("%s/v2/repositories", c.URL, c.ORG)
	fmt.Printf("URL:%s", url)
	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Body:%s", string(body))
	return nil
}
 */


func (p *DockerhubRegistryPlugin) ListRepositories() (repos []*Repository, err error) {
	repos = []*Repository{}
	output, err := p.listRepositoriesRequest("")
	if err != nil {
		return nil, err
	}

	//fmt.Printf("next: %s\n", output.Next)

	repos = append(repos, output.Results...)

	next := output.Next
	for {
		if next == "" {
			return repos, nil
		}
		output, err := p.listRepositoriesRequest(next)
		if err != nil {
			return nil, err
		}
		next = output.Next
		repos = append(repos, output.Results...)
	}
}

func (p *DockerhubRegistryPlugin) GetRepository(name string) (*Repository, error) {

	baseURL := "https://hub.docker.com/v2/repositories/segment"
	url := fmt.Sprintf("%s/%s", baseURL, name)

	data, status, err := p.doRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	if status != 200 {
		return nil, fmt.Errorf("status: %d", status)
	}

	repo := &Repository{}

	err = json.Unmarshal(data, repo)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (p *DockerhubRegistryPlugin) DeleteRepository(repo *Repository) error {
	baseURL := "https://hub.docker.com/v2/repositories/segment"
	url := fmt.Sprintf("%s/%s", baseURL, repo.Name)

	data, status, err := p.doRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	fmt.Printf("Status: %d\n", status)

	fmt.Printf("Response: %s\n", string(data))

	return nil
}

func (p *DockerhubRegistryPlugin) listRepositoriesRequest(next string) (*RepositoryOutput, error) {
	baseURL := "https://hub.docker.com/v2/repositories/segment/?page=1&page_size=100"
	var url string
	if next != "" {
		url = fmt.Sprint(next)
	} else {
		url = baseURL
	}

	req, err := p.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	output := &RepositoryOutput{}

	err = json.Unmarshal(body, output)
	if err != nil {
		return nil, err
	}

	return output, nil
}




func (p *DockerhubRegistryPlugin) GetManifest(image, tag string) (string, error) {
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/segment/%s/manifests/%s", image, tag)
	fmt.Printf("URL: %s\n", url)
	req, err := p.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	resp, err := p.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)

	fmt.Printf("Resp body: %s", string(body))
	return "", nil
}



type RepositoryOutput struct {
	Count int `json:"count"`
	Next string `json:"next"`
	Previous string `json:"previous"`
	Results []*Repository `json:"results"`

}

func (r RepositoryOutput) JSON() string {
	js, _ := json.MarshalIndent(r, "", "  ")
	return string(js)
}

type Repository struct {
	User string `json:"user"`
	Name string `json:"name"`
	Namespace string `json:"namespace"`
	RepositoryType string `json:"repository_type"`
	Status int `json:"status"`
	Description string `json:"description"`
	IsPrivate bool `json:"is_private"`
	IsAutomated bool `json:"is_automated"`
	CanEdit bool `json:"can_edit"`
	StarCount int `json:"star_count"`
	PullCount int `json:"pull_count"`
	LastUpdated time.Time `json:"last_updated"`
	IsMigrated bool `json:"is_migrated"`
}


func (r Repository) JSON() string {
	js, _ := json.MarshalIndent(r, "", "  ")
	return string(js)
}


type TagOutput struct {
	Count int `json:"count"`
	Next string `json:"next"`
	Results []*Tag `json:"results"`
}

type Tag struct {
	Name string `json:"name"`
	FullSize int `json:"full_size"`
	Images []*Image `json:"images"`
	ID int `json:"id"`
	Repository int `json:"repository"`
	Creator int `json:"creator"`
	LastUpdater int `json:"last_updater"`
	LastUpdated time.Time `json:"last_updated"`
	ImageID string `json:"image_id"`
	V2 bool `json:"v2"`
}

type Image struct {
	Size int `json:"size"`
	Architecture string `json:"architecture"`
	Variant string `json:"variant"`
	Features string `json:"features"`
	OS string `json:"os"`
	OSVersion string `json:"os_version"`
	OSFeatures string `json:"os_features"`
}

