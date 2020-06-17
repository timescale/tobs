package tests

import (
    "fmt"
    "os/exec"
    "testing"
    "time"
)

func testInstall(t testing.TB, name string, filename string) {
    var install *exec.Cmd

    if name == ""  && filename == "" {
        fmt.Println("Running 'ts-obs install'")
        install = exec.Command("/home/archen/go/bin/ts-obs", "install")
    } else if name == "" {
        fmt.Printf("Running 'ts-obs install -f %v'\n", filename)
        install = exec.Command("/home/archen/go/bin/ts-obs", "install", "-f", filename)
    } else if filename == "" {
        fmt.Printf("Running 'ts-obs install -n %v'\n", name)
        install = exec.Command("/home/archen/go/bin/ts-obs", "install", "-n", name)
    } else {
        fmt.Printf("Running 'ts-obs install -n %v -f %v'\n", name, filename)
        install = exec.Command("/home/archen/go/bin/ts-obs", "install", "-n", name, "-f", filename)
    }

    out, err := install.CombinedOutput()
    if err != nil {
        fmt.Println(string(out))
        t.Fatal(err)
    }
}

func testHelmInstall(t testing.TB, name string, filename string) {
    var install *exec.Cmd

    if name == ""  && filename == "" {
        fmt.Println("Running 'ts-obs helm install'")
        install = exec.Command("/home/archen/go/bin/ts-obs", "helm", "install")
    } else if name == "" {
        fmt.Printf("Running 'ts-obs helm install -f %v'\n", filename)
        install = exec.Command("/home/archen/go/bin/ts-obs", "helm", "install", "-f", filename)
    } else if filename == "" {
        fmt.Printf("Running 'ts-obs helm install -n %v'\n", name)
        install = exec.Command("/home/archen/go/bin/ts-obs", "helm", "install", "-n", name)
    } else {
        fmt.Printf("Running 'ts-obs helm install -n %v -f %v'\n", name, filename)
        install = exec.Command("/home/archen/go/bin/ts-obs", "helm", "install", "-n", name, "-f", filename)
    }
    
    out, err := install.CombinedOutput()
    if err != nil {
        fmt.Println(string(out))
        t.Fatal(err)
    }
}

func testUninstall(t testing.TB, name string) {
    var uninstall *exec.Cmd

    if name == ""  {
        fmt.Println("Running 'ts-obs uninstall'")
        uninstall = exec.Command("/home/archen/go/bin/ts-obs", "uninstall")
    } else {
        fmt.Printf("Running 'ts-obs uninstall -n %v'\n", name)
        uninstall = exec.Command("/home/archen/go/bin/ts-obs", "uninstall", "-n", name)
    }

    out, err := uninstall.CombinedOutput()
    if err != nil {
        fmt.Println(string(out))
        t.Fatal(err)
    }
}

func testHelmUninstall(t testing.TB, name string) {
    var uninstall *exec.Cmd

    if name == ""  {
        fmt.Println("Running 'ts-obs helm uninstall'")
        uninstall = exec.Command("/home/archen/go/bin/ts-obs", "uninstall")
    } else {
        fmt.Printf("Running 'ts-obs helm uninstall -n %v'\n", name)
        uninstall = exec.Command("/home/archen/go/bin/ts-obs", "uninstall", "-n", name)
    }

    out, err := uninstall.CombinedOutput()
    if err != nil {
        fmt.Println(string(out))
        t.Fatal(err)
    }
}

func testHelmGetYaml(t testing.TB) {
    var getyaml *exec.Cmd

    fmt.Println("Running 'ts-obs helm get-yaml'")
    getyaml = exec.Command("/home/archen/go/bin/ts-obs", "helm", "get-yaml")

    out, err := getyaml.CombinedOutput()
    if err != nil {
        fmt.Println(string(out))
        t.Fatal(err)
    }
}

func TestInstallation(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping installation tests")
    }

    testHelmGetYaml(t)

    testUninstall(t, "")
    testInstall(t, "", "")
    testHelmUninstall(t, "")
    testHelmInstall(t, "", "")
    testUninstall(t, "")
    testHelmInstall(t, "", "")
    testHelmUninstall(t, "")
    testInstall(t, "", "")
    testUninstall(t, "")

    testInstall(t, "sd-fo9ods-oe93", "")
    testHelmUninstall(t, "sd-fo9ods-oe93")
    testHelmInstall(t, "x98-2cn4-ru2-9cn48u", "")
    testUninstall(t, "x98-2cn4-ru2-9cn48u")
    testHelmInstall(t, "as-dn-in234i-n", "")
    testHelmUninstall(t, "as-dn-in234i-n")
    testInstall(t, "we-3oiwo3o-s-d", "")
    testUninstall(t, "we-3oiwo3o-s-d")

    testInstall(t, "f1", "./testdata/f1.yml")
    testHelmUninstall(t, "f1")
    testHelmInstall(t, "f2", "./testdata/f2.yml")
    testUninstall(t, "f2")
    testHelmInstall(t, "f3", "./testdata/f3.yml")
    testHelmUninstall(t, "f3")
    testInstall(t, "f4", "./testdata/f4.yml")
    testUninstall(t, "f4")

    testInstall(t, "", "")
    fmt.Println("Sleep to wait for pods to initialize")
    time.Sleep(300 * time.Second)
}
