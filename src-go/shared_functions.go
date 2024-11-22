package main

import (
	"time"
)

const timeformat = "15.04" //This allows us to do string comparisons for times
func nowBetweenTimes(start time.Time, end time.Time) bool {
	if !start.IsZero() && !end.IsZero() {
		nowtime := time.Now().UTC().Format(timeformat)
		starttime := start.UTC().Format(timeformat)
		endtime := end.UTC().Format(timeformat)
		isvalid := false
		if starttime < endtime {
			//Time frame within same day (end hour larger than start hour)
			isvalid = (nowtime > starttime && nowtime < endtime)
		} else {
			//Time frame crosses into the next day
			// (end hour earlier than start hour)
			isvalid = (nowtime > starttime || nowtime < endtime)
		}
		return isvalid
	}
	return true
}

func nowValidWeekday(valid []string) bool {
	if len(valid) < 1 {
		return true //no restrictions
	}
	nowday := ""
	switch time.Now().Weekday() {
	case time.Sunday: nowday = "su"
	case time.Monday: nowday = "mo"
	case time.Tuesday: nowday = "tu"
	case time.Wednesday: nowday = "we"
	case time.Thursday: nowday = "th"
	case time.Friday: nowday = "fr"
	case time.Saturday: nowday = "sa"
	default: return false
	}
	isvalid := false
	for _, vd := range valid {
		if vd == nowday {
			isvalid = true
			break
		}
	}
	return isvalid
}