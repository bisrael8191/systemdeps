# systemdeps
Utility library that manages systemd service dependencies using drop-in
configuration files instead of modifying the installed unit file.

Service dependencies can be stored in a JSON file so that it is similar
to defining services and depends_on links in docker compose.

For example, the base services can be installed using generic Ansible roles
then independently ordered per VM based on it's needs.

Optionally, can create a top-level application to manage all underlying processes
using standard systemd commands:
  systemctl [start | stop | restart | enable | disable] application.target

References

https://www.freedesktop.org/software/systemd/man/systemd.unit.html
https://www.freedesktop.org/software/systemd/man/systemd.target.html


**Documentation**: https://godoc.org/github.com/bisrael8191/systemdeps

## Use as a library
```go
import (
	"fmt"
    "log"
    "github.com/bisrael8191/systemdeps"
)

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
}
```

## Run as application

### Build/Install
* Get dependencies: `go get`
* Install to $GOPATH/bin: `go install bin/entry.go`

### Run
* Help: `$GOBIN/entry -h`
```
Usage of entry:
  -app string
        create a top-level application to manage all services
  -config string
        path of dependency file (default "dependencies.json")
  -dryrun
        don't modify system, print out modified files
  -systemdpath string
        systemd path (default "/etc/systemd/system/")
```

* Test (output files without modifying systemd): `$GOBIN/entry -config dependencies.json -app my-test-app -dryrun`

* Dependencies with top-level app: `$GOBIN/entry -config dependencies.json -app my-test-app`

* Dependencies without top-level app: `$GOBIN/entry -config dependencies.json`

* Use user-space systemd: `$GOBIN/entry -config dependencies.json -app my-test-app -systemdpath ~/.config/systemd/user/`
  * See Arch wiki for more details: https://wiki.archlinux.org/index.php/Systemd/User#Basic_setup

