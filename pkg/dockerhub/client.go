package dockerhub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	repositoriesURL = "https://hub.docker.com/v2/repositories"
)

type Client struct {
	*http.Client
	Header http.Header
	AuthToken string
	URL string
	ORG string
	Username string
	Password string
}

func NewClient(username, password, org, url string) *Client {
	c := &http.Client{
		Timeout: time.Second * 30,
	}
	if url == "" {
		url = "https://hub.docker.com"
	}

	h := http.Header{}
	h.Set("Content-Type","application/json")

	return &Client{
		Client: c,
		Header:h,
		URL: url,
		ORG: org,
		Username: username,
		Password: password,
	} }

func (c *Client) GetAuthToken() (token string, err error) {
	login := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: c.Username,
		Password: c.Password,
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

	req.Header = c.Header
	resp, err := c.Client.Do(req)
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
	c.AuthToken = tokenResp.Token
	return "", nil
}

func (c *Client) NewRequest(method, url string, payload io.Reader) (*http.Request, error) {
	if c.AuthToken == "" {
		token, err := c.GetAuthToken()
		if err != nil {
			return nil, err
		}
		c.AuthToken = token
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

func (c *Client) ListRepositories() (repos []*Repository, err error) {
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

func (c *Client) ListTags(imageName string) (tags []*Tag, err error) {
	tags = []*Tag{}
	output, err := c.listImageTags(imageName, "")
	if err != nil {
		return nil, err
	}

	tags = append(tags, output.Results...)
	next := output.Next

	for {
		if next == "" {
			return tags, nil
		}
		output, err := c.listImageTags(imageName, next)
		if err != nil {
			return nil, err
		}

		tags = append(tags, output.Results...)
		next = output.Next
	}
}

func (c *Client) listImageTags(image, next string) (output *TagOutput, err error) {
	baseURL := fmt.Sprintf("https://hub.docker.com/v2/repositories/segment/%s/tags/?page_size=100", image)
	var url string
	if next != "" {
		url = next
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

	output = &TagOutput{}

	err = json.Unmarshal(body, &output)
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

func (c *Client) DeleteTag(image, tag string) error {
	//DELETE "$PROTO$REG/v2/$repo/manifests/$digest"
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/segment/%s/tags/%s/", image, tag)
	req, err := c.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 204 {
		return fmt.Errorf("%s", resp.Status)
	}
	return nil
}

func (c Client) MostRecentTag(tags []*Tag) (tag *Tag) {
	if len(tags) > 0 {
		tag = tags[0]
		for _, t := range tags {
			if t.LastUpdated.After(tag.LastUpdated.UTC()) {
				tag = t
			}
		}
		return tag
	}
	return nil
}



