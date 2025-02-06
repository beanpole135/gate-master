package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func LoadContactFromForm(r *http.Request) (Contact, error) {
	// Parse the form
	r.ParseForm()
	contactid := r.Form.Get("contactid")
	email := r.Form.Get("email")
	celltype := r.Form.Get("celltype")
	is_primary := r.Form.Get("isprimary") == formChecked
	is_active := r.Form.Get("isactive") == formChecked
	is_utility := r.Form.Get("isutility") == formChecked
	is_delivery := r.Form.Get("isdelivery") == formChecked

	AC := Contact{}
	var err error
	if celltype == "email"{
		AC.Email = email
		AC.CellType = ""
	} else if isValidCellType(celltype) {
		AC.CellType = celltype
		AC.PhoneNum, err = sanitizePhone(email)
		if err != nil { 
			return AC, err
		}
	} else {
		return AC, fmt.Errorf("Invalid Contact Type")
	}
	if contactid != "" {
		cid, err := strconv.ParseInt(contactid, 10, 32)
		if err != nil { 
			return AC, err
		}
		AC.ContactID = cid
	}
	AC.IsActive = is_active
	AC.IsUtility = is_utility
	AC.IsDelivery = is_delivery
	AC.IsPrimary = is_primary
	return AC, nil
}