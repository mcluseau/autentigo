package rbac

// Rule is a simple RBAC rule to match a role
type Rule struct {
	Role   string
	Users  []string
	Groups []string
}

func (r Rule) Match(user *User) bool {
	for _, u := range r.Users {
		if u == user.Name {
			return true
		}
	}

	for _, ug := range user.Groups {
		for _, rg := range r.Groups {
			if ug == rg {
				return true
			}
		}
	}

	return false
}
