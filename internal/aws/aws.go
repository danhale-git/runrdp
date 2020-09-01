package aws

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"io/ioutil"
	"log"
)

// Creates and validates a new AWS session. If region is an empty string, .aws/config region settings will be used.
func NewSession(profile, region string) *session.Session {
	return session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
		Config: aws.Config{
			Region: &region,
		},
	}))

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

	return nil, errors.New(fmt.Sprintf("instance with ID %s was not found", id))
}

func GetPassword(sess *session.Session, instanceID, keyFilePath string) (string, error) {
	svc := ec2.New(sess)

	input := ec2.GetPasswordDataInput{InstanceId: &instanceID}
	output, err := svc.GetPasswordData(&input)
	if err != nil {
		return "", err
	}

	decodedPasswordData, err := base64.StdEncoding.DecodeString(*output.PasswordData)
	if err != nil {
		return "", err
	}

	b, err := rsaDecrypt(decodedPasswordData, keyFilePath)
	if err != nil {
		return "", err
	}

	return string(b), nil
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

func rsaDecrypt(ciphertext []byte, keyFile string) ([]byte, error) {
	// Read the private key
	privateKey, err := ioutil.ReadFile(keyFile)
	if err != nil {
		log.Fatalf("read key file: %s", err)
	}

	// Extract the PEM-encoded data block
	block, _ := pem.Decode(privateKey)
	if block == nil {
		log.Fatalf("bad key data: %s", "not PEM-encoded")
	}
	if got, want := block.Type, "RSA PRIVATE KEY"; got != want {
		log.Fatalf("unknown key type %q, want %q", got, want)
	}

	// Parse the private key
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("parse private key: %s", err))
	}

	return rsa.DecryptPKCS1v15(nil, priv, ciphertext)
}
