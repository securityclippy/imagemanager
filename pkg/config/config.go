package config


type Config struct {
	DockerhubOrg string `json:"dockerhub_org"`
	//days since last activity to mark repository for deprecation
	DeprecationThresholdDays int `json:"deprecation_threshold_days"`
	DeletionThresholdDays int `json:"deletion_threshold_days"`
	//Slack channel for notifications
	SlackChannel string `json:"slack_channel"`
	//amount of days of warning before a repository is deprecated
	DeprecationWarningDays int `json:"deprecation_warning_days"`
	DeletionWarningDays int `json:"deletion_warning_days"`
	DatabaseBucket string `json:"database_bucket"`
	ElasticsearchEndpoint string `json:"elasticsearch_endpoint"`
	ElasticsearchIndex string `json:"elasticsearch_index"`
}
