package rdp

import (
	"fmt"
	"os/exec"
	"strings"
)

func Connect(rdp *RDP, debug bool) error {
	cmdkeyCreate := exec.Command("cmdkey",
		fmt.Sprintf("/generic:%s", rdp.Address),
		fmt.Sprintf("/user:%s", rdp.Username),
		fmt.Sprintf("/pass:%s", rdp.Password),
	)
	cmdkeyDelete := exec.Command("cmdkey",
		fmt.Sprintf("/delete:%s", rdp.Address),
	)

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

	defer func() {
		if debug {
			fmt.Println(cmdkeyDelete.String())
		}
		if err := cmdkeyDelete.Run(); err != nil {
			panic(err)
		}
	}()

	if debug {
		fmt.Println(strings.Replace(cmdkeyCreate.String(), rdp.Password, "REMOVED", 1))
	}
	if err := cmdkeyCreate.Run(); err != nil {
		return fmt.Errorf("creating credentials using cmdkey.exe: %w", err)
	}

	if debug {
		fmt.Println(startSession.String())
	}
	if err := startSession.Run(); err != nil {
		return fmt.Errorf("running rdp session using mstsc.exe: %w", err)
	}

	return nil
}
