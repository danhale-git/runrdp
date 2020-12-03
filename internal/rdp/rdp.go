package rdp

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/spf13/viper"

	"github.com/skratchdot/open-golang/open"
)

// DefaultPort is the standard port for RDP connections.
const DefaultPort = "3389"

var path = "connection.rdp"

// Connect writes an RDP file, runs it then deletes it 1 second later.
func Connect(host, user, pass string) {
	cmdlineUser := viper.GetString("username")
	cmdlinePass := viper.GetString("password")

	if cmdlineUser != "" {
		user = cmdlineUser
	}

	if cmdlinePass != "" {
		pass = cmdlinePass
	}

	fb := fileBody(host, user, pass)

	if pass != "" {
		fb = CrossPlatformAuthHandler(fb, pass)
	}

	writeFile(fb, path)

	runRDPFile(path)
	// Ensure the file is deleted. Wait for 1 second before deleting it to allow the RDP application to read it.
	defer deleteFile()
	time.Sleep(1 * time.Second)
}

func fileBody(host, user, pass string) string {
	body := fmt.Sprintf(
		`auto connect:i:1
		prompt for credentials:i:0
		full address:s:%s`,
		host,
	)

	if user != "" {
		body += fmt.Sprintf("\nusername:s:%s", user)
	}

	return body
}

func deleteFile() {
	os.Remove(path)
}

func writeFile(body, path string) {
	err := ioutil.WriteFile(path, []byte(body), 0644)
	if err != nil {
		panic(err)
	}
}

func runRDPFile(runPath string) {
	open.Run(runPath)
}
