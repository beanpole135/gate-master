package main

import (
	"fmt"
)

func OpenGateAndNotify(acct *Account, code *AccountCode) error {
	// Now determine who to notify and send out notices
	var emails []string
	msg := "%s is entering the neighborhood"
	subject := fmt.Sprintf("%s Gate Notification", CONFIG.SiteName)
	var gl GateLog
	if code != nil {
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

	} else if acct != nil {
		msg = fmt.Sprintf(msg, fmt.Sprintf("%s %s has opened the gate for somebody", acct.FirstName, acct.LastName))
		//Web Portal used - no need to notify anyone?
		gl.AccountID = acct.AccountID
		gl.OpenedName = fmt.Sprintf("%s, %s", acct.LastName, acct.FirstName)
		gl.UsedWeb = true
	} else {
		// Unknown who is opening the gate
		return fmt.Errorf("Unknown Gate Open Request - denied")
	}
	// Snap a picture from the gate
	gl.GatePicture = CAM.TakePicture()

	// Open the Gate

	// Record the gate log
	_, err := DB.GateLogInsert(&gl)
	if err != nil {
		fmt.Println("Error inserting GateLog:", err)
	}

	// Now send all the notification emails
	for _, to := range emails {
		CONFIG.Email.SendEmail(to, subject, msg, false)
	}
	return nil
}
