package ec2instances

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Tags is tag filter data from the ec2 host config.
type Tags struct {
	IncludeTags []string
	ExcludeTags []string
	Separator   string
}

// IncludeFilter returns the AWS filters for this configured tags.
func (t *Tags) IncludeFilter() ([]*ec2.Filter, error) {
	if t.Separator == "" {
		return nil, fmt.Errorf("key/value separator is empty string")
	}

	filters := make([]*ec2.Filter, 0)

	for i := 0; i < len(t.IncludeTags); i++ {
		split := strings.Split(t.IncludeTags[i], t.Separator)

		var name string

		var value *string

		switch len(split) {
		// tag:<key> - The key/value combination of a tag assigned to the resource.
		// Use the tag key in the filter name and the tag value as the filter value.
		case 2: //nolint:gomnd
			name = fmt.Sprintf("tag:%s", split[0])
			value = &split[1]
		// tag-key - The key of a tag assigned to the resource. Tag value is the value.
		case 1:
			name = "tag-key"
			value = &split[0]

		default:
			return nil, fmt.Errorf("'%s' in includetags filter contains more than one separator (%s)",
				t.IncludeTags[i], t.Separator)
		}

		if exists, index := t.filterExists(name, filters); exists {
			// Key already exists
			filters[index].Values = append(filters[index].Values, value)
		} else {
			// Key is new
			newFilter := &ec2.Filter{
				Name:   &name,
				Values: []*string{value},
			}

			filters = append(filters, newFilter)
		}
	}

	return filters, nil
}

func (t *Tags) filterExists(name string, filters []*ec2.Filter) (exists bool, index int) {
	for i, filter := range filters {
		if *filter.Name == name {
			exists = true
			index = i

			return
		}
	}

	exists = false

	return
}

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

// GetInstance abstracts retrieving an EC2 instance from the AWS API. If an ID is provided, that instance will be
// requested, otherwise the given tag filter will be used to search for the instance.
func GetInstance(svc ec2iface.EC2API, id, tagSeparator string, include, exclude []string) (*ec2.Instance, error) {
	switch {
	case id != "":
		return InstanceFromID(svc, id)
	case len(include) > 0:
		filter := Tags{
			IncludeTags: include,
			ExcludeTags: exclude,
			Separator:   tagSeparator,
		}

		return InstanceFromTagFilter(svc, filter)
	default:
		return nil, fmt.Errorf("no includetags or id field in config, must provide one")
	}
}

// InstanceFromID returns the details of the AWS EC2 instance with the given ID or an error if it isn't found.
func InstanceFromID(svc ec2iface.EC2API, id string) (*ec2.Instance, error) {
	filters := []*ec2.Filter{{Name: aws.String("instance-id"), Values: []*string{&id}}}
	instances, err := getInstances(svc, &ec2.DescribeInstancesInput{Filters: filters})

	if err != nil {
		return nil, err
	}

	switch len(instances) {
	case 0:
		return nil, fmt.Errorf("instance with ID %s was not found", id)
	case 1:
		return &instances[0], nil
	default:
		return nil, fmt.Errorf("%d instances were found when filtering for instance id %s", len(instances), id)
	}
}

// InstanceFromTagFilter attempts to find a single instance based on the given tag filters. If more than one or 0
// instances are found it returns an error.
func InstanceFromTagFilter(svc ec2iface.EC2API, tags Tags) (*ec2.Instance, error) {
	filters, err := tags.IncludeFilter()

	if err != nil {
		return nil, err
	}

	instances, err := getInstances(svc, &ec2.DescribeInstancesInput{Filters: filters})

	if err != nil {
		return nil, err
	}

	eligible := make([]ec2.Instance, 0)

	for _, inst := range instances {
		if inst.Tags == nil {
			continue
		}

		if !containsOneOfTags(inst.Tags, tags.ExcludeTags) {
			eligible = append(eligible, inst)
		}
	}

	switch {
	case len(eligible) > 1:
		fmt.Println("Multiple EC2 instances found with given tags:")

		for i := range eligible {
			fmt.Printf("%d. %s\n", i+1, instanceName(&eligible[i]))
		}

		fmt.Print("\nEnter number to connect: ")

		// Read int from command line input
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')

		if err != nil {
			panic(err)
		}

		selected, err := strconv.Atoi(strings.Trim(text, "\r\n"))

		// User entered an invalid value, assume they want to cancel
		if err != nil || selected > len(eligible) || selected == 0 {
			return nil, errors.New("no instance was chosen")
		}

		return &eligible[selected-1], nil
	case len(eligible) == 0:
		return nil, fmt.Errorf("no instances found with matching tags")
	default:
		return &eligible[0], nil
	}
}

func instanceName(instance *ec2.Instance) string {
	for _, tag := range instance.Tags {
		if *tag.Key == "Name" {
			return *tag.Value
		}
	}

	return ""
}

func containsOneOfTags(instanceTags []*ec2.Tag, givenTags []string) bool {
	for _, t := range instanceTags {
		for _, s := range givenTags {
			split := strings.Split(s, ":")
			if *t.Key == split[0] &&
				(len(split) == 1 || *t.Value == split[1]) {
				return true
			}
		}
	}

	return false
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

func getInstances(svc ec2iface.EC2API, input *ec2.DescribeInstancesInput) ([]ec2.Instance, error) {
	response, err := svc.DescribeInstances(input)

	if err != nil {
		return nil, err
	}

	var instances []ec2.Instance

	for _, r := range response.Reservations {
		for _, i := range r.Instances {
			instances = append(instances, *i)
		}
	}

	return instances, nil
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
