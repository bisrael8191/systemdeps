package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bisrael8191/systemdeps"
	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/unit"
)

func generateExampleUnit(processName string) *systemdeps.UnitFile {
	return &systemdeps.UnitFile{
		Options: []*unit.UnitOption{
			{Section: "Unit", Name: "Description", Value: strings.Title(processName) + " example service"},
			{Section: "Service", Name: "Type", Value: "simple"},
			{Section: "Service", Name: "ExecStart", Value: "/bin/sleep infinity"},
			{Section: "Service", Name: "Restart", Value: "on-failure"},
		},
	}
}

func main() {
	var systemdPath string
	var configFile string
	flag.StringVar(&systemdPath, "systemdpath", os.Getenv("HOME")+"/.config/systemd/user/", "systemd path")
	flag.StringVar(&configFile, "config", "dependencies.json", "path of dependency file")
	flag.Parse()

	if !strings.HasSuffix(systemdPath, "/") {
		systemdPath = systemdPath + "/"
	}

	userSystemd := strings.Contains(systemdPath, "user")

	processes, err := systemdeps.ReadDependencyFile(configFile)
	if err != nil {
		panic(err)
	}

	for _, p := range processes.Processes {
		unitName := systemdPath + p.Name + ".service"
		unitStr := generateExampleUnit(p.Name).String()

		fmt.Println("Adding unit file:", unitName)

		err := ioutil.WriteFile(unitName, []byte(unitStr), 0644)
		if err != nil {
			panic(err)
		}
	}

	// Open systemd connection
	var conn *dbus.Conn
	if userSystemd {
		conn, err = dbus.NewUserConnection()
		if err != nil {
			panic(err)
		}
	} else {
		conn, err = dbus.NewSystemConnection()
		if err != nil {
			panic(err)
		}
	}
	defer conn.Close()

	// Reload all units
	err = conn.Reload()
	if err != nil {
		panic(err)
	}
}
