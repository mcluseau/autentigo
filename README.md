## Running

```
openssl req -new -newkey rsa:2048 -days 365 -nodes -x509 -keyout tls.key -out tls.crt -subj /CN=localhost
export TLS_CRT="$(<tls.crt)" TLS_KEY="$(<tls.key)"
autorizo
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

