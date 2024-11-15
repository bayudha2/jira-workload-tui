package services

import (
	"io"
	"net/http"
	"time"
)

type ServiceType interface {
  handleFailedFetch()
	FetchMembers()
	FetchUsers()
	FetchIssues(FetchWorklogPayload) error
	FetchWorklogs(string) (*WorklogField, error)
	GetUsersName() []string
  GetUser() userValues
	GetWorklogs() WorklogData
  GetSummaryLog() SummaryLog
	InitService()
	createRequest(string, string, io.Reader) (*http.Request, error)
	formatWorklogsData(WorklogRes) error
	mapWorklogData([]WorklogsWorklog, map[int]FormattedWorklogData, *int, *int, *int)
  sortLogs([]Logs, time.Time, Logs) []Logs
}
