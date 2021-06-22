package helm

import (
	"time"
)

// ChartSpec defines the values of a helm chart
type ChartSpec struct {
	ReleaseName string `json:"release"`
	ChartName   string `json:"chart"`
	Namespace   string `json:"namespace"`
	CreateNamespace bool `json:"createNamespace"`

	// use string instead of map[string]interface{}
	// https://github.com/kubernetes-sigs/kubebuilder/issues/528#issuecomment-466449483
	// and https://github.com/kubernetes-sigs/controller-tools/pull/317
	// +optional
	ValuesYaml string `json:"valuesYaml,omitempty"`

	// files passed using --values / -f
	ValuesFiles []string `json:"valuesFiles,omitempty"`

	// +optional
	Version string `json:"version,omitempty"`

	// +optional
	DisableHooks bool `json:"disableHooks,omitempty"`

	// +optional
	Replace bool `json:"replace,omitempty"`

	// +optional
	Wait bool `json:"wait,omitempty"`

	// +optional
	DependencyUpdate bool `json:"dependencyUpdate,omitempty"`

	// +optional
	Timeout time.Duration `json:"timeout,omitempty"`

	// +optional
	GenerateName bool `json:"generateName,omitempty"`

	// +optional
	NameTemplate string `json:"NameTemplate,omitempty"`

	// +optional
	Atomic bool `json:"atomic,omitempty"`

	// +optional
	SkipCRDs bool `json:"skipCRDs,omitempty"`

	// +optional
	UpgradeCRDs bool `json:"upgradeCRDs,omitempty"`

	// +optional
	SubNotes bool `json:"subNotes,omitempty"`

	// +optional
	Force bool `json:"force,omitempty"`

	// +optional
	ResetValues bool `json:"resetValues,omitempty"`

	// +optional
	ReuseValues bool `json:"reuseValues,omitempty"`

	// +optional
	Recreate bool `json:"recreate,omitempty"`

	// +optional
	MaxHistory int `json:"maxHistory,omitempty"`

	// +optional
	CleanupOnFail bool `json:"cleanupOnFail,omitempty"`

	// +optional
	DryRun bool `json:"dryRun,omitempty"`
}
