package config

type JiraConfigType interface {
	GetEmail() string
	GetUserToken() string
	GetAtlassianURL() string
	GetOrgID() string
	GetTeamID() string
	GetJiraProject() string
}
