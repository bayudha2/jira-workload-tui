package services

import (
	"io"
	"net/http"
	"time"
)

type ServiceType interface {
	FetchMembers()
	FetchUsers()
	FetchIssues(FetchWorklogPayload)
	FetchWorklogs(string) *WorklogField
	GetUsersName() []string
  GetUser() userValues
	GetWorklogs() WorklogData
  GetSummaryLog() SummaryLog
	InitService()
	createRequest(string, string, io.Reader) (*http.Request, error)
	formatWorklogsData(WorklogRes)
	mapWorklogData([]WorklogsWorklog, map[int]FormattedWorklogData, *int, *int, *int)
  sortLogs([]Logs, time.Time, Logs) []Logs
}
