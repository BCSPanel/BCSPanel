package user

import (
	"fmt"
	"io/fs"
	"os"
)

const passwordFileMode = fs.FileMode(0600)

type PasswordV1 struct {
	Password [64]byte
	Salt     [16]byte
}

func getPasswordFromId(id int) (password *PasswordV1, err error) {
	if id == 0 {
		err = fmt.Errorf("getPasswordFromId: id == 0")
		return
	}
	fb, err := os.ReadFile(fmt.Sprint(DataUsersDirPath, "/", id, ".passwd"))
	if err != nil {
		return
	}
	if len(fb) < 1+64+16 {
		err = fmt.Errorf("getPasswordFromId: The file length is incorrect")
		return
	}
	if fb[0] != 1 {
		err = fmt.Errorf("getPasswordFromId: undefined version %d", fb[0])
		return
	}
	password = &PasswordV1{}
	copy(password.Password[:], fb[1:65])
	copy(password.Salt[:], fb[65:81])
	return
}

func GetPasswordFromId(id int) (password *PasswordV1, err error) {
	lock.Lock()
	defer lock.Unlock()
	return getPasswordFromId(id)
}

func (password *PasswordV1) writeForId(id int) (err error) {
	if id == 0 {
		err = fmt.Errorf("writePasswordForId: id == 0")
		return
	}
	b := make([]byte, 1+64+16)
	b[0] = 1
	copy(b[1:65], password.Password[:])
	copy(b[65:81], password.Salt[:])

	err = os.WriteFile(fmt.Sprint(DataUsersDirPath, "/", id, ".passwd"), b, passwordFileMode)
	return
}

func (password *PasswordV1) WriteForId(id int) (err error) {
	lock.Lock()
	defer lock.Unlock()
	return password.writeForId(id)
}
