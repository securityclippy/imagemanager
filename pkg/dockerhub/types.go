package dockerhub

import (
	"encoding/json"
	"time"
)

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
