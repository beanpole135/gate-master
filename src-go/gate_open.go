package main

import (
	"fmt"
	"time"
)

func CheckPINAndOpen(pin string) error {
	ac, err := DB.AccountCodeMatch(pin)
	if err != nil {
		return err
	}
	// if ac==nil, invalid PIN
	err = OpenGateAndNotify(nil, ac)
	if ac == nil || err != nil {
		return fmt.Errorf("Invalid PIN")
	}
	return nil
}

func OpenGateAndNotify(acct *Account, code *AccountCode) error {
	// Now determine who to notify and send out notices
	var emails []string
	msg := "%s is entering the neighborhood"
	subject := fmt.Sprintf("%s Gate Notification", CONFIG.SiteName)
	var gl GateLog
	gl.TimeOpened = time.Now()
	gl.OpenedName = "unknown"
	if code != nil && code.IsValid() {
		msg = fmt.Sprintf(msg, code.Label)
		var contacts []Contact
		var err error
		//Gate PIN used - notify everybody associated
		if !code.IsUtility && !code.IsDelivery {
			// Just notify account holder
			contacts, err = DB.ContactsForAccountNotify(code.AccountID)
		} else {
			// Notify everyone listed for these types of entries
			contacts, err = DB.ContactsForAllNotify(code.IsUtility, code.IsDelivery, code.IsContractor, code.IsMail)
		}
		if err != nil {
			fmt.Println("Error reading Contacts:", err)
		}
		for _, c := range contacts {
			// Assemble the list of people to contacts
			emails = append(emails, c.ContactEmail())
		}
		// Assemble the gate log
		gl.AccountID = code.AccountID
		gl.OpenedName = code.Label
		gl.UsedWeb = false
		gl.UsedCode = code.Code
		gl.Success = true

	} else if acct != nil {
		msg = fmt.Sprintf(msg, fmt.Sprintf("%s %s has opened the gate.", acct.FirstName, acct.LastName))
		//Web Portal used - no need to notify anyone?
		gl.AccountID = acct.AccountID
		gl.OpenedName = fmt.Sprintf("%s, %s", acct.LastName, acct.FirstName)
		gl.UsedWeb = true
		gl.Success = true
	} else {
		// Unknown who is opening the gate
		gl.Success = false
		gl.UsedWeb = false //if web is used, never get a failure/invalid
	}
	// Snap a picture from the gate
	gl.GatePicture = CAM.TakePicture()

	// Open the Gate
	if gl.Success {
		fmt.Println("Opening Gate!!")
		CONFIG.Gate.OpenGate()
		CONFIG.Keypad.DisplayOnLCD("Welcome!", 2)
	}

	// Record the gate log
	_, err := DB.GateLogInsert(&gl)
	if err != nil {
		fmt.Println("Error inserting GateLog:", err)
	}
	go SaveCSVLog(gl, CONFIG.LogsDir)

	if !gl.Success {
		return fmt.Errorf("unknown gate open request - denied")
	}
	// Now send all the notification emails for successes
	msg = fmt.Sprintf("[%s] ", time.Now().Format("Jan _2: 03:04 MST")) + msg
	for _, to := range emails {
		CONFIG.Email.SendEmail(to, subject, msg, false)
	}
	return nil
}
