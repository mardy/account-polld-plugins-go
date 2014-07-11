package main

import (
	"fmt"
	"os"

	"launchpad.net/account-polld/accounts"
)

func main() {
	// Expects a list of service names as command line arguments
	for data := range accounts.WatchForService(os.Args[1:]...) {
		fmt.Printf("%#v\n", data)
	}
}
