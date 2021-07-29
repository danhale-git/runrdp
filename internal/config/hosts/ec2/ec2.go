package ec2

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"

	"github.com/aws/aws-sdk-go/service/ec2"
)

/*type InstanceDescriber interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}*/

// NewSession creates and validates a new AWS session. If region is an empty string, .aws/config region settings will be
// used. A new EC2 service is returned.
func NewSession(profile, region string) ec2iface.EC2API {
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
		Config: aws.Config{
			Region: &region,
		},
	}

	sess := session.Must(session.NewSessionWithOptions(opts))

	return ec2.New(sess)
}

// GetInstances gets EC2 instances from the AWS API by calling describe-instances with the given ID and filter JSON.
func GetInstances(svc ec2iface.EC2API, id, filterJSON string) ([]*ec2.Instance, error) {
	var filters []*ec2.Filter

	if filterJSON != "" {
		if err := json.Unmarshal([]byte(filterJSON), &filters); err != nil {
			return nil, fmt.Errorf("invalid filter json: %w", err)
		}
	}

	var ids []*string
	if id != "" {
		ids = []*string{&id}
	}

	out, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters:     filters,
		InstanceIds: ids,
	})
	if err != nil {
		return nil, fmt.Errorf("getting instances from aws: %w", err)
	}

	instances := make([]*ec2.Instance, 0)

	for _, r := range out.Reservations {
		instances = append(instances, r.Instances...)
	}

	return instances, nil
}

// GetTag retrieves the value of key from the instance.
func GetTag(instance *ec2.Instance, key string) *string {
	for _, tag := range instance.Tags {
		if *tag.Key == key {
			return tag.Value
		}
	}

	return nil
}

// IsRunning returns true if the instance state is 'running'.
func IsRunning(instance *ec2.Instance) bool {
	return *instance.State.Name == ec2.InstanceStateNameRunning
}

// ChooseInstance displays a picker reads a choice from the user. The inpur parameter should be os.Stdin.
func ChooseInstance(instances []*ec2.Instance, input io.Reader) (*ec2.Instance, error) {
	fmt.Println("Multiple EC2 instances:")

	for i, instance := range instances {
		n := ""
		for _, tag := range instance.Tags {
			if *tag.Key == "Name" {
				n = *tag.Value
			}
		}
		fmt.Printf("%d. %s - %s\n", i+1, n, *instance.State.Name)
	}

	fmt.Print("\nEnter number to choose: ")

	// Read int from command line input
	reader := bufio.NewReader(input)
	text, err := reader.ReadString('\n')

	if err != nil {
		return nil, fmt.Errorf("reading input: %s", err)
	}

	selected, err := strconv.Atoi(strings.Trim(text, "\r\n"))

	// User entered an invalid value, assume they want to cancel
	if err != nil || selected > len(instances) || selected == 0 {
		return nil, errors.New("no instance was chosen")
	}

	return instances[selected-1], nil
}

// GetPassword returns the initial administrator credentials for the given EC2 instance or an error if they are not
// found
func GetPassword(svc ec2iface.EC2API, instanceID string, privateKey []byte) (string, error) {
	input := ec2.GetPasswordDataInput{InstanceId: &instanceID}
	output, err := svc.GetPasswordData(&input)

	if err != nil {
		return "", fmt.Errorf("get instance password: %s", err)
	}

	decodedPasswordData, err := base64.StdEncoding.DecodeString(*output.PasswordData)
	if err != nil {
		return "", fmt.Errorf("decode password data: %s", err)
	}

	b, err := rsaDecrypt(decodedPasswordData, privateKey)
	if err != nil {
		return "", fmt.Errorf("decrypt password: %s", err)
	}

	return string(b), nil
}

// ReadPrivateKey reads the key with the given name (ignoring extension if needed) in the given directory.
func ReadPrivateKey(sshDirectory, keyName string) ([]byte, error) {
	// Read the private key
	privateKey, err := ioutil.ReadFile(path.Join(sshDirectory, keyName))
	if err != nil {
		names, err := fileNames(sshDirectory)

		if err != nil {
			return nil, err
		}

		// Try ignoring the extension
		for _, d := range names {
			noExt := d[:len(d)-len(filepath.Ext(d))]
			if noExt == keyName {
				return ioutil.ReadFile(path.Join(sshDirectory, d))
			}
		}
	} else {
		return privateKey, nil
	}

	return nil, fmt.Errorf("private key %s not found in directory %s", keyName, sshDirectory)
}

func fileNames(directory string) ([]string, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("reading SSH key directory '%s': %s", directory, err)
	}

	names := make([]string, 0)

	for _, f := range files {
		n := f.Name()
		if strings.TrimSpace(n) == "" {
			continue
		}

		names = append(names, n)
	}

	return names, nil
}

func rsaDecrypt(toDecrypt, privateKey []byte) ([]byte, error) {
	// Extract the PEM-encoded data block
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, fmt.Errorf("bad shh private key data")
	}

	if got, want := block.Type, "RSA PRIVATE KEY"; got != want {
		return nil, fmt.Errorf("unknown key type %q, want %q", got, want)
	}

	// Parse the private key
	parsedPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %s", err)
	}

	return rsa.DecryptPKCS1v15(nil, parsedPrivateKey, toDecrypt)
}
