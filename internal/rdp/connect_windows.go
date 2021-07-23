package rdp

import (
	"fmt"
	"os/exec"
	"strings"
)

func Connect(rdp *RDP, debug bool) error {
	if rdp.Address == "" {
		return fmt.Errorf("address is an empty string, nothing to connect to")
	}

	if rdp.Username != "" {
		cmdkeyCreateArgs := []string{
			fmt.Sprintf("/generic:%s", rdp.Address),
			fmt.Sprintf("/user:%s", rdp.Username),
		}

		if rdp.Password != "" {
			cmdkeyCreateArgs = append(cmdkeyCreateArgs, fmt.Sprintf("/pass:%s", rdp.Password))
		}

		cmdkeyCreate := exec.Command("cmdkey", cmdkeyCreateArgs...)
		cmdkeyDelete := exec.Command("cmdkey", fmt.Sprintf("/delete:%s", rdp.Address))

		if debug {
			fmt.Println(strings.Replace(cmdkeyCreate.String(), rdp.Password, "REMOVED", 1))
		}
		if err := cmdkeyCreate.Run(); err != nil {
			return fmt.Errorf("creating credentials using cmdkey.exe: %w", err)
		}

		defer func() {
			if debug {
				fmt.Println(cmdkeyDelete.String())
			}
			if err := cmdkeyDelete.Run(); err != nil {
				panic(err)
			}
		}()
	}

	mstscArgs := []string{
		fmt.Sprintf("/v:%s:%s", rdp.Address, rdp.Port),
	}

	if rdp.Width != 0 {
		mstscArgs = append(mstscArgs, fmt.Sprintf("/w:%d", rdp.Width))
	}
	if rdp.Height != 0 {
		mstscArgs = append(mstscArgs, fmt.Sprintf("/h:%d", rdp.Height))
	}

	if rdp.Fullscreen {
		mstscArgs = append(mstscArgs, "/f")
	}
	if rdp.Public {
		mstscArgs = append(mstscArgs, "/public")
	}
	if rdp.Span {
		mstscArgs = append(mstscArgs, "/span")
	}

	startSession := exec.Command("mstsc", mstscArgs...)

	if debug {
		fmt.Println(startSession.String())
	}
	if err := startSession.Run(); err != nil {
		return fmt.Errorf("running rdp session using mstsc.exe: %w", err)
	}

	return nil
}
