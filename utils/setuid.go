package utils

import (
	"os/user"
	"strconv"
	"syscall"
)

func SetUserID() {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		panic(err)
	}
	err = syscall.Seteuid(uid)
	if err != nil {
		panic(err)
	}
}
