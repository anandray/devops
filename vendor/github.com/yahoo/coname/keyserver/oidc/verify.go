package oidc

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type ErrExpired struct {
	err error
}

func (e ErrExpired) Error() string {
	return e.err.Error()
}

// VerifyIDToken parses and validates the ID token received from the provider
// Apart from the signature validation, we care about the following fields:
// exp - token must not be expired
// iat - token must not be older than a duration(specified in the config)
// iss - must match issuer specified in the config
// aud - must match the clientID specified in the config
// email_verified - must be true
// nonce - must be validated by the client
func (c *Client) VerifyIDToken(token string) (email string, err error) {

	tok, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("\"kid\" not a string")
		}
		if c.pubKeys == nil {
			return nil, fmt.Errorf("No public key found to verify token")
		}
		return c.pubKeys[kid], nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return "", &ErrExpired{err: err}
			}
		}
		return "", err
	}
	if tok.Valid {
		claims := tok.Claims.(jwt.MapClaims)
		aud, ok := claims["aud"].(string)
		if !ok {
			return "", fmt.Errorf("\"aud\" not a string")
		}
		if aud != c.ClientID {
			return "", fmt.Errorf("ClientID invalid - got (%s) wanted (%s)", aud, c.ClientID)
		}

		iss, ok := claims["iss"].(string)
		if !ok {
			return "", fmt.Errorf("\"iss\" not a string")
		}
		if iss != c.Issuer {
			return "", fmt.Errorf("iss invalid - got (%s) wanted (%s)", iss, c.Issuer)
		}

		iat, ok := claims["iat"].(float64)
		if !ok {
			return "", fmt.Errorf("\"iat\" not an integer")
		}

		if c.Validity != 0 { // skip this check if Validity=0
			tm := time.Unix(int64(iat), 0)
			if time.Now().Sub(tm) > c.Validity {
				return "", fmt.Errorf("\"iat\" too old")
			}
		}
		emailVrf, ok := claims["email_verified"].(bool)
		if !ok {
			return "", fmt.Errorf("\"email_verified\" missing or invalid type")
		}
		if !emailVrf {
			return "", fmt.Errorf("email not verified")
		}
		return claims["email"].(string), err
	}
	return "", fmt.Errorf("Invalid token")
}
