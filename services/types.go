package services

// members
type pageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type resultMember []struct {
	AccountId string `json:"accountId"`
}

type TeamMemberRes struct {
	PageInfo pageInfo     `json:"pageInfo"`
	Results  resultMember `json:"results"`
}

// users
type userValues struct {
	Self        string `json:"self"`
	AccountId   string `json:"accountId"`
	AccountType string `json:"accountType"`
	DisplayName string `json:"displayName"`
	EmailAdrres string `json:"emailAddress"`
	Active      bool   `json:"active"`
	TimeZone    string `json:"timeZone"`
}

type UserRes struct {
	Self       string       `json:"self"`
	MaxResults int          `json:"maxResults"`
	StartAt    int          `json:"startAt"`
	Total      int          `json:"total"`
	IsLast     bool         `json:"isLast"`
	Values     []userValues `json:"values"`
}

// worklogs

type WorklogsWorklog struct {
	Self             string `json:"self"`
	Comment          string `json:"comment"`
	Created          string `json:"created"`
	Updated          string `json:"updated"`
	Started          string `json:"started"`
	TimeSpent        string `json:"timeSpent"`
	TimeSpentSeconds int    `json:"timeSpentSeconds"`
	Id               string `json:"id"`
	IssueId          string `json:"issueId"`
}

type WorklogField struct {
	StartAt    int               `json:"startAt"`
	MaxResults int               `json:"maxResults"`
	Total      int               `json:"total"`
	Worklogs   []WorklogsWorklog `json:"worklogs"`
}

type FieldIssue struct {
	Worklog WorklogField `json:"worklog"`
}

type IssuesWorklog struct {
	Expand string     `json:"expand"`
	Id     string     `json:"id"`
	Self   string     `json:"self"`
	Key    string     `json:"key"`
	Fields FieldIssue `json:"fields"`
}

type WorklogRes struct {
	Expand     string          `json:"expand"`
	StartAt    int             `json:"startAt"`
	MaxResults int             `json:"maxResults"`
	Total      int             `json:"total"`
	Issues     []IssuesWorklog `json:"issues"`
}

type Logs struct {
	TimeRange string
	Comment   string
}

type FormattedWorklogData struct {
	TimeSpent int
	Logs      []Logs
}
