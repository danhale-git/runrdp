package rdp

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"

	"github.com/atotto/clipboard"
	"github.com/skratchdot/open-golang/open"
)

// Connect writes an RDP file, runs it then deletes it 1 second later.
//func Connect(host, user, pass, path string, width, height, scale int) {
func Connect(rdp *RDP, debug bool) error {
	fb := fileBody(rdp.Address, rdp.Username)
	fb = settings(fb, rdp.Width, rdp.Height, 100)

	if rdp.Password != "" {
		fmt.Println("WARNING: Writing secret to clipboard - be careful where you paste!")

		err := clipboard.WriteAll(rdp.Password)

		if err != nil {
			log.Fatalf("writing password to clipboard: %s", err)
		}
	}

	path := viper.GetString("tempfile-path")

	writeFile(fb, path)

	runRDPFile(viper.GetString("tempfile-path"))
	// Ensure the file is deleted. Wait for 1 second before deleting it to allow the RDP application to read it.
	defer deleteFile(path)
	time.Sleep(1 * time.Second)

	return nil
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
	_ = os.Remove(path)
}

func writeFile(body, path string) {
	err := ioutil.WriteFile(path, []byte(body), 0644)
	if err != nil {
		panic(err)
	}
}

func runRDPFile(runPath string) {
	_ = open.Run(runPath)
}
