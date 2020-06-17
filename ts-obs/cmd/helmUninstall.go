package cmd

import (
    "bytes"
    "errors"
	"fmt"
    "io"
    "os"
    "os/exec"

	"github.com/spf13/cobra"
)

// helmUninstallCmd represents the helm uninstall command
var helmUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstalls Timescale Observability",
	RunE: helmUninstall,
}

func init() {
	helmCmd.AddCommand(helmUninstallCmd)
    helmUninstallCmd.Flags().StringP("name", "n", "ts-obs", "Release name")
    helmUninstallCmd.Flags().BoolP("pvc", "", false, "Remove Persistent Volume Claims")
}

func helmUninstall(cmd *cobra.Command, args []string) error {
    var err error

    if len(args) != 0 {
        return errors.New("\"ts-obs helm uninstall\" requires 0 arguments")
    }

    var name string
    name, err = cmd.Flags().GetString("name")
    if err != nil {
        return err
    }

    var pvc bool
    pvc, err = cmd.Flags().GetBool("pvc")
    if err != nil {
        return err
    }

    var stdbuf bytes.Buffer
    mw := io.MultiWriter(os.Stdout, &stdbuf)

    uninstall := exec.Command("helm", "uninstall", name)

    uninstall.Stdout = mw
    uninstall.Stderr = mw
    fmt.Println("Uninstalling Timescale Observability")
    err = uninstall.Run()
    if err != nil {
        return err
    }

    fmt.Println("Deleting remaining artifacts")
    err = kubeDeleteService(name + "-config")
    if err != nil {
        fmt.Println(err, ", skipping")
    }
    
    err = kubeDeleteEndpoint(name)
    if err != nil {
        fmt.Println(err, ", skipping")
    }

    if !pvc {
        return nil
    }

    fmt.Println("Getting Persistent Volume Claims")
    pvcnames, err := kubeGetPVCNames(map[string]string{"release" : name})
    if err != nil {
        return err
    }

    fmt.Println("Removing Persistent Volume Claims")
    for _, s := range pvcnames {
        err = kubeDeletePVC(s)
        if err != nil {
            return err
        }
    }

    return nil
}
