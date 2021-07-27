module github.com/timescale/tobs/cli

go 1.15

require (
	github.com/jackc/pgx/v4 v4.8.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/sergi/go-diff v1.2.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.0
	helm.sh/helm/v3 v3.6.1
	k8s.io/api v0.21.0
	k8s.io/apiextensions-apiserver v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/cli-runtime v0.21.0
	k8s.io/client-go v0.21.0
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/yaml v1.2.0
)
