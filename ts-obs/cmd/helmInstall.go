package cmd

import (
    "errors"
	"fmt"
    "io"
    "os"
    "os/exec"

	"github.com/spf13/cobra"
)

// helmInstallCmd represents the helm install command
var helmInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs Timescale Observability",
	RunE: helmInstall,
}

func init() {
	helmCmd.AddCommand(helmInstallCmd)
    helmInstallCmd.Flags().StringP("name", "n", "ts-obs", "Release name")
    helmInstallCmd.Flags().StringP("filename", "f", "", "YAML configuration file to load")
}

func helmInstall(cmd *cobra.Command, args []string) error {
    var err error

    if len(args) != 0 {
        return errors.New("\"ts-obs helm install\" requires 0 arguments")
    }

    var name string
    name, err = cmd.Flags().GetString("name")
    if err != nil {
        return err
    }

    var file string
    file, err = cmd.Flags().GetString("filename")
    if err != nil {
        return err
    }

    w := io.Writer(os.Stdout)

    addchart := exec.Command("helm", "repo", "add", "timescale", "https://charts.timescale.com")

    addchart.Stdout = w
    addchart.Stderr = w
    fmt.Println("Adding Timescale Helm chart")
    err = addchart.Run()
    if err != nil {
        return err
    }

    update := exec.Command("helm", "repo", "update")

    update.Stdout = w
    update.Stderr = w
    fmt.Println("Updating local chart info")
    err = update.Run()
    if err != nil {
        return err
    }

    var install *exec.Cmd
    if file == "" {
        install = exec.Command("helm", "install", name, "timescale/timescale-observability", "--devel")
    } else {
        install = exec.Command("helm", "upgrade", "--install", name, "--values", file, "timescale/timescale-observability", "--devel")
    }

    install.Stdout = w
    install.Stderr = w
    fmt.Println("Installing Timescale Observability")
    err = install.Run()
    if err != nil {
        return err
    }

    return nil
}
