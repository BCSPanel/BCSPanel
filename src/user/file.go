package user

import (
	"os"

	"github.com/bytedance/sonic"
)

const dir = "data/users/"
const filePerm = os.FileMode(0600)

func userNameToSysFileName(name string) string {
	return dir + "@" + name
}

func readUserFromSysFile(name string) (*User, error) {
	b, err := os.ReadFile(userNameToSysFileName(name))
	if err != nil {
		return nil, err
	}
	u := &User{}
	err = sonic.ConfigDefault.Unmarshal(b, u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func userSysFileExist(name string) bool {
	_, err := os.Stat(userNameToSysFileName(name))
	return err == nil
}

func writeUserToSysFile(u *User) error {
	b, err := sonic.ConfigDefault.Marshal(u)
	if err == nil {
		os.MkdirAll(dir, filePerm)
		err = os.WriteFile(userNameToSysFileName(u.Name), b, filePerm)
	}
	return err
}

func removeUserSysFile(name string) error {
	return os.Remove(userNameToSysFileName(name))
}
