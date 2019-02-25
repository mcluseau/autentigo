package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/mirror"
)

var (
	etcd *clientv3.Client

	etcdURL    = flag.String("etcd", "http://localhost:2379", "etcd URL")
	etcdPrefix = flag.String("etcd-prefix", "/users", "Prefix of etcd keys")
	passwdFile = flag.String("passwd-file", "passwd", "Dovecot passwd file")

	values = map[string]string{}
)

func main() {
	flag.Parse()

	// connect to etcd
	var err error
	etcd, err = clientv3.NewFromURL(*etcdURL)
	if err != nil {
		log.Fatal("failed to connect to etcd: ", err)
	}

	if len(*etcdPrefix) != 0 && !strings.HasSuffix(*etcdPrefix, "/") {
		*etcdPrefix += "/"
	}

	log.Print("connected to etcd with prefix ", *etcdPrefix)

	// prepare mirror
	rev := loadState()

	if rev == 0 {
		rev = initializeFromScratch()
	} else {
		loadFile()
	}

	followChanges(rev)
}

func initializeFromScratch() (rev int64) {
	log.Print("initializing from scratch")
	sync := mirror.NewSyncer(etcd, *etcdPrefix, 0)

	gets, errors := sync.SyncBase(context.Background())

	for get := range gets {
		rev = get.Header.Revision
		for _, kv := range get.Kvs {
			setValue(kv.Key, kv.Value)
		}
	}

	if err := <-errors; err != nil {
		log.Fatal("failed to read from etcd: ", err)
	}

	save(rev)
	return
}

func followChanges(rev int64) {
	log.Print("following changes")

	ctx := context.Background()

	for {
		sync := mirror.NewSyncer(etcd, *etcdPrefix, rev)
		for update := range sync.SyncUpdates(ctx) {
			if err := update.Err(); err != nil {
				log.Print("syncer error: ", err)
				continue
			}

			for _, event := range update.Events {
				if event.Kv == nil {
					delValue(event.Kv.Key)
				} else {
					setValue(event.Kv.Key, event.Kv.Value)
				}
			}

			rev = update.Header.Revision
			save(rev)
		}

		log.Print("sync updates ended, restarting")
	}
}

func keyFrom(key []byte) string {
	return string(key[len(*etcdPrefix):])
}

func setValue(fullKey []byte, value []byte) {
	key := keyFrom(fullKey)
	log.Print("update on ", key)

	v := struct {
		PasswordHash string `json:"password_hash"`
	}{}

	if err := json.Unmarshal(value, &v); err != nil {
		log.Fatal("failed to parse value at key ", string(fullKey), ": ", err)
	}

	h := v.PasswordHash
	if !strings.HasPrefix(h, "{") {
		h = "{SHA256.HEX}" + h
	}

	values[key] = h
}

func delValue(fullKey []byte) {
	key := keyFrom(fullKey)
	log.Print("delete on ", key)

	delete(values, key)
}
