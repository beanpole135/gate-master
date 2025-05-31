package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func LoadAccountCodeFromForm(r *http.Request) (AccountCode, error) {
	// Parse the form
	r.ParseForm()
	acodeid := r.Form.Get("acodeid")
	code := r.Form.Get("code")
	codelength := parseFormInt(r.Form.Get("codelength"))
	label := r.Form.Get("label")
	is_active := r.Form.Get("isactive") == formChecked
	is_utility := r.Form.Get("isutility") == formChecked
	is_delivery := r.Form.Get("isdelivery") == formChecked
	is_contractor := r.Form.Get("iscontractor") == formChecked
	is_mail := r.Form.Get("ismail") == formChecked
	date_start := parseFormDate(r.Form.Get("dstart"))
	date_end := parseFormDate(r.Form.Get("dend"))
	time_start := parseFormTime(r.Form.Get("tstart"))
	time_end := parseFormTime(r.Form.Get("tend"))
	day_Su := r.Form.Get("d_sunday") == formChecked
	day_Mo := r.Form.Get("d_monday") == formChecked
	day_Tu := r.Form.Get("d_tuesday") == formChecked
	day_We := r.Form.Get("d_wednesday") == formChecked
	day_Th := r.Form.Get("d_thursday") == formChecked
	day_Fr := r.Form.Get("d_friday") == formChecked
	day_Sa := r.Form.Get("d_saturday") == formChecked

	AC := AccountCode{}
	AC.CodeLength = codelength
	if acodeid != "" {
		//Loading a pre-exising account code
		num, err := strconv.ParseInt(acodeid, 10, 64)
		if err != nil {
			return AC, err
		}
		list, err := DB.AccountCodeSelectAll(0, num)
		if err != nil || len(list) != 1 {
			return AC, fmt.Errorf("invalid account code ID")
		}
		AC = list[0]
	}
	// Validate the inputs
	if code == "" && AC.Code == "" && codelength == 0 {
		return AC, fmt.Errorf("missing PIN Code")
	}
	if code != "" && !validatePinCodeFormat(code) {
		return AC, fmt.Errorf("PIN code must be 4 or more numbers")
	}

	if label == "" && AC.Label == "" {
		return AC, fmt.Errorf("missing Description")
	}
	if time_start == nil {
		tn := time.Now()
		time_start = &tn
	}
	startPlus1year := time_start.AddDate(1, 0, 0)
	if time_end == nil || time_end.After(startPlus1year) || time_end.Before(*time_start) {
		time_end = &startPlus1year
	}

	// Populate the accountcode fields
	if code != "" && AC.Code == "" {
		//Can only set PIN code for new accountcodes (no changing later)
		AC.Code = code
	}
	if label != "" {
		AC.Label = label
	}
	AC.IsActive = is_active
	AC.IsUtility = is_utility
	AC.IsDelivery = is_delivery
	AC.IsContractor = is_contractor
	AC.IsMail = is_mail
	AC.ValidDays = []string{} //Reset and reload
	if day_Su {
		AC.ValidDays = append(AC.ValidDays, "su")
	}
	if day_Mo {
		AC.ValidDays = append(AC.ValidDays, "mo")
	}
	if day_Tu {
		AC.ValidDays = append(AC.ValidDays, "tu")
	}
	if day_We {
		AC.ValidDays = append(AC.ValidDays, "we")
	}
	if day_Th {
		AC.ValidDays = append(AC.ValidDays, "th")
	}
	if day_Fr {
		AC.ValidDays = append(AC.ValidDays, "fr")
	}
	if day_Sa {
		AC.ValidDays = append(AC.ValidDays, "sa")
	}
	if date_start != nil {
		AC.DateStart = *date_start
	} else {
		AC.DateStart = time.Time{}
	}
	if date_end != nil {
		AC.DateEnd = *date_end
	} else {
		AC.DateStart = time.Time{}
	}
	if time_start != nil {
		AC.TimeStart = *time_start
	} else {
		AC.DateStart = time.Time{}
	}
	if time_end != nil {
		AC.TimeEnd = *time_end
	} else {
		AC.DateStart = time.Time{}
	}
	return AC, nil
}
