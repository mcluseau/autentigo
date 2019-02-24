package main

import (
	"flag"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

var (
	rbacFile = flag.String("rbac-file", "/etc/autentigo/rbac.yaml", "HTTP bind specification")

	rbacRules []RBACRule
)

type RBACRule struct {
	Role   string
	Users  []string
	Groups []string
}

func loadRBAC() (err error) {
	r := make([]RBACRule, 0)

	ba, err := ioutil.ReadFile(*rbacFile)
	if err != nil {
		return
	}

	if err = yaml.UnmarshalStrict(ba, &r); err != nil {
		return
	}

	rbacRules = r
	return
}
