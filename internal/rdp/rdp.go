package rdp

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/skratchdot/open-golang/open"
)

var path = "connection.rdp"

func Connect(host, user, pass string) {
	createFile(host, user, pass)
	runRDPFile(path)
	// Ensure the file is deleted. Wait for 1 second before deleting it to allow the RDP application to read it.
	defer deleteFile()
	time.Sleep(1 * time.Second)
}

func createFile(host, user, pass string) {
	enc, err := encrypt(pass)
	if err != nil {
		log.Fatalf("encrypt failed: %v", err)
	}

	body := fmt.Sprintf(
		`auto connect:i:1
		prompt for credentials:i:0
		full address:s:%s
		username:s:%s
		password 51:b:%s`,
		host,
		user,
		fmt.Sprintf("%x", enc),
	)
	writeFile(body, path)
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
