package timescaledb_secrets

import (
	"fmt"
	"os"

	"github.com/timescale/tobs/cli/pkg/k8s"
)

func CreateTimescaleDBSecrets(releaseName, namespace string, enableS3Backup bool) error {
	// Previous helm install used to create namespace if it doesn't exist
	// but as we are creating secrets prior to deploying tobs. We are verifying
	// namespace if doesn't create one.
	err := k8s.VerifyNamespaceIfNotCreate(namespace)
	if err != nil {
		return err
	}

	err = k8s.CreateTimescaleDBCertificates(releaseName, namespace)
	if err != nil {
		return err
	}

	err = k8s.CreateTimescaleDBCredentials(releaseName, namespace)
	if err != nil {
		return err
	}

	if enableS3Backup {
		err = createS3BackupForTimescaleDB(releaseName, namespace)
		if err != nil {
			return err
		}
	}

	return nil
}

func createS3BackupForTimescaleDB(releaseName, namespace string) error {
	s3 := k8s.S3Details{}
	fmt.Print("We'll be asking a few questions about S3 buckets, keys, secrets and endpoints.\n\nFor background information, visit these pages:\n\nAmazon Web Services:\n- https://docs.aws.amazon.com/AmazonS3/latest/gsg/CreatingABucket.html\n- https://docs.aws.amazon.com/IAM/latest/UserGuide/introduction.html\n\nDigital Ocean:\n- https://developers.digitalocean.com/documentation/spaces/#aws-s3-compatibility\n\nGoogle Cloud:\n- https://cloud.google.com/storage/docs/migrating#migration-simple\n\n")

	// if values are available in env variables pick up them else
	// read from the user input.
	fmt.Println("What is the name of the S3 bucket?")
	s3.BucketName = os.Getenv("PGBACKREST_REPO1_S3_BUCKET")
	if s3.BucketName == "" {
		fmt.Scanln(&s3.BucketName)
	}

	fmt.Println("\nWhat is the name of the S3 endpoint? (leave blank for default)")
	s3.EndpointName = os.Getenv("PGBACKREST_REPO1_S3_ENDPOINT")
	if s3.EndpointName == "" {
		fmt.Scanln(&s3.EndpointName)
	}

	fmt.Println("\nWhat is the region of the S3 endpoint? (leave blank for default)")
	s3.EndpointRegion = os.Getenv("PGBACKREST_REPO1_S3_REGION")
	if s3.EndpointRegion == "" {
		fmt.Scanln(&s3.EndpointRegion)
	}

	fmt.Println("\nWhat is the S3 Key to use?")
	s3.Key = os.Getenv("PGBACKREST_REPO1_S3_KEY")
	if s3.Key == "" {
		fmt.Scanln(&s3.Key)
	}

	fmt.Println("\nWhat is the S3 Secret to use?")
	s3.Secret = os.Getenv("PGBACKREST_REPO1_S3_KEY_SECRET")
	if s3.Secret == "" {
		fmt.Scanln(&s3.Secret)
	}

	fmt.Println()
	err := k8s.CreateTimescaleDBPgBackRest(releaseName, namespace, s3)
	if err != nil {
		return err
	}

	return nil
}
