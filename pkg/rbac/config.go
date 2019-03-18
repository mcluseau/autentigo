package rbac

import (
	"io/ioutil"
	"net/http"

	yaml "github.com/projectcalico/go-yaml-wrapper"
)

// Config is a simple RBAC configuration.
type Config struct {
	// Roles every user has.
	Base []string

	// Rules to determine a user's roles.
	Rules []Rule
}

var _ Interface = &Config{}

func FromFile(path string) (config *Config, err error) {
	ba, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	return FromBytes(ba)
}

func FromBytes(ba []byte) (config *Config, err error) {
	config = &Config{}

	if err = yaml.UnmarshalStrict(ba, config); err != nil {
		return
	}

	return
}

func (c *Config) RolesOf(user *User) (roles []string) {
	roles = make([]string, 0)
	roles = append(roles, c.Base...)

	for _, rule := range c.Rules {
		if rule.Match(user) {
			roles = append(roles, rule.Role)
		}
	}

	return
}

func (c *Config) Match(role string, user *User) bool {
	if user == nil {
		return false
	}

	for _, baseRole := range c.Base {
		if baseRole == role {
			return true
		}
	}

	for _, rule := range c.Rules {
		if rule.Match(user) {
			return true
		}
	}

	return false
}

func (c *Config) MatchRequest(role string, req *http.Request, validationCrt []byte) (authn, authz bool) {
	u := UserFromRequest(req, validationCrt)
	if u == nil {
		return false, false
	}

	return true, c.Match(role, u)
}
