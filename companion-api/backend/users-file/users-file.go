package usersfile

import (
	"io"
	"strconv"
	"strings"
	"sync"

	"github.com/mcluseau/autorizo/auth"
	"github.com/mcluseau/autorizo/companion-api/api"
	"github.com/mcluseau/autorizo/companion-api/backend"
)

var toBool = map[string]bool{
	"true":  true,
	"yes":   true,
	"1":     true,
	"false": false,
	"no":    false,
	"0":     false,
}

type fileClient struct {
	filePath string
}

// New Client to manage users with a csv file backend
func New(filePath string) backend.Client {
	return &fileClient{
		filePath: filePath,
	}
}

var _ backend.Client = &fileClient{}

func (fc *fileClient) CreateUser(id string, user *backend.UserData) (err error) {

	oldUser := &backend.UserData{}
	oldUser, err = fc.getUser(id)

	if oldUser != nil {
		err = api.ErrUserAlreadyExist
	} else if err == api.ErrMissingUser {
		err = fc.putUser(id, user.PasswordHash, user.ExtraClaims)
	}

	return
}

func (fc *fileClient) UpdateUser(id string, update func(user *backend.UserData) error) (err error) {
	user := &backend.UserData{}
	user, err = fc.getUser(id)

	if err == nil && user != nil {
		err = update(user)
		if err == nil {
			err = fc.putUser(id, user.PasswordHash, user.ExtraClaims)
		}
	}

	return
}

func (fc *fileClient) DeleteUser(id string) error {

	var wg sync.WaitGroup

	reader, err := newUsersFileReader(fc.filePath)
	if err != nil {
		return err
	}
	defer reader.close()

	writer, err := newUsersFileWriter()
	if err != nil {
		return err
	}
	defer writer.save(fc.filePath)

	recordExist := false
	for {
		record, err := reader.read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if id == record[0] {
			recordExist = true
		} else {
			wg.Add(1)
			go func(r []string) {
				writer.write(r)
				wg.Done()
			}(record)
		}

	}

	wg.Wait()
	if !recordExist {
		return api.ErrMissingUser
	}

	return nil
}

func (fc *fileClient) putUser(id, passwordHash string, claims auth.ExtraClaims) error {
	var wg sync.WaitGroup

	reader, err := newUsersFileReader(fc.filePath)
	if err != nil {
		return err
	}
	defer reader.close()

	writer, err := newUsersFileWriter()
	if err != nil {
		return err
	}
	defer writer.save(fc.filePath)

	recordExist := false
	for {
		record, err := reader.read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if id == record[0] {
			record[1] = passwordHash
			record[2] = claims.DisplayName
			record[3] = claims.Email
			record[4] = strconv.FormatBool(claims.EmailVerified)
			record[5] = strings.Join(claims.Groups, ",")
			recordExist = true
		}

		wg.Add(1)
		go func(record []string) {
			writer.write(record)
			wg.Done()
		}(record)

	}

	wg.Wait()
	if !recordExist {
		writer.write([]string{
			id,
			passwordHash,
			claims.DisplayName,
			claims.Email,
			strconv.FormatBool(claims.EmailVerified),
			strings.Join(claims.Groups, ","),
		})
	}

	return nil
}

func (fc *fileClient) getUser(id string) (*backend.UserData, error) {

	reader, err := newUsersFileReader(fc.filePath)
	if err != nil {
		return nil, err
	}
	defer reader.close()

	for {
		record, err := reader.read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if len(record) < 2 {
			// record too short
			continue
		}

		if id != record[0] {
			continue
		}

		user := &backend.UserData{
			PasswordHash: record[1],
		}

		l := len(record)
		switch {
		case l >= 6:
			user.ExtraClaims.Groups = strings.Split(record[5], ",")
			fallthrough
		case l == 5:
			user.ExtraClaims.EmailVerified = toBool[record[4]]
			fallthrough
		case l == 4:
			user.ExtraClaims.Email = record[3]
			fallthrough
		case l == 3:
			user.ExtraClaims.DisplayName = record[2]
		}

		return user, nil
	}

	return nil, api.ErrMissingUser
}
