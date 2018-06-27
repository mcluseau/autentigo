package etcd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/mcluseau/autorizo/api"
)

func New(prefix string, endpoints []string) api.Authenticator {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})

	if err != nil {
		log.Fatal("failed to connect to etcd: ", err)
	}

	timeout := 5 * time.Second
	if timeoutEnv := os.Getenv("ETCD_TIMEOUT"); timeoutEnv != "" {
		timeout, err = time.ParseDuration(timeoutEnv)
		if err != nil {
			log.Fatalf("invalid ETCD_TIMEOUT %q: %v", timeoutEnv, timeout)
		}
	}

	return &etcdAuth{
		prefix:  prefix,
		client:  client,
		timeout: timeout,
	}
}

type etcdAuth struct {
	prefix  string
	client  *clientv3.Client
	timeout time.Duration
}

var _ api.Authenticator = &etcdAuth{}

type User struct {
	PasswordHash string `json:"password_hash"`
	ExtraClaims
}

type ExtraClaims struct {
	Email         string   `json:"email,omitempty"`
	EmailVerified bool     `json:"email_verified,omitempty"`
	Groups        []string `json:"groups,omitempty"`
}

type Claims struct {
	jwt.StandardClaims
	ExtraClaims
}

func (a *etcdAuth) Authenticate(user, password string, expiresAt time.Time) (claims jwt.Claims, err error) {
	ba := sha256.Sum256([]byte(password))
	passwordHash := hex.EncodeToString(ba[:])

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	resp, err := a.client.Get(ctx, path.Join(a.prefix, user))
	if err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		err = api.ErrInvalidAuthentication
		return
	}

	u := User{}
	if err = json.Unmarshal(resp.Kvs[0].Value, &u); err != nil {
		return
	}

	if u.PasswordHash != passwordHash {
		err = api.ErrInvalidAuthentication
		return
	}

	claims = Claims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: expiresAt.Unix(),
			Subject:   user,
		},
		ExtraClaims: u.ExtraClaims,
	}
	return
}
