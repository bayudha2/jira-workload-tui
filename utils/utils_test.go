package utils

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStrToPtr(t *testing.T) {
	str := "test"
	res := StrToPtr(str)
	require.IsType(t, &str, res)
}

func TestItoPtr(t *testing.T) {
	num := 1
	res := ItoPtr(num)
	require.IsType(t, &num, res)
}

func TestFilterStrings(t *testing.T) {
	errTest := []string{"example 1", "example 2", "example 3", "example 4", "example 21"}
	expected := []string{"example 2", "example 21"}
	find := "2"

	res := FilterStrings(errTest, find)

	require.Equal(t, expected, res)
}

func TestCalculateRGB(t *testing.T) {
	var res string
	rgb := [3]int{85, 255, 255}
	expect := "[38;2;85;255;255m█[0m"

	CalculateRGB(&res, '█', 1, rgb, rgb)
	format := fmt.Sprintf("%s\033[0m", res)

	require.Len(t, strings.Split(format, ""), len(expect))
}

func TestCalculateRangeDateInMonth(t *testing.T) {
	tcs := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "calculate normal",
			test: func(t *testing.T) {
				month := 1
				year := 2020
				expectFrom := "2020-01-01"
				expectTo := "2020-01-31"

				from, to := CalculateRangeDateInMonth(month, year)

				require.Equal(t, expectFrom, from)
				require.Equal(t, expectTo, to)
			},
		},
		{
			name: "calculate with montth date is more then one digit",
			test: func(t *testing.T) {
				month := 12
				year := 2020
				expectFrom := "2020-12-01"
				expectTo := "2020-12-31"

				from, to := CalculateRangeDateInMonth(month, year)

				require.Equal(t, expectFrom, from)
				require.Equal(t, expectTo, to)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func TestFormatSecondToHourMinute(t *testing.T) {
	tcs := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "format second to hour minute",
			test: func(t *testing.T) {
				second := 5400
				expect := "1h 30m"
				res := FormatSecondToHourMinute(second, false)

				require.Equal(t, expect, res)
			},
		},
		{
			name: "with long format",
			test: func(t *testing.T) {
				second := 41400
				expect := "11 hours 30 minutes"
				res := FormatSecondToHourMinute(second, true)

				require.Equal(t, expect, res)
			},
		},
		{
			name: "with only minutes",
			test: func(t *testing.T) {
				second := 1800
				expect := "30 minutes"
				res := FormatSecondToHourMinute(second, true)

				require.Equal(t, expect, res)
			},
		},
		{
			name: "with only hours",
			test: func(t *testing.T) {
				second := 7200
				expect := "2 hours"
				res := FormatSecondToHourMinute(second, true)

				require.Equal(t, expect, res)
			},
		},
		{
			name: "with long format with correct suffix",
			test: func(t *testing.T) {
				second := 3660
				expect := "1 hour 1 minute"
				res := FormatSecondToHourMinute(second, true)

				require.Equal(t, expect, res)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func TestFormatCommentDesc(t *testing.T) {
	tcs := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "default format",
			test: func(t *testing.T) {
				dummyText := `
          Lorem Ipsum is simply dummy text of the printing and typesetting industry.
          Lorem Ipsum has been the industry's standard dummy text ever since the 1500s,
          when an unknown printer took a galley of type and scrambled it to make a type specimen
          `

				expect := []string{
					"Lorem Ipsum is simply dummy text of the printing and typeset",
					"ting industry. Lorem Ipsum has been the industry's standard ",
					"dummy text ever since the 1500s, when an unknown printer too",
					"k a galley of type and scrambled it to make a type specimen",
				}

				res := FormatCommentDesc(dummyText, 60)
				require.Equal(t, expect, res)
			},
		},
		{
			name: "format but text is too long",
			test: func(t *testing.T) {
				dummyText := `
          Lorem Ipsum is simply dummy text of the printing and typesetting industry.
          Lorem Ipsum has been the industry's standard dummy text ever since the 1500s,
          when an unknown printer took a galley of type and scrambled it to make a type specimen
          `

				expect := []string{"Lorem Ipsu", "m is simpl", "y dummy te", "xt of the "}

				res := FormatCommentDesc(dummyText, 10)
				require.Equal(t, expect, res)
			},
		},
		{
			name: "format, text is short",
			test: func(t *testing.T) {
				dummyText := "Lorem Ipsum is"
				expect := []string{"Lorem Ipsum is"}

				res := FormatCommentDesc(dummyText, 100)
				require.Equal(t, expect, res)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func TestValidateENVs(t *testing.T) {
	tcs := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "got no empty map",
			test: func(t *testing.T) {
				dummy := map[string]string{
					"test_1": "exist",
					"test_2": "exist",
					"test_3": "exist",
					"test_4": "exist",
					"test_5": "exist",
				}

				err := ValidateENVs(dummy)
				require.NoError(t, err)
			},
		},
		{
			name: "got empty map",
			test: func(t *testing.T) {
				dummy := map[string]string{
					"test_1": "exist",
					"test_2": "exist",
					"test_3": "",
					"test_4": "exist",
					"test_5": "exist",
				}

				err := ValidateENVs(dummy)
				require.Error(t, err)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func TestGetWorkDays(t *testing.T) {
	tcs := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "default behaviour",
			test: func(t *testing.T) {
				month := 1
				year := 2024

				expectTargetMonth := 662400
				expectTargetToday := 662400

				targetMonth, targetToday := GetWorkDays(month, year)
				require.Equal(t, expectTargetMonth, targetMonth)
				require.Equal(t, expectTargetToday, targetToday)
			},
		},
		{
			name: "error formatting month year int not valid",
			test: func(t *testing.T) {
				month := -1
				year := 2024

				expectTargetMonth := 0
				expectTargetToday := 0

				targetMonth, targetToday := GetWorkDays(month, year)
				require.Equal(t, expectTargetMonth, targetMonth)
				require.Equal(t, expectTargetToday, targetToday)
			},
		},
		{
			name: "get equal current month and year",
			test: func(t *testing.T) {
				now := time.Now()
				month := int(now.Month())
				year := now.Year()

				expectTargetMonth := 604800
				expectTargetToday := 288000

				targetMonth, targetToday := GetWorkDays(month, year)
				require.Equal(t, expectTargetMonth, targetMonth)
				require.Equal(t, expectTargetToday, targetToday)
			},
		},
		{
			name: "get more then current month and year",
			test: func(t *testing.T) {
				nowAfter := time.Now().AddDate(1, 1, 0)
				month := int(nowAfter.Month())
				year := nowAfter.Year()

				expectTargetMonth := 662400
				expectTargetToday := 0

				targetMonth, targetToday := GetWorkDays(month, year)
				require.Equal(t, expectTargetMonth, targetMonth)
				require.Equal(t, expectTargetToday, targetToday)
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}

func TestGetWeekdays(t *testing.T) {
	now := time.Now()
	startDate := now.AddDate(1, 0, 0)
	endDate := now

	res := getWeekdays(startDate, endDate)
	require.Equal(t, 0, res)
}
