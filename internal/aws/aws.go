package aws

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// NewSession creates and validates a new AWS session. If region is an empty string, .aws/config region settings will be
// used.
func NewSession(profile, region string) *session.Session {
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
		Config: aws.Config{
			Region: &region,
		},
	}

	sess := session.Must(session.NewSessionWithOptions(opts))

	return sess
}

// InstanceFromID returns the details of the AWS EC2 instance with the given ID or an error if it isn't found.
func InstanceFromID(sess *session.Session, id string) (*ec2.Instance, error) {
	name := "instance-id"
	filters := []*ec2.Filter{{Name: &name, Values: []*string{&id}}}
	instances, err := getInstances(sess, &ec2.DescribeInstancesInput{Filters: filters})

	if err != nil {
		return nil, err
	}

	for _, inst := range instances {
		if *inst.InstanceId == id {
			//fmt.Printf("%+v\n", inst)
			return &inst, nil
		}
	}

	return nil, fmt.Errorf("instance with ID %s was not found", id)
}

// InstanceFromTagFilter attempts to find a single instance based on the given tag filters. If more than one or 0
// instances are found it returns an error.
func InstanceFromTagFilter(sess *session.Session, include, exclude []string) (*ec2.Instance, error) {
	filters := make([]*ec2.Filter, len(include))
	for i := 0; i < len(filters); i++ {
		split := strings.Split(include[i], ":")

		var name string

		var values []*string

		switch len(split) {
		// tag:<key> - The key/value combination of a tag assigned to the resource.
		// Use the tag key in the filter name and the tag value as the filter value.
		case 2:
			name = fmt.Sprintf("tag:%s", split[0])
			values = []*string{&split[1]}
		// tag-key - The key of a tag assigned to the resource. Tag value is the value.
		case 1:
			name = "tag-key"
			values = []*string{&split[0]}

		default:
			return nil, fmt.Errorf("item %d in includetags filter contains more than one colon (:)", i)
		}

		filters[i] = &ec2.Filter{
			Name:   &name,
			Values: values,
		}
	}

	instances, err := getInstances(sess, &ec2.DescribeInstancesInput{Filters: filters})

	if err != nil {
		return nil, err
	}

	eligible := make([]ec2.Instance, 0)

	for _, inst := range instances {
		if inst.Tags == nil {
			continue
		}

		if containsTag(inst.Tags, include) && !containsTag(inst.Tags, exclude) {
			eligible = append(eligible, inst)
		}
	}

	if len(eligible) > 1 {
		return nil, fmt.Errorf("%d instances found with given tags, must identify a single instance", len(eligible))
	} else if len(eligible) == 0 {
		return nil, fmt.Errorf("no instances found with matching tags")
	} else {
		return &eligible[0], nil
	}
}

func containsTag(instanceTags []*ec2.Tag, givenTags []string) bool {
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
func GetPassword(profile, region, instanceID, sshDirectory string) (string, error) {
	sess := NewSession(profile, region)
	instance, err := InstanceFromID(sess, instanceID)

	if err != nil {
		return "", fmt.Errorf("get instance from id: %s", err)
	}

	svc := ec2.New(sess)

	input := ec2.GetPasswordDataInput{InstanceId: instance.InstanceId}
	output, err := svc.GetPasswordData(&input)

	if err != nil {
		return "", fmt.Errorf("get instance password: %s", err)
	}

	decodedPasswordData, err := base64.StdEncoding.DecodeString(*output.PasswordData)
	if err != nil {
		return "", fmt.Errorf("decode password data: %s", err)
	}

	// Read the private key
	privateKey, err := ioutil.ReadFile(path.Join(sshDirectory, *instance.KeyName))
	if err != nil {
		// Try ignoring the extension
		for _, d := range fileNames(sshDirectory) {
			noExt := d[:len(d)-len(filepath.Ext(d))]
			if noExt == *instance.KeyName {
				privateKey, err = ioutil.ReadFile(path.Join(sshDirectory, d))
				if err != nil {
					return "", fmt.Errorf("read key file: %s", err)
				}
			}
		}
	}

	b, err := rsaDecrypt(decodedPasswordData, privateKey)
	if err != nil {
		return "", fmt.Errorf("decrypt password: %s", err)
	}

	return string(b), nil
}

func fileNames(directory string) []string {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		log.Fatal(err)
	}

	names := make([]string, 0)

	for _, f := range files {
		n := f.Name()
		if strings.TrimSpace(n) == "" {
			continue
		}

		names = append(names, n)
	}

	return names
}

func getInstances(sess *session.Session, input *ec2.DescribeInstancesInput) ([]ec2.Instance, error) {
	svc := ec2.New(sess)
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
		return nil, fmt.Errorf("bad key data: %s", "not PEM-encoded")
	}

	if got, want := block.Type, "RSA PRIVATE KEY"; got != want {
		return nil, fmt.Errorf("unknown key type %q, want %q", got, want)
	}

	// Parse the private key
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %s", err)
	}

	return rsa.DecryptPKCS1v15(nil, priv, toDecrypt)
}
