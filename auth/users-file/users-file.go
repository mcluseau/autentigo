package usersfile

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/mcluseau/autorizo/api"
)

func New(filePath string) api.Authenticator {
	return &usersFileAuth{
		filePath: filePath,
	}
}

type usersFileAuth struct {
	filePath string
}

var _ api.Authenticator = usersFileAuth{}

func (a usersFileAuth) Authenticate(user, password string, expiresAt time.Time) (jwt.Claims, error) {
	ba := sha256.Sum256([]byte(password))
	passwordHash := hex.EncodeToString(ba[:])

	f, err := os.Open(a.filePath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	r := bufio.NewReader(f)

	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		fileUser, hash := parts[0], parts[1]

		if user == fileUser && hash == passwordHash {
			return jwt.StandardClaims{
				IssuedAt:  time.Now().Unix(),
				ExpiresAt: expiresAt.Unix(),
				Subject:   user,
			}, nil
		}
	}

	return nil, api.ErrInvalidAuthentication
}
