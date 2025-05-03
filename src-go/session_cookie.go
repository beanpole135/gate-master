package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/securecookie"
)

const cookie_name = "smcl"

var s *securecookie.SecureCookie

// Initialization function for the cookie system
func setupSecureCookies() {
	cookie_hashkey := []byte(CONFIG.Auth.HashKey)   //32-64 characters recommended
	cookie_blockkey := []byte(CONFIG.Auth.BlockKey) //Must be 32 characters long
	updateConfig := false
	if len(cookie_hashkey) == 0 {
		cookie_hashkey = securecookie.GenerateRandomKey(32)
		updateConfig = true
	}
	if len(cookie_blockkey) == 0 {
		cookie_blockkey = securecookie.GenerateRandomKey(32)
		updateConfig = true
	}
	if CONFIG.Auth.JwtSecret == "" {
		CONFIG.Auth.JwtSecret = RandomString(32)
		updateConfig = true
	}
	if updateConfig {
		CONFIG.Auth.HashKey = string(cookie_hashkey)
		CONFIG.Auth.BlockKey = string(cookie_blockkey)
		UpdateConfig(CONFIG)
	}
	//Note: If random keys are used - then whenever the service is restarted
	//  any clients currently going through the curity process will be rendered invalid
	//  (clients currently on the Curity login page/process)
	//  and this will result in a BadRequest error post-Curity
	s = securecookie.New(cookie_hashkey, cookie_blockkey)
}

// SetTokenCookies : Post-login cookie setting routine
func SetTokenCookie(w http.ResponseWriter, r *http.Request, at string) {
	// at: access token
	SetSecureCookieToken(w, r, at)
}

func DeleteTokenCookie(w http.ResponseWriter, r *http.Request) {
	deleteSecureCookie(w, r, cookie_name)
}

// === REFRESH TOKEN ===
func SetSecureCookieToken(w http.ResponseWriter, r *http.Request, at string) {
	value := map[string]string{
		"at": at,
	}
	_ = writeSecureCookie(w, r, cookie_name, value)
}

func ReadSecureCookieTokens(w http.ResponseWriter, r *http.Request) (at string) {
	value, err := readSecureCookie(w, r, cookie_name)
	if err != nil {
		return ""
	}
	if val, ok := value["at"]; ok {
		at = val
	}
	return at
}

// ----------------------------
// Internal simplification functions
// ----------------------------
func writeSecureCookie(w http.ResponseWriter, r *http.Request, name string, vals map[string]string) error {
	encoded, err := s.Encode(name, vals)
	if err == nil {
		cookie := &http.Cookie{
			Name:     name,
			Value:    encoded,
			Path:     "/",
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
	}
	return err
}

func deleteSecureCookie(w http.ResponseWriter, r *http.Request, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "", //Setting a blank value for the cookie is what removes it
		Path:     "/",
		MaxAge:   -1, //so the browser knows to wipe/remove it as expired automatically
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

func readSecureCookie(w http.ResponseWriter, r *http.Request, name string) (map[string]string, error) {
	value := make(map[string]string)
	cookie, err := r.Cookie(name)
	if err == nil {
		err = s.Decode(name, cookie.Value, &value)
		if err != nil {
			fmt.Println("Could not decode cookie")
			return nil, err
		}
	}
	return value, err
}
