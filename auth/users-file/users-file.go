package usersfile

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"io"
	"os"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/mcluseau/autorizo/api"
	"github.com/mcluseau/autorizo/auth"
)

var yesValues = map[string]bool{
	"true": true,
	"yes":  true,
	"1":    true,
}

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

	r := csv.NewReader(f)
	r.Comma = ':'

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if len(record) < 2 {
			// record too short
			continue
		}

		fileUser, hash := record[0], record[1]

		if user != fileUser || hash != passwordHash {
			continue
		}

		claims := auth.ExtraClaims{}

		l := len(record)
		switch {
		case l >= 6:
			claims.Groups = strings.Split(record[5], ",")
			fallthrough
		case l == 5:
			claims.EmailVerified = yesValues[record[4]]
			fallthrough
		case l == 4:
			claims.Email = record[3]
			fallthrough
		case l == 3:
			claims.DisplayName = record[2]
		}

		return auth.Claims{
			StandardClaims: jwt.StandardClaims{
				IssuedAt:  time.Now().Unix(),
				ExpiresAt: expiresAt.Unix(),
				Subject:   user,
			},
			ExtraClaims: claims,
		}, nil
	}

	return nil, api.ErrInvalidAuthentication
}
