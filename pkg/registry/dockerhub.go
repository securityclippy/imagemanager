package registry

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"fmt"
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
	header.Set("Authorization", fmt.Sprintf("JWT %s", c.AuthToken))
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
	output, err := c.listRepositoriesRequest("")
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
		output, err := c.listRepositoriesRequest(next)
		if err != nil {
			return nil, err
		}
		next = output.Next
		repos = append(repos, output.Results...)
	}
}

func (c *Client) GetRepository(name string) (*Repository, error) {

	baseURL := "https://hub.docker.com/v2/repositories/segment"
	url := fmt.Sprintf("%s/%s", baseURL, name)

	data, status, err := c.doRequest("GET", url, nil)

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

func (c *Client) DeleteRepository(repo *Repository) error {
	baseURL := "https://hub.docker.com/v2/repositories/segment"
	url := fmt.Sprintf("%s/%s", baseURL, repo.Name)

	data, status, err := c.doRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	fmt.Printf("Status: %d\n", status)

	fmt.Printf("Response: %s\n", string(data))

	return nil
}

func (c *Client) listRepositoriesRequest(next string) (*RepositoryOutput, error) {
	baseURL := "https://hub.docker.com/v2/repositories/segment/?page=1&page_size=100"
	var url string
	if next != "" {
		url = fmt.Sprint(next)
	} else {
		url = baseURL
	}

	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Do(req)
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




func (c *Client) GetManifest(image, tag string) (string, error) {
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/segment/%s/manifests/%s", image, tag)
	fmt.Printf("URL: %s\n", url)
	req, err := c.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	resp, err := c.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)

	fmt.Printf("Resp body: %s", string(body))
	return "", nil
}




