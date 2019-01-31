package client

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

func (c *Client) Validate(tokenString string, claims jwt.Claims) (isValid bool, err error) {
	if c.validationCrt == nil {
		err = c.RefreshValidationCertificate()
		if err != nil {
			return
		}
	}

	if claims == nil {
		claims = jwt.MapClaims{}
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		switch alg := token.Method.Alg(); alg {
		case "ES256", "ES384", "ES512":
			return jwt.ParseECPublicKeyFromPEM(c.validationCrt)

		case "RS256", "RS384", "RS512":
			return jwt.ParseRSAPublicKeyFromPEM(c.validationCrt)

		default:
			return nil, fmt.Errorf("unknown signing method: %s", alg)
		}
	})

	if err != nil {
		return
	}

	isValid = token.Valid
	return
}

func (c *Client) RefreshValidationCertificate() (err error) {
	resp, err := http.Get(c.ServerURL + "/validation-certificate")
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected HTTP status: %d (%s)", resp.StatusCode, resp.Status)
	}

	crt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	c.validationCrt = crt
	return
}
