package auth

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Users struct {
	Users []User `yaml:"users"`
}

func (u Users) ValidateUser(username, password string) bool {
	for _, user := range u.Users {
		if user.Username == username && user.Password == password {
			return true
		}
	}

	return false
}

func (u Users) AreUsers() bool {
	return len(u.Users) > 0
}

func ReadUsersFromFile(userFile string) *Users {
	fileLocation := usersFile(userFile)
	return readConfig(fileLocation)

}

func usersFile(path string) string {
	if len(path) > 0 {
		return path
	}
	return "users.yml"
}

func readConfig(filename string) *Users {
	var users *Users
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return &Users{}
	}
	users, err = unmarshal(data)

	return users
}

func unmarshal(data []byte) (*Users, error) {
	var users *Users
	err := yaml.Unmarshal(data, &users)

	return users, err
}
