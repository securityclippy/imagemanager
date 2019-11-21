package dockerhub

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
)

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
