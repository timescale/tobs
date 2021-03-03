package timescaledb_secrets

import (
	"fmt"

	"github.com/timescale/tobs/cli/pkg/k8s"
)

func DeleteTimescaleDBSecrets(releaseName, namespace string, backUpEnabled bool) {
	fmt.Println("Deleting TimescaleDB secrets...")
	credentialsSecret := []string{releaseName + "-credentials", releaseName + "-certificate"}
	if backUpEnabled {
		credentialsSecret = append(credentialsSecret, releaseName+"-pgbackrest")
	}
	for _, s := range credentialsSecret {
		err := k8s.DeleteSecret(s, namespace)
		if err != nil {
			fmt.Printf("failed to delete %s secret %v\n", s, err)
		}
	}
}
