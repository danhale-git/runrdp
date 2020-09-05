package configure

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func CheckExistence(path, description string, dir bool) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("%s does not exist\ncreate at %s? (y/n): ", description, path)

		yes := interactiveYesNo()
		if yes {
			if dir {
				fmt.Println("created directory ", path)

				err = os.MkdirAll(path, 700)
				if err != nil {
					fmt.Printf("failed to create directory %s: %s\n", path, err)
				}
			} else {
				fmt.Println("created file ", path)

				_, err = os.Create(path)
				if err != nil {
					fmt.Printf("failed to create file %s: %s\n", path, err)
				}
			}
			// User chose to create
			return true
		}
		// User chose not to create
		return false
	}
	// Item already exists
	return true
}

func interactiveString(lower bool) (string, bool) {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	if lower {
		t := strings.TrimSpace(strings.ToLower(text))
		return t, t != ""
	}

	t := strings.TrimSpace(text)
	return t, t != ""
}

func interactiveYesNo() bool {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	return strings.TrimSpace(strings.ToLower(text)) == "y"
}
