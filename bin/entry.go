package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/bisrael8191/systemdeps"
)

func main() {
	var systemdPath string
	var configFile string
	var dryRun bool
	var appName string
	flag.StringVar(&systemdPath, "systemdpath", "/etc/systemd/system/", "systemd path")
	flag.StringVar(&configFile, "config", "dependencies.json", "path of dependency file")
	flag.BoolVar(&dryRun, "dryrun", false, "don't modify system, print out modified files")
	flag.StringVar(&appName, "app", "", "create a top-level application to manage all services")
	flag.Parse()

	// JSON file testing
	processes, err := systemdeps.ReadDependencyFile(configFile)
	if err != nil {
		log.Fatal(err)
		return
	}

	cycle, graph := systemdeps.HasCycle(processes)
	if cycle {
		fmt.Println("Cycle found between", graph.CycleStart, "and", graph.CycleEnd)
	} else {
		fmt.Println("No dependency cycles found, starting systemd configuration")
		err := systemdeps.ConfigureSystemd(systemdPath, appName, dryRun, processes)
		if err != nil {
			log.Fatal(err)
		}
	}
}
