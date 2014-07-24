package main

import (
	"fmt"
	"os"

	"launchpad.net/account-polld/accounts"
)

func main() {
	// Expects a list of service names as command line arguments
	for data := range accounts.NewWatcher(os.Args[1]).C {
		if data.Error != nil {
			fmt.Println("Failed to authenticate account", data.AccountId, ":", data.Error)
		} else {
			fmt.Printf("%#v\n", data)
		}
	}
}
