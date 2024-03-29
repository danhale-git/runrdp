package secretsmanager

// Use this code snippet in your app.
// If you need more information about configurations or implementing the sample code, visit the AWS docs:
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// NewSession creates and validates a new AWS session. If region is an empty string, .aws/config region settings will be
// used. A new Secrets Manager service is returned.
func NewSession(profile, region string) secretsmanageriface.SecretsManagerAPI {
	opts := session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           profile,
		Config: aws.Config{
			Region: &region,
		},
	}

	sess := session.Must(session.NewSessionWithOptions(opts))

	return secretsmanager.New(sess)
}

// Get retrieves the secret with the given key from AWS Secrets Manager.
func Get(svc secretsmanageriface.SecretsManagerAPI, secretKey string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretKey),
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		if awserror, ok := err.(awserr.Error); ok {
			return "", fmt.Errorf("getting secret from secrets manager: %s", awserror)
		}

		return "", fmt.Errorf("getting secret from secrets manager: %s", err)
	}

	var secretString, _ string
	if result.SecretString != nil {
		secretString = *result.SecretString
	} else {
		return "", fmt.Errorf("secret type of '%s' is binary, not string", secretKey)
	}

	return secretString, nil
}
