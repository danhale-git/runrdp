package aws

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/spf13/viper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type EC2Host struct {
	Private bool
	ID      string
	Profile string
	Region  string
	//Port    int
}

func (h EC2Host) Socket() string {
	session := newSession(h.Profile, h.Region)
	instance, err := instanceFromID(session, h.ID)

	if err != nil {
		fmt.Printf("Error Getting EC2 instance: %s ", err)
		os.Exit(1)
	}

	if h.Private {
		return *instance.PrivateIpAddress
	}

	return *instance.PublicIpAddress // :<port>
}

type EC2GetPassword struct {
	*EC2Host
}

func (p EC2GetPassword) Retrieve() (username, password string) {
	// Get instance password
	password, err := getPassword(
		p.Profile,
		p.Region,
		p.ID,
		viper.GetString("SSHDirectory"),
	)
	if err != nil {
		fmt.Println("Error retrieving EC2 Administrator password: ", err)
		os.Exit(1)
	}

	username = "Administrator"

	return
}

// Creates and validates a new AWS session. If region is an empty string, .aws/config region settings will be used.
func newSession(profile, region string) *session.Session {
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

func instanceFromID(sess *session.Session, id string) (*ec2.Instance, error) {
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

func getPassword(profile, region, instanceID, sshDirectory string) (string, error) {
	sess := newSession(profile, region)
	instance, err := instanceFromID(sess, instanceID)

	if err != nil {
		return "", err
	}

	svc := ec2.New(sess)

	input := ec2.GetPasswordDataInput{InstanceId: instance.InstanceId}
	output, err := svc.GetPasswordData(&input)

	if err != nil {
		return "", err
	}

	decodedPasswordData, err := base64.StdEncoding.DecodeString(*output.PasswordData)
	if err != nil {
		return "", err
	}

	b, err := rsaDecrypt(decodedPasswordData, path.Join(sshDirectory, *instance.KeyName))
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
		return nil, fmt.Errorf("parse private key: %s", err)
	}

	return rsa.DecryptPKCS1v15(nil, priv, ciphertext)
}
