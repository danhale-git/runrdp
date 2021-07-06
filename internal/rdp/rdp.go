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

// Connect writes an RDP file, runs it then deletes it 1 second later.
func Connect(host, user, pass, path string, width, height, scale int) {
	cmdlineUser := viper.GetString("username")
	cmdlinePass := viper.GetString("password")

	if cmdlineUser != "" {
		user = cmdlineUser
	}

	if cmdlinePass != "" {
		pass = cmdlinePass
	}

	fb := fileBody(host, user)
	fb = settings(fb, width, height, scale)

	if pass != "" {
		fb = CrossPlatformAuthHandler(fb, pass)
	}

	writeFile(fb, path)

	runRDPFile(path)
	// Ensure the file is deleted. Wait for 1 second before deleting it to allow the RDP application to read it.
	defer deleteFile(path)
	time.Sleep(1 * time.Second)
}

func fileBody(host, user string) string {
	body := fmt.Sprintf(
		`screen mode id:i:1
auto connect:i:1
prompt for credentials:i:0
full address:s:%s`,
		host,
	)

	if user != "" {
		body += fmt.Sprintf("\nusername:s:%s", user)
	}

	return body
}

func settings(body string, width, height, scale int) string {
	if width != 0 {
		body = fmt.Sprintf("%s\ndesktopwidth:i:%d", body, width)
	}
	if height != 0 {
		body = fmt.Sprintf("%s\ndesktopheight:i:%d", body, height)
	}
	if scale != 0 {
		body = fmt.Sprintf("%s\ndesktopscalefactor:i:%d", body, scale)
	}
	return body
}

func deleteFile(path string) {
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
