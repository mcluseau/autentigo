package rbac

import "net/http"

// Interface of an RBAC backend
type Interface interface {
	Match(role string, user *User) bool
	MatchRequest(role string, req *http.Request, validationCrt []byte) (authn, authz bool)
}

// User describes a user for the simple RBAC backend
type User struct {
	Name   string
	Groups []string
}

var (
	// Default interface used for default matchers.
	Default Interface

	// DefaultValidationCertificate used for default matchers.
	DefaultValidationCertificate []byte
)

// SetDefaults sets everything up for default matchers.
func SetDefaults(iface Interface, validationCrt []byte) {
	Default = iface
	DefaultValidationCertificate = validationCrt
}

func Match(role string, user *User) bool {
	if Default == nil {
		return false
	}

	return Default.Match(role, user)
}

func MatchRequest(role string, req *http.Request) (authn, authz bool) {
	if Default == nil {
		return false, false
	}

	return Default.MatchRequest(role, req, DefaultValidationCertificate)
}
