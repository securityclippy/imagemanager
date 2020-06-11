package dockerhub

import (
	"io/ioutil"
	"fmt"
	"encoding/json"
)

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

func (c *Client) ListRepositoriesPublic() (publicRepos []*Repository, err error) {
	repos := []*Repository{}

	repos, err = c.ListRepositories()

	if err != nil {
		return nil, err
	}

	publicRepos = []*Repository{}

	for _, r := range repos {
		if !r.IsPrivate {
			publicRepos = append(publicRepos, r)
		}
	}

	return publicRepos, nil
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


/*
func (c *Client) DeleteRepository(repo *Repository) error {
	baseURL := "https://hub.docker.com/v2/repositories/segment"
	url := fmt.Sprintf("%s/%s", baseURL, repo.Name)

	data, status, err := c.doRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	fmt.Printf("Status: %d\n", status)

	fmt.Printf("Response: %s\n", string(data))

	req, err := c.NewRequest("DELETE", "https://hub.docker.com/repository/docker/segment/integration-zapier",  nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.Username, c.Password)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}


	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(body))


	return nil
}
 */

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