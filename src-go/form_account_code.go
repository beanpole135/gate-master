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
	label := r.Form.Get("label")
	is_active := r.Form.Get("isactive") == formChecked
	is_utility := r.Form.Get("isutility") == formChecked
	is_delivery := r.Form.Get("isdelivery") == formChecked
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
	if acodeid != "" {
		//Loading a pre-exising account code
		num, err := strconv.ParseInt(acodeid, 10, 64)
		if err != nil {
			return AC, err
		}
		list, err := DB.AccountCodeSelectAll(0, num)
		if err != nil || len(list) != 1 {
			return AC, fmt.Errorf("Invalid account code ID")
		}
		AC = list[0]
	}
	// Validate the inputs
	if code == "" && AC.Code == "" {
		return AC, fmt.Errorf("Missing PIN Code")
	}
	if code != "" && !validatePinCodeFormat(code) {
		return AC, fmt.Errorf("PIN code must be 4 or more numbers")
	}
	if label == "" && AC.Label == "" {
		return AC, fmt.Errorf("Missing Description")
	}

	if (time_start != nil && time_end == nil) || (time_start == nil && time_end != nil) {
		return AC, fmt.Errorf("Both start and end times must be provided, or neither of them")
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
