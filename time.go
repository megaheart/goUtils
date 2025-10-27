package goUtils

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type DateTime time.Time

func (d DateTime) String() string {
	return time.Time(d).Format("2006-01-02 15:04:05.999999999")
}

var const_DATETIME_FORMATS = []string{
	"2006-01-02",
	"02/01/2006",
	"2006-01-02 15:04",
	"15:04 02/01/2006",
	"2006-01-02 15:04:05",
	"15:04:05 02/01/2006",
	"2006-01-02 15:04:05.999999999",
	"15:04:05.999999999 02/01/2006",
	"2006-1-2 15:4",
	"15:4 2/1/2006",
	"2006-1-2 15:4:5",
	"15:4:5 2/1/2006",
	"2006-1-2 15:4:5.999999999",
	"15:4:5.999999999 2/1/2006",
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05.999999999Z07:00",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05.999999999Z",
	"2006-01-02T15:04Z07:00",
	"2006-01-02T15:04Z",
}

func (d *DateTime) UnmarshalJSON(data []byte) error {
	s := string(data)
	// Remove quotes from the string
	s = strings.Trim(s, "\"")

	for _, format := range const_DATETIME_FORMATS {
		parsedTime, err := time.Parse(format, s)
		if err == nil {
			*d = DateTime(parsedTime)
			return nil
		}
	}

	return errors.New("Invalid DateTime format")
}

type Date time.Time

func (d Date) String() string {
	return time.Time(d).Format("2006-01-02")
}

var const_DATE_FORMATS = []string{
	"2006-01-02",
	"02/01/2006",
	"06-01-02",
	"02/01/06",
	"2006-1-2",
	"2/1/2006",
	"06-1-2",
	"2/1/06",
}

func (d *Date) UnmarshalJSON(data []byte) error {
	s := string(data)
	// Remove quotes from the string
	s = strings.Trim(s, "\"")

	for _, format := range const_DATE_FORMATS {
		parsedTime, err := time.Parse(format, s)
		if err == nil {
			*d = Date(parsedTime)
			return nil
		}
	}

	return errors.New("Invalid Date format")
}

type TimeInDay time.Duration

func (t TimeInDay) String() string {
	return time.Time{}.Add(time.Duration(t)).Format("15:04:05.999999999")
}

var const_TIME_FORMATS = []string{
	"15:04:05.999999999",
	"15:04:05",
	"15:04",
	"15:4:5.999999999",
	"15:4:5",
	"15:4",
}

func (t *TimeInDay) UnmarshalJSON(data []byte) error {
	s := string(data)
	// Remove quotes from the string
	s = strings.Trim(s, "\"")

	for _, format := range const_TIME_FORMATS {
		parsedTime, err := time.Parse(format, s)
		if err == nil {
			*t = TimeInDay(parsedTime.Sub(time.Time{}))
			return nil
		}
	}

	return errors.New("Invalid TimeInDay format")
}

func GetTimeInDay(t time.Time) TimeInDay {
	return TimeInDay(t.Sub(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())))
}

type Weekday time.Weekday

func (wd Weekday) String() string {
	return time.Weekday(wd).String()
}

func (wd *Weekday) UnmarshalJSON(data []byte) error {
	var weekdayStr string = string(data)
	// Remove quotes from the string
	weekdayStr = strings.Trim(weekdayStr, "\"")

	// Convert the string to a time.Weekday value
	switch weekdayStr {
	case "Sunday":
	case "Sun":
	case "sunday":
	case "sun":
		*wd = Weekday(time.Sunday)
	case "Monday":
	case "Mon":
	case "monday":
	case "mon":
		*wd = Weekday(time.Monday)
	case "Tuesday":
	case "Tue":
	case "tuesday":
	case "tue":
		*wd = Weekday(time.Tuesday)
	case "Wednesday":
	case "Wed":
	case "wednesday":
	case "wed":
		*wd = Weekday(time.Wednesday)
	case "Thursday":
	case "Thu":
	case "thursday":
	case "thu":
		*wd = Weekday(time.Thursday)
	case "Friday":
	case "Fri":
	case "friday":
	case "fri":
		*wd = Weekday(time.Friday)
	case "Saturday":
	case "Sat":
	case "saturday":
	case "sat":
		*wd = Weekday(time.Saturday)
	default:
		return fmt.Errorf("invalid weekday: %s", weekdayStr)
	}
	return nil
}
