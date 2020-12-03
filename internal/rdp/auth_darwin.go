package rdp

import (
	"fmt"
	"log"

	"github.com/atotto/clipboard"
)

// CrossPlatformAuthHandler handles the usage of the password. On Windows this is encrypted and included in the .rdp
// file. On Mac OS this is not possible. Currently Mac users will just have the password copied to their clipboard.
// The correct way to do this on mac is probably using the KeyRing. See https://github.com/danhale-git/runrdp/issues/38.
func CrossPlatformAuthHandler(fileBody, password string) string {
	fmt.Println("WARNING: Writing secret to clipboard - be careful where you paste!")

	err = clipboard.WriteAll(password)

	if err != nil {
		log.Fatalf("writing password to clipboard: %s", err)
	}

	return ""
}
