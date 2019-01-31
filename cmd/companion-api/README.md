## Running

#### With file backend

```sh
export AUTH_BACKEND=file \
export AUTH_FILE="autorizo-users.csv" \
companion-api
```

#### With etcd backend

```sh
export AUTH_BACKEND=etcd \
export ETCD_ENDPOINTS=http://localhost:2379 \
export ETCD_PREFIX=/users \
companion-api
```

### Flags

```
companion-api --help
```

### Environment

| Variable         | Description                                                                            |
| ---------------- | -------------------------------------------------------------------------------------- |
| `ETCD_TIMEOUT`   | Simple etcd timeout (default: 5s)                                                      |
| `ETCD_PREFIX`    | Prefix before the etcd key (default: none)                                             |
| `ETCD_ENDPOINTS` | Etcd endpoints (format: `ETCD_ENDPOINTS`=http://localhost:2379,http://localhost:4001 ) |
| `AUTH_FILE`      | Backend file (required if `AUTH_BACKEND`=file)                                         |
| `AUTH_BACKEND`   | Choose an authentication backend (required)                                            |

### Auth backends

#### stupid

Stupid backend does not need the companion-api.

#### file

Reads or update a content file, defined by the `AUTH_FILE` env, in the format:

```
<user name>:<password SHA256 (hex)>:email:email_validated:groups
```

#### LDAP simple bind

Please feel free to use a ldap client instead of the companion-api.

#### etcd lookup

Update or looks up the user in etcd, with a key like `prefix/user-name`. Takes an optionnal `ETCD_TIMEOUT` to change the lookup timeout.
