## Running

```
openssl req -new -newkey rsa:2048 -days 365 -nodes -x509 -keyout tls.key -out tls.crt -subj /CN=localhost
export TLS_CRT="$(<tls.crt)" TLS_KEY="$(<tls.key)"
export SIGNING_METHOD=RS256
autorizo
```

### Request examples

Simple authentication:
```
$ curl -H'Content-Type: application/json' localhost:8080/simple -d'{"user":"test-user","password":"test-password"}'
{
  "token": "<TOKEN>"
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
$ curl --basic --user test-user:test-password localhost:8080/basic
{
  "token": "<TOKEN>"
 }
```

Basic authentication, setting only a cookie (also supported on /simple):
```
$ curl --basic --user test-user:test-password localhost:8080/basic -H'X-Set-Cookie: token' -i
HTTP/1.1 201 Created
Set-Cookie: token=<TOKEN>; HttpOnly; Secure
Date: ...
Content-Length: 0


```

### Flags

```
autorizo --help
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
<user name>:<password SHA256 (hex)>
```

Adding an entry:

```
echo test-user:$(echo -n test-password |sha256sum |awk '{print $1}') >>users
```

#### LDAP simple bind

Tries to bind to an LDAP server, defined by the `LDAP_SERVER` env, with the given credentials and using `LDAP_USER`
as a username template.

Example:
```
AUTH_BACKEND=ldap-bind \
LDAP_SERVER=ldap://localhost:389 \
LDAP_USER=uid=%s,ou=users,dc=example,dc=com \
autorizo
```

#### etcd lookup

Looks up the user in etcd, with a key like `prefix/user-name`. Takes an optionnal `ETCD_TIMEOUT` to change the lookup timeout.

Example:
```sh
AUTH_BACKEND=etcd \
ETCD_ENDPOINTS=http://localhost:2379 \
ETCD_PREFIX=/users \
autorizo
```

Allowed extra claims in the etcd object:
```json
{
    "password_hash": "<password sha256, hex encoded)>",
    "groups": [ "app1-admin", "app2-reader" ],
    "email": "user@host",
    "email_verified": true
}
```
