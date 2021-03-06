[![Go Report Card](https://goreportcard.com/badge/github.com/mcluseau/autentigo)](https://goreportcard.com/report/github.com/mcluseau/autentigo)

## Running

```
openssl req -new -newkey rsa:2048 -days 365 -nodes -x509 -keyout tls.key -out tls.crt -subj /CN=localhost
export TLS_CRT="$(<tls.crt)" TLS_KEY="$(<tls.key)"
export SIGNING_METHOD=RS256
autentigo
```

### Request examples

Simple authentication:
```
$ curl -H'Content-Type: application/json' localhost:8080/simple -d'{"user":"test-user","password":"test-password"}' |jq .
{
  "token": "<TOKEN>",
  "claims": {
    "exp": 1531110508,
    "iat": 1531106908,
    "sub": "test-user",
    "display_name": "Display Name",
    "email": "email@example.com",
    "email_verified": true,
    "groups": [
      "group1",
      "group2"
    ]
  }
  "claims":
}
```

Basic authentication:
```
$ curl -i localhost:8080/basic
HTTP/1.1 401 Unauthorized
Www-Authenticate: Basic realm="Autorizo"
Date: Wed, 27 Jun 2018 06:50:59 GMT
Content-Length: 14
Content-Type: text/plain; charset=utf-8

Unauthorized.
```

```
$ curl --basic --user test-user:test-password localhost:8080/basic |jq .
{
  "token": "<TOKEN>",
  "claims": {
    "exp": 1531110508,
    "iat": 1531106908,
    "sub": "test-user",
    "display_name": "Display Name",
    "email": "email@example.com",
    "email_verified": true,
    "groups": [
      "group1",
      "group2"
    ]
  }
}
```

Basic authentication, setting only a cookie (also supported on /simple):
```
$ curl --basic --user test-user:test-password localhost:8080/basic -H'X-Set-Cookie: token' -i
HTTP/1.1 200 OK
Content-Type: application/json
Set-Cookie: token=<TOKEN>; HttpOnly; Secure
Date: Thu, 28 Jun 2018 22:59:57 GMT
Content-Length: 67

{
  "exp": 1530230397,
  "iat": 1530226797,
  "sub": "test-user",
  ...
 }
```

### Flags

```
autentigo --help
```

### Environment

| Variable         | Description
| ---------------- | ------------------------------------------------
| `TLS_CRT`        | The certificate to check tokens
| `TLS_KEY`        | The key to sign tokens
| `SIGNING_METHOD` | The signing method to use (https://tools.ietf.org/html/rfc7518#section-3.1)
| `AUTH_BACKEND`   | choose an authentication backend (default: stupid)

### Auth backends

#### stupid

Always accept the given credentials.

#### file

Reads a file, defined by the `AUTH_FILE` env, in the format:

```
<user name>:<password SHA256 (hex)>:email:email_validated:groups
```

Only user and password are required.

Adding an entry can be done this way:
```
echo test-user:$(echo -n test-password |sha256sum |awk '{print $1}'):Display Name:email@example.com:yes:group1,group2 >>users
```

#### LDAP simple bind

Tries to bind to an LDAP server, defined by the `LDAP_SERVER` env, with the given credentials and using `LDAP_USER`
as a username template.

Example:
```
AUTH_BACKEND=ldap-bind \
LDAP_SERVER=ldap://localhost:389 \
LDAP_USER=uid=%s,ou=users,dc=example,dc=com \
autentigo
```

#### etcd lookup

Looks up the user in etcd, with a key like `prefix/user-name`. Takes an optionnal `ETCD_TIMEOUT` to change the lookup timeout.

Example:
```sh
AUTH_BACKEND=etcd \
ETCD_ENDPOINTS=http://localhost:2379 \
ETCD_PREFIX=/users \
autentigo
```

Allowed extra claims in the etcd object:
```json
{
    "password_hash": "<password sha256, hex encoded)>",
    "groups": [ "app1-admin", "app2-reader" ],
    "display_name": "Display Name",
    "email": "user@host",
    "email_verified": true
}
```

#### SQL database lookup

Looks up the user in the SQL database.

Example:
```sh
AUTH_BACKEND=sql \
SQL_DRIVER=postgres\
SQL_DSN="user=postgres password=postgres host=localhost dbname=postgres sslmode=disable"\
SQL_USER_TABLE=users \
autentigo
```

Allowed extra claims in the object:
```json
{
    "password_hash": "<password sha256, hex encoded)>",
    "groups": [ "app1-admin", "app2-reader" ],
    "display_name": "Display Name",
    "email": "user@host",
    "email_verified": true
}
```
