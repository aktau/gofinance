package util

import (
	"fmt"
	"os/user"
)

func Home() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("error: can't get current user, returning working dir, ", err)
		return "."
	}
	return usr.HomeDir
}
