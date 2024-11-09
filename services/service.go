package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
	"tui/config"
	"tui/utils"

	termhandler "tui/term-handler"
)

type FetchWorklogPayload struct {
	Name  string
	Year  int
	Month int
}

type WorklogData struct {
	LastDate int
	Name     string
	Month    int
	Year     int
	Data     map[int]FormattedWorklogData
}

type SummaryLog struct {
	TotalBacklog   int
	TotalWorklog   int
	TotalTimeSpent int
}

type ServiceApp struct {
	wg         *sync.WaitGroup
	localWg    *sync.WaitGroup
	mutex      *sync.Mutex
	handler    termhandler.TermhandlerType
	config     config.JiraConfigType
	client     *http.Client
	accoundIds resultMember
	users      []userValues
	worklogs   WorklogData
	summaryLog SummaryLog
}

func NewService(
	wg *sync.WaitGroup,
	mutex *sync.Mutex,
	handler *termhandler.TermhandlerType,
	config *config.JiraConfigType,
) ServiceType {
	return &ServiceApp{
		wg:         wg,
		mutex:      mutex,
		localWg:    new(sync.WaitGroup),
		handler:    *handler,
		config:     *config,
		client:     &http.Client{},
		accoundIds: resultMember{},
		users:      []userValues{},
		worklogs:   WorklogData{},
		summaryLog: SummaryLog{},
	}
}

func (s *ServiceApp) createRequest(
	method string,
	url string,
	body io.Reader,
) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	emailENV := s.config.GetEmail()
	userTokenENV := s.config.GetUserToken()

	basicA64 := fmt.Sprintf("%s:%s", emailENV, userTokenENV)
	basicAuth := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(basicA64)))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", basicAuth)
	return req, nil
}

// FetchMembers implements ServiceType.
func (s *ServiceApp) FetchMembers() {
	baseURI := s.config.GetAtlassianURL()
	teamId := s.config.GetTeamID()
	orgId := s.config.GetOrgID()

	urlFetchMember := fmt.Sprintf(
		"%s/gateway/api/public/teams/v1/org/%s/teams/%s/members",
		baseURI,
		orgId,
		teamId,
	)

	req, err := s.createRequest(http.MethodPost, urlFetchMember, nil)
	if err != nil {
		// TODO: handle when fail creating request
		log.Printf("error creating new req: %v", err)
		return
	}

	resp, err := s.client.Do(req)
	if err != nil {
		// TODO: handle when fail fetching members
		log.Printf("error fetching member team: %v", err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var resBody TeamMemberRes
	decoder.Decode(&resBody)

	s.accoundIds = resBody.Results
}

func (s *ServiceApp) FetchUsers() {
	baseURI := s.config.GetAtlassianURL()
	params := ""
	for _, ids := range s.accoundIds {
		params += fmt.Sprintf("accountId=%s&", ids.AccountId)
	}

	urlGetUsers := fmt.Sprintf(
		"%s/rest/api/2/user/bulk?maxResults=50&%s",
		baseURI,
		params,
	)
	req, err := s.createRequest(http.MethodGet, urlGetUsers, nil)
	if err != nil {
		// TODO: handle when fail creating request
		log.Printf("error creating req: %v", err)
		return
	}

	resp, err := s.client.Do(req)
	if err != nil {
		// TODO: handle when fail fetching users
		log.Printf("error getting users: %v", err)
		return
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var bodyRes UserRes
	decoder.Decode(&bodyRes)

	s.users = bodyRes.Values
}

// GetUsersData implements ServiceType.
func (s *ServiceApp) GetUsersName() []string {
	res := []string{}
	for _, user := range s.users {
		res = append(res, user.DisplayName)
	}
	return res
}

func (s *ServiceApp) GetUser() userValues {
	var user userValues
	for _, val := range s.users {
		if val.DisplayName == s.worklogs.Name {
			user = val
			break
		}
	}

	return user
}

func (s *ServiceApp) GetWorklogs() WorklogData {
	return s.worklogs
}

func (s *ServiceApp) GetSummaryLog() SummaryLog {
	return s.summaryLog
}

func (s *ServiceApp) FetchIssues(param FetchWorklogPayload) {
	baseURI := s.config.GetAtlassianURL()
	project := s.config.GetJiraProject()

	var user userValues
	url := fmt.Sprintf("%s/rest/api/2/search", baseURI)

	fromDate, toDate := utils.CalculateRangeDateInMonth(param.Month, param.Year)
	getSpesificUser(s.users, &user, param.Name)

	payload := fmt.Sprintf(`{
    "jql": "project IN (%s) AND assignee = %s AND worklogDate >= %s AND worklogDate <= %s ORDER BY created DESC",
    "fields": ["worklog"]
    }`, project, user.AccountId, fromDate, toDate)

	payloadReader := bytes.NewReader([]byte(payload))
	req, err := s.createRequest(http.MethodPost, url, payloadReader)
	if err != nil {
		// TODO: handle when fail creating request
		log.Printf("error creating req: %v", err)
		return
	}

	res, err := s.client.Do(req)
	if err != nil {
		// TODO: handle when fail Fetching worklog
		log.Printf("error fetching issues: %v", err)
		return
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var resBody WorklogRes
	decoder.Decode(&resBody)
	s.worklogs.Month = param.Month
	s.worklogs.Year = param.Year
	s.worklogs.Name = param.Name
	s.formatWorklogsData(resBody)
}

func (s *ServiceApp) FetchWorklogs(id string) *WorklogField {
	baseURI := s.config.GetAtlassianURL()
	url := fmt.Sprintf("%s/rest/api/2/issue/%s/worklog", baseURI, id)
	req, err := s.createRequest(http.MethodGet, url, nil)
	if err != nil {
		// TODO: handle when fail creating request
		log.Printf("error creating req: %v", err)
		return nil
	}

	res, err := s.client.Do(req)
	if err != nil {
		// TODO: handle when fail Fetching worklog
		log.Printf("error fetching worklog: %v", err)
		return nil
	}
	defer res.Body.Close()
	var resBody WorklogField
	json.NewDecoder(res.Body).Decode(&resBody)
	return &resBody
}

func getSpesificUser(items []userValues, item *userValues, name string) {
	for _, user := range items {
		if user.DisplayName == name {
			*item = user
			break
		}
	}
}

func (s *ServiceApp) formatWorklogsData(worklogData WorklogRes) {
	wkData := map[int]FormattedWorklogData{}
	lastDate := 0
	totalTimeSpent := 0
	totalWorklog := 0

	for _, issue := range worklogData.Issues {
		s.localWg.Add(1)
		go func(issueItem IssuesWorklog) {
			if issueItem.Fields.Worklog.Total > 20 {
				wlField := s.FetchWorklogs(issueItem.Id)
				s.mapWorklogData(
					wlField.Worklogs,
					wkData,
					&lastDate,
					&totalTimeSpent,
					&totalWorklog,
				)
			} else {
				s.mapWorklogData(
					issueItem.Fields.Worklog.Worklogs,
					wkData,
					&lastDate,
					&totalTimeSpent,
					&totalWorklog,
				)
			}
		}(issue)
	}

	s.localWg.Wait()

	s.summaryLog = SummaryLog{
		TotalBacklog:   worklogData.Total,
		TotalWorklog:   totalWorklog,
		TotalTimeSpent: totalTimeSpent,
	}

	s.worklogs = WorklogData{
		Month:    s.worklogs.Month,
		Year:     s.worklogs.Year,
		Name:     s.worklogs.Name,
		LastDate: lastDate,
		Data:     wkData,
	}
}

func (s *ServiceApp) mapWorklogData(
	arr []WorklogsWorklog,
	wkData map[int]FormattedWorklogData,
	lastDate *int,
	totalTimeSpent *int,
	totalWorklog *int,
) {
	iso8601Layout := "2006-01-02T15:04:05-0700"
	hhMmLayout := "15:04"

	for _, worklog := range arr {
		parsed, _ := time.Parse(iso8601Layout, worklog.Started)
		if int(parsed.Month()) != s.worklogs.Month {
			continue
		}

		s.mutex.Lock()
		day := parsed.Day()
		if day > *lastDate {
			*lastDate = day
		}
		s.mutex.Unlock()

		timeSpent := worklog.TimeSpentSeconds

		s.mutex.Lock()
		*totalWorklog += 1
		*totalTimeSpent += timeSpent
		s.mutex.Unlock()

		startTime := fmt.Sprintf("%s", parsed.Format(hhMmLayout))
		endTime := fmt.Sprintf(
			"%s",
			parsed.Add(time.Duration(timeSpent)*time.Second).Format(hhMmLayout),
		)
		timeRange := fmt.Sprintf("%s - %s", startTime, endTime)

		s.mutex.Lock()
		wkLogs := wkData[day].Logs
		logs := []Logs{}

		if len(wkLogs) > 0 {
			logs = s.sortLogs(wkLogs, parsed, Logs{
				Comment:   worklog.Comment,
				TimeRange: timeRange,
				Started:   parsed,
			})
		} else {
			logs = append(wkLogs, Logs{
				Comment:   worklog.Comment,
				TimeRange: timeRange,
				Started:   parsed,
			})
		}

		wkData[day] = FormattedWorklogData{
			TimeSpent: wkData[day].TimeSpent + timeSpent,
			Logs:      logs,
		}
		s.mutex.Unlock()
	}
	s.localWg.Done()
}

// Does place new item in exact location sequentially according to date clock
func (s *ServiceApp) sortLogs(arrLogs []Logs, current time.Time, newItem Logs) []Logs {
	var iPrevP *int
	res := []Logs{}

	for i := 0; i < len(arrLogs); i++ {
		if arrLogs[len(arrLogs)-i-1].Started.After(current) {
			iPrevP = &i
		}
	}

	if iPrevP != nil {
		arrLogsCopy := make([]Logs, len(arrLogs))
		copy(arrLogsCopy, arrLogs)

		arrLogsL := append(arrLogsCopy[:len(arrLogsCopy)-*iPrevP-1], newItem)
		res = append(arrLogsL, arrLogs[len(arrLogs)-*iPrevP-1:]...)
	} else {
		res = append(arrLogs, newItem)
	}

	return res
}

// InitService implements ServiceType.
func (s *ServiceApp) InitService() {
	s.handler.MoveCursor(termhandler.Position{2, 2})
	s.handler.Draw("Loading... please kindly wait...")
	s.handler.Render()

	s.FetchMembers()
	s.FetchUsers()
}
