package etcd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
	jwt "github.com/dgrijalva/jwt-go"

	"github.com/mcluseau/autentigo/api"
	"github.com/mcluseau/autentigo/auth"
)

const (
	oauthprefix = "/oauth"
)

// New Authenticator with etcd backend
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

// User describe an user stored in etcd
type User struct {
	PasswordHash string `json:"password_hash"`
	auth.ExtraClaims
}

func (a *etcdAuth) Authenticate(user, password string, expiresAt time.Time) (claims jwt.Claims, err error) {

	ba := sha256.Sum256([]byte(password))
	passwordHash := hex.EncodeToString(ba[:])

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	u := &User{}
	if u, err = a.getUser(ctx, user); err != nil {
		return
	}

	if u.PasswordHash != passwordHash {
		err = api.ErrInvalidAuthentication
		return
	}

	claims = auth.Claims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: expiresAt.Unix(),
			Subject:   user,
		},
		ExtraClaims: u.ExtraClaims,
	}
	return
}

func (a *etcdAuth) getUser(ctx context.Context, userID string) (user *User, err error) {
	resp, err := a.client.Get(ctx, path.Join(a.prefix, userID))
	if err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		err = api.ErrInvalidAuthentication
		return
	}

	user = &User{}
	err = json.Unmarshal(resp.Kvs[0].Value, user)
	return
}

func (a *etcdAuth) FindUser(clientID, provider string, expiresAt time.Time) (userID string, claims jwt.Claims, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	var resp *clientv3.GetResponse
	if resp, err = a.client.Get(ctx, path.Join(oauthprefix, a.prefix, provider, clientID)); err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		err = errors.New("unknown user")
		return
	}

	userID = string(resp.Kvs[0].Value)
	user := &User{}
	if user, err = a.getUser(ctx, userID); err != nil {
		return
	}

	claims = auth.Claims{
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: expiresAt.Unix(),
			Subject:   userID,
		},
		ExtraClaims: user.ExtraClaims,
	}
	return
}
