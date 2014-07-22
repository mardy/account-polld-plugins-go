package main

import (
	"fmt"
	"os"

	"launchpad.net/account-polld/accounts"
)

func main() {
	// Expects a list of service names as command line arguments
	for data := range accounts.NewWatcher(os.Args[1:]...).C {
		fmt.Printf("%#v\n", data)
	}
}
