package configure

import (
	"fmt"
	"os"

	"github.com/danhale-git/runrdp/internal/aws"
)

type IPHost struct {
	Address string
	//port    int
}

func (h IPHost) Socket() string {
	return h.Address // :<port>
}

func (h IPHost) Credentials() Cred {
	return nil
}

type EC2Host struct {
	Private bool
	GetCred bool
	ID      string
	Profile string
	Region  string
	//Port    int
}

func (h EC2Host) Socket() string {
	sess := aws.NewSession(h.Profile, h.Region)
	instance, err := aws.InstanceFromID(sess, h.ID)

	if err != nil {
		fmt.Printf("Error Getting EC2 instance: %s ", err)
		os.Exit(1)
	}

	if h.Private {
		return *instance.PrivateIpAddress
	}

	return *instance.PublicIpAddress // :<port>
}

func (h EC2Host) Credentials() Cred {
	if h.GetCred {
		return EC2GetPassword{Host: &h}
	}

	return nil
}
