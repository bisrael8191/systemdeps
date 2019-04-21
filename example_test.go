package systemdeps

import (
	"fmt"
	"log"
)

// Sample use of the systemdeps library.
// Creates a top-level systemd target that controls all sub-processes
// and correctly orders all of the process dependencies.
func Example() {
	// Load JSON process dependencies
	// (optional, can also create the Process structs manually, see systemdeps_test.go)
	processes, err := ReadDependencyFile("dependencies.json")
	if err != nil {
		log.Fatal(err)
		return
	}

	// Check process dependencies for cycles
	cycle, graph := HasCycle(processes)
	if cycle {
		fmt.Println("Cycle found between", graph.CycleStart, "and", graph.CycleEnd)
	} else {
		// No cycles found, create required drop-in files and configure systemd
		err := ConfigureSystemd("/etc/systemd/system", "my-test-app", true, processes)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Output:
	// ##### /etc/systemd/system/my-test-app.target #####
	// [Unit]
	// Description=My-Test-App top level service
	//
	// [Install]
	// WantedBy=multi-user.target
	//
	// ##########
	//
	// ##### /etc/systemd/system/postgresql-example.service.d/dependencies.conf #####
	// [Install]
	// WantedBy=my-test-app.target
	//
	// [Unit]
	// PartOf=my-test-app.target
	// Wants=my-test-app.target docker.service
	// After=my-test-app.target docker.service
	//
	// ##########
	//
	// ##### /etc/systemd/system/influxdb-example.service.d/dependencies.conf #####
	// [Install]
	// WantedBy=my-test-app.target
	//
	// [Unit]
	// PartOf=my-test-app.target
	// Wants=my-test-app.target docker.service
	// After=my-test-app.target docker.service
	//
	// ##########
	//
	// ##### /etc/systemd/system/process1.service.d/dependencies.conf #####
	// [Install]
	// WantedBy=my-test-app.target
	//
	// [Unit]
	// PartOf=my-test-app.target
	// Wants=my-test-app.target postgresql-example.service influxdb-example.service
	// After=my-test-app.target postgresql-example.service influxdb-example.service
	//
	// ##########
	//
	// ##### /etc/systemd/system/process2.service.d/dependencies.conf #####
	// [Install]
	// WantedBy=my-test-app.target
	//
	// [Unit]
	// PartOf=my-test-app.target
	// Wants=my-test-app.target postgresql.service
	// After=my-test-app.target postgresql.service
	//
	// ##########
	//
	// ##### /etc/systemd/system/process3.service.d/dependencies.conf #####
	// [Install]
	// WantedBy=my-test-app.target
	//
	// [Unit]
	// PartOf=my-test-app.target
	// Wants=my-test-app.target process1.service
	// After=my-test-app.target process1.service
	//
	// ##########
	//
	// ##### /etc/systemd/system/process4.service.d/dependencies.conf #####
	// [Install]
	// WantedBy=my-test-app.target
	//
	// [Unit]
	// PartOf=my-test-app.target
	// Wants=my-test-app.target process3.service
	// After=my-test-app.target process3.service
	//
	// ##########
	//
	// ##### /etc/systemd/system/process5.service.d/dependencies.conf #####
	// [Install]
	// WantedBy=my-test-app.target
	//
	// [Unit]
	// PartOf=my-test-app.target
	// Wants=my-test-app.target process4.service
	// After=my-test-app.target process4.service
	//
	// ##########
	//
	// ##### /etc/systemd/system/process6.service.d/dependencies.conf #####
	// [Install]
	// WantedBy=my-test-app.target
	//
	// [Unit]
	// PartOf=my-test-app.target
	// Wants=my-test-app.target process5.service postgresql-example.service influxdb-example.service
	// After=my-test-app.target process5.service postgresql-example.service influxdb-example.service
	//
	// ##########
}
