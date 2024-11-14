package utils

import (
	"fmt"
	"log"
	"strings"
	"time"
)

var WORKING_HOURS = 8

func StrToPtr(str string) *string {
	return &str
}

func ItoPtr(x int) *int {
	return &x
}

func CalculateRGB(str *string, char rune, length int, from [3]int, to [3]int) {
	valueR := from[0]
	valueG := from[1]
	valueB := from[2]

	deviationDistanceR := (to[0] - from[0]) / length
	deviationDistanceG := (to[1] - from[1]) / length
	deviationDistanceB := (to[2] - from[2]) / length

	result := ""
	for i := 0; i < length; i++ {
		result += fmt.Sprintf("\033[38;2;%d;%d;%dm%s", valueR, valueG, valueB, string(char))
		valueR += deviationDistanceR
		valueG += deviationDistanceG
		valueB += deviationDistanceB
	}

	*str = result
}

func FilterStrings(arr []string, str string) []string {
	res := []string{}
	for _, item := range arr {
		toLowerItem := strings.ToLower(item)
		toLowerStr := strings.ToLower(str)

		if isContain := strings.Contains(toLowerItem, toLowerStr); isContain {
			res = append(res, item)
		}
	}

	return res
}

func CalculateRangeDateInMonth(month int, year int) (string, string) {
	formatMonth := func(m int) string {
		if m < 10 {
			return fmt.Sprintf("0%d", m)
		}

		return fmt.Sprintf("%d", m)
	}(month)

	lastDay := getlastDateOfMonth(month, year)

	fromDate := fmt.Sprintf("%d-%s-01", year, formatMonth)
	toDate := fmt.Sprintf("%d-%s-%d", year, formatMonth, lastDay)

	return fromDate, toDate
}

func FormatSecondToHourMinute(seconds int, isLongFormat bool) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	fHours := "h"
	fMinutes := "m"
	gap := ""

	if isLongFormat {
		switch {
		case hours > 1:
			fHours = "hours"
		default:
			fHours = "hour"
		}

		switch {
		case minutes > 1:
			fMinutes = "minutes"
		default:
			fMinutes = "minute"
		}
		gap = " "
	}

	sHours := fmt.Sprintf("%d%s%s", hours, gap, fHours)
	sMinutes := fmt.Sprintf("%d%s%s", minutes, gap, fMinutes)

	if hours > 0 && minutes > 0 {
		return fmt.Sprintf("%s %s", sHours, sMinutes)
	} else if hours > 0 {
		return sHours
	}

	return sMinutes
}

func FormatCommentDesc(s string, n int) []string {
	normalizedT := strings.ReplaceAll(s, "\n", " ")
	normalizedT = strings.Join(strings.Fields(normalizedT), " ")
	if len(normalizedT) > n*4 {
		normalizedT = normalizedT[0 : n*4]
	}

	if n < 0 || len(normalizedT) < n {
		return []string{normalizedT}
	}

	var result []string
	for i := 0; i < len(normalizedT); i += n {
		end := i + n
		if end > len(normalizedT) {
			end = len(normalizedT)
		}

		result = append(result, normalizedT[i:end])
	}
	return result
}

func GetWorkDays(month int, year int) (int, int) {
	tMonth := 0
	tToday := 0

	if month <= 0 || year <= 0 {
		log.Printf("error month or/and year is not valid number")
		return tMonth, tToday
	}

	now := time.Now()
	startDate, lastDate := CalculateRangeDateInMonth(month, year)
	parsedStartDate, _ := time.Parse("2006-01-02", startDate)
	parsedLastDate, _ := time.Parse("2006-01-02", lastDate)

	weekDayCountWhole := getWeekdays(parsedStartDate, parsedLastDate)
	if month < int(now.Month()) && year <= now.Year() { // given month is behind of current month
		tMonth = weekDayCountWhole * WORKING_HOURS * 60 * 60
		tToday = tMonth
	} else if month == int(now.Month()) && year == now.Year() { // given date is same with current date
		weekDayCountUntilN := getWeekdays(time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()), now)

		tMonth = weekDayCountWhole * WORKING_HOURS * 60 * 60
		tToday = weekDayCountUntilN * WORKING_HOURS * 60 * 60
	} else { // given month is after the current month (or future)
		tMonth = weekDayCountWhole * WORKING_HOURS * 60 * 60
	}

	return tMonth, tToday
}

func getWeekdays(startDate, endDate time.Time) int {
	if startDate.After(endDate) {
		return 0
	}

	count := 0
	for current := startDate; !current.After(endDate); current = current.AddDate(0, 0, 1) {
		weekday := current.Weekday()
		if weekday != time.Saturday && weekday != time.Sunday {
			count++
		}
	}
	return count
}

func getlastDateOfMonth(month int, year int) int {
	nextMonth := month + 1
	if nextMonth > 12 {
		nextMonth = 1
	}

	firstOfNextMonth := time.Date(year, time.Month(nextMonth), 1, 0, 0, 0, 0, time.Now().Location())
	return firstOfNextMonth.AddDate(0, 0, -1).Day()
}

func ValidateENVs(envs map[string]string) error {
	for key, val := range envs {
		if val == "" {
			return fmt.Errorf("error validating: %s", key)
		}
	}

	return nil
}
