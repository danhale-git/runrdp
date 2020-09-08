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

// Creates and validates a new AWS session. If region is an empty string, .aws/config region settings will be used.
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

func InstanceFromID(sess *session.Session, id string) (*ec2.Instance, error) {
	instances, err := getInstances(sess)

	if err != nil {
		return nil, err
	}

	for _, i := range instances {
		if *i.InstanceId == id {
			return &i, nil
		}
	}

	return nil, fmt.Errorf("instance with ID %s was not found", id)
}

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

func getInstances(sess *session.Session) ([]ec2.Instance, error) {
	svc := ec2.New(sess)
	response, err := svc.DescribeInstances(nil)

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
