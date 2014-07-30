package main

import (
	"fmt"
	"os"

	"launchpad.net/account-polld/qtcontact"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage:", os.Args[0], "[email address]")
		os.Exit(1)
	}

	path := qtcontact.GetAvatar(os.Args[1])
	fmt.Println("Avatar found:", path)
}
