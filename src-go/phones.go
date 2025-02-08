package main

import (
	"fmt"
)

func isValidCellType(typ string) bool {
	switch typ {
	case "att", "tmobile", "verizon", "sprint", "uscell", "boost", "cricket", "googlefi", "metropcs", "virgin":
		return true
	default:
		return false
	}
}

func sanitizePhone(phone string) (string, error) {
	//Must be reduced to a 10-digit number-only string
	var nphone string
	for _, ch := range phone {
		switch ch {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			nphone += string(ch)
		}
	}
	if len(nphone) != 10 {
		return "", fmt.Errorf("Invalid phone number (must be 10 digits)")
	}
	return nphone, nil
}

func DisplayPhone(phone string) string {
	if len(phone) != 10 {
		return phone // can't reformat
	}
	return fmt.Sprintf("(%s) %s-%s", phone[:3], phone[3:6], phone[6:])
}

func phoneToEmail(phone string, typ string) (string, error) {
	phn, err := sanitizePhone(phone)
	if err != nil {
		return "", err
	}
	eml := ""
	switch typ {
	case "att":
		eml = "%s@txt.att.net" //AT&T
	case "tmobile":
		eml = "%s@tmomail.net" //T-Mobile
	case "verizon":
		eml = "%s@vtext.com" //Verizon or XFinity
	case "sprint":
		eml = "%s@messaging.sprintpcs.com" //Sprint
	case "uscell":
		eml = "%s@email.uscc.net" //US Cellular
	case "boost":
		eml = "%s@sms.myboostmobile.com" //Boost Mobile
	case "cricket":
		eml = "%s@sms.cricketwireless.net" //Cricket Wireless
	case "googlefi":
		eml = "%s@msg.fi.google.com" //Google Fi
	case "metropcs":
		eml = "%s@mymetropcs.com" //MetroPCS
	case "virgin":
		eml = "%s@vmobl.com" //Virgin Mobile
	default:
		return "", fmt.Errorf("Invalid phone type")
	}
	return fmt.Sprintf(eml, phn), nil
}
