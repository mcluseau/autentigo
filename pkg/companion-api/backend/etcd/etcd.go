package etcd

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/mcluseau/autentigo/pkg/companion-api/api"
	"github.com/mcluseau/autentigo/pkg/companion-api/backend"
)

const (
	oauthprefix = "/oauth"
)

type etcdClient struct {
	prefix  string
	client  *clientv3.Client
	timeout time.Duration
}

// New Client to manage users with an etcd backend
func New(prefix string, endpoints []string) backend.Client {
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

	return &etcdClient{
		prefix:  prefix,
		client:  client,
		timeout: timeout,
	}
}

var _ backend.Client = &etcdClient{}

func (e *etcdClient) CreateUser(id string, user *backend.UserData) (err error) {
	oldUser := &backend.UserData{}
	oldUser, err = e.getUser(id)

	if oldUser != nil {
		err = api.ErrUserAlreadyExist
	} else if err == api.ErrMissingUser {
		err = e.putUser(id, user)
	}

	return
}

func (e *etcdClient) UpdateUser(id string, update func(user *backend.UserData) error) (err error) {
	user := &backend.UserData{}
	user, err = e.getUser(id)

	if err == nil && user != nil {
		err = update(user)
		if err == nil {
			err = e.putUser(id, user)
		}
	}

	return
}

func (e *etcdClient) DeleteUser(id string) (err error) {
	user := &backend.UserData{}
	user, err = e.getUser(id)

	if err == nil && user != nil {
		ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
		defer cancel()

		_, err = e.client.Delete(ctx, path.Join(e.prefix, id))
	}
	return
}

func (e *etcdClient) PutUserID(provider, clientID, userID string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	_, err = e.client.Put(ctx, path.Join(oauthprefix, e.prefix, provider, clientID), userID)

	return
}

func (e *etcdClient) GetUserID(provider, clientID string) (userID string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	var resp *clientv3.GetResponse
	if resp, err = e.client.Get(ctx, path.Join(oauthprefix, e.prefix, provider, clientID)); err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		err = api.ErrMissingUser
		return
	}

	userID = string(resp.Kvs[0].Value)
	return
}

func (e *etcdClient) getUser(id string) (user *backend.UserData, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	var resp *clientv3.GetResponse
	if resp, err = e.client.Get(ctx, path.Join(e.prefix, id)); err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		err = api.ErrMissingUser
		return
	}

	user = &backend.UserData{}
	err = json.Unmarshal(resp.Kvs[0].Value, user)
	return
}

func (e *etcdClient) putUser(id string, user *backend.UserData) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	u, err := json.Marshal(*user)
	if err == nil {
		_, err = e.client.Put(ctx, path.Join(e.prefix, id), string(u))
	}

	return

}
