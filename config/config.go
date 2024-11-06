package config

import (
	"log"
	"os"
	"tui/utils"

	"github.com/joho/godotenv"
)

type JiraCredConfig struct {
	Email          string
	UserToken      string
	AtlassianURL   string
	OrganizationID string
	TeamID         string
	JiraProject    string
}

func NewConfig() JiraConfigType {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading env: %v", err)
		return nil
	}

	email := os.Getenv("ATLASSIAN_USER_EMAIL")
	userToken := os.Getenv("ATLASSIAN_USER_TOKEN")
	baseURL := os.Getenv("ATLASSIAN_URL")
	orgId := os.Getenv("ATLASSIAN_ORGANIZATION_ID")
	teamId := os.Getenv("ATLASSIAN_TEAM_ID")
	project := os.Getenv("ATLASSIAN_PROJECT")

	if env, notValid := utils.ValidateENVs(map[string]string{
		"email":     email,
		"userToken": userToken,
		"baseURL":   baseURL,
		"orgId":     orgId,
		"teamId":    teamId,
		"project":   project,
	}); notValid {
		log.Fatalf("Error setup env config: %v", *env)
		return nil
	}

	return &JiraCredConfig{
		Email:          email,
		UserToken:      userToken,
		AtlassianURL:   baseURL,
		OrganizationID: orgId,
		TeamID:         teamId,
		JiraProject:    project,
	}
}

// GetAtlassianURL implements JiraConfigType.
func (j *JiraCredConfig) GetAtlassianURL() string {
	return j.AtlassianURL
}

// GetEmail implements JiraConfigType.
func (j *JiraCredConfig) GetEmail() string {
	return j.Email
}

// GetJiraProject implements JiraConfigType.
func (j *JiraCredConfig) GetJiraProject() string {
	return j.JiraProject
}

// GetOrgID implements JiraConfigType.
func (j *JiraCredConfig) GetOrgID() string {
	return j.OrganizationID
}

// GetTeamID implements JiraConfigType.
func (j *JiraCredConfig) GetTeamID() string {
	return j.TeamID
}

// GetUserToken implements JiraConfigType.
func (j *JiraCredConfig) GetUserToken() string {
	return j.UserToken
}
