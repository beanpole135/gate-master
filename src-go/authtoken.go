package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthToken struct {
	UserId int32
	IsAdmin bool
}

// CreateSignedToken : Convert the CurrentToken struct into a signed JWT
func CreateSignedToken(ct AuthToken, jwtSecretKey string, expiresIn int) (string, error) {
	// ct : Current Token structure we need to include in the JWT
	// jwtSecretKey : A static value read from the service config - this is the encryption key for the tokens
	// expiresIn : Integer number of seconds that the token is considered valid

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		//Standard IANA claims (https://datatracker.ietf.org/doc/html/rfc7519#section-4.1)
		//"iss":            claims["iss"],		//ISSuer: Optional use - no need for it at the moment (sig verify is enough)
		//"sub":            claims["sub"],		//SUBject: Optional use - no need for it at the moment (sig verify is enough)
		"aud": ct.UserId,                        //real AUDience: Optional use - we will put the database User ID here.
		"nbf": time.Now().Unix() - 5,                //Not BeFore: Unix Timestamp (allow 5s margin for clock skew)
		"iat": time.Now().Unix(),                    //Issued AT: Unix Timestamp
		"exp": time.Now().Unix() + int64(expiresIn), //EXPires: Unix Timestamp
		//Non-standard / Custom Claims
		"iad": ct.IsAdmin, //Is ADmin
		
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return "", fmt.Errorf("CreateJwtToken: cannot sign: %w", err)
	}

	return tokenString, nil
}

// ReadCurrentToken : Primary function for reading the network context and pulling out the current token
func ReadSignedToken(tok string, jwtSecretKey string, checkTime bool) (AuthToken, error) {
	// ctx: The mrpc network context for the API request
	// jwtSecretKey: A static value read from the service config - this is the encryption key for the tokens
	// checkTime: if True - will also validate the JWT timestamps for validity

	// Dev Note: The checkTime bypass is needed for the case of refresh token usage.
	//   In this case, the refresh token is valid but the access token might be expired.
	//   This bypass allows us to grab info off the expired access token before re-generating a new JWT

	var ct AuthToken

	// Sanitize the token string
	tokenString := tok
	if strings.HasPrefix(tok, "Bearer ") {
		tokenString = tok[7:] //Chop the token type off the front of the string
	}

	// Now verify the signature on the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// First verify the signing algorithm is what you expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		// Now return the secret key for the signature for parsing/validation
		return []byte(jwtSecretKey), nil
	})
	if err != nil {
		return ct, fmt.Errorf("cannot parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		//Cannot read the claims off the token
		return ct, fmt.Errorf("invalid token")
	}
	if checkTime {
		// One final token validation check (validating timestamps and such)
		if !ok || !token.Valid {
			return ct, fmt.Errorf("invalid token")
		}
	}
	// At this point, we verified we have a valid token - now pull the info into the CurrentToken struct
	// Dev Note: This mapping needs to 100% match the token generation routine otherwise we get weird errors
	ct.UserId, _ = parseClaimInt(claims["aud"])
	ct.IsAdmin, _ = claims["iad"].(bool)
	return ct, nil
}

func parseClaimInt(val interface{}) (int32, bool) {
	//Numeric claims always get detected as floats for some reason
	f, err := val.(float64)
	return int32(f), err
}
