package systemdeps

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/unit"
)

// Create and manage Systemd Unit files
// Refs:
//    https://www.freedesktop.org/software/systemd/man/systemd.unit.html
//    https://www.digitalocean.com/community/tutorials/understanding-systemd-units-and-unit-files
//    https://godoc.org/github.com/coreos/go-systemd

// Represent an entire systemd unit file.
type UnitFile struct {
	Options []*unit.UnitOption
}

// Output unit file as a string.
func (u *UnitFile) String() string {
	outReader := unit.Serialize(u.Options)
	outBytes, _ := ioutil.ReadAll(outReader)
	return fmt.Sprintf("%v", string(outBytes))
}

// Output unit file as a serialized byte array.
func (u *UnitFile) Bytes() []byte {
	outReader := unit.Serialize(u.Options)
	outBytes, _ := ioutil.ReadAll(outReader)
	return outBytes
}

// Dashify the given application string.
func formatAppName(appName string) string {
	return strings.ReplaceAll(strings.ToLower(appName), " ", "-")
}

// Create the top level application systemd unit file.
func createApplication(appName string, dryRun bool) (*UnitFile, string) {
	appFile := &UnitFile{
		Options: []*unit.UnitOption{
			{Section: "Unit", Name: "Description", Value: strings.Title(appName) + " top level service"},
			{Section: "Install", Name: "WantedBy", Value: "multi-user.target"},
		},
	}

	return appFile, formatAppName(appName) + ".target"
}

// Create the dependency drop-in file.
func createDropin(appUnitName string, p Process) *UnitFile {
	depFile := &UnitFile{}

	// Create top level dependency
	if len(appUnitName) != 0 {
		depFile.Options = append(depFile.Options, unit.NewUnitOption("Install", "WantedBy", appUnitName))
		depFile.Options = append(depFile.Options, unit.NewUnitOption("Unit", "PartOf", appUnitName))
	}

	// Create the correct order/require string
	depString := appUnitName
	depString = depString + " " + strings.Join(p.Dependencies, ".service ") + ".service"

	// Add all requirements to the unit file (if exists)
	if len(depString) != 0 {
		depFile.Options = append(depFile.Options, unit.NewUnitOption("Unit", "Wants", depString))
		depFile.Options = append(depFile.Options, unit.NewUnitOption("Unit", "After", depString))
	}

	return depFile
}

// Check if the file exists.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Compare any existing unit file with the updated one for modifications.
func unitFileChanged(unitPath string, updatedUnitFile *UnitFile) bool {
	// If file doesn't exist yet, needs to change
	if !fileExists(unitPath) {
		return true
	}

	// Read in file and parse unit options
	file, err := os.Open(unitPath)
	if err != nil {
		return true
	}
	defer file.Close()
	existingOpts, err := unit.Deserialize(file)
	if err != nil {
		return true
	}

	return !unit.AllMatch(existingOpts, updatedUnitFile.Options)
}

// Write a unit or drop-in to a file.
func updateUnitFile(systemdPath string, unitName string, unitFile *UnitFile, dryRun bool) (bool, error) {
	unitPath := systemdPath + unitName
	changed := unitFileChanged(unitPath, unitFile)

	if changed {
		if dryRun {
			// Just print out the file if doing a dry run
			fmt.Printf("##### %s #####\n", unitPath)
			fmt.Println(unitFile)
			fmt.Printf("##########\n\n")
		} else {
			// Write the unit file, overwrites the existing one if exists
			err := os.MkdirAll(systemdPath, 0755)
			if err != nil {
				return changed, err
			}

			fmt.Printf("Writing change file: %s\n", unitPath)
			err = ioutil.WriteFile(unitPath, unitFile.Bytes(), 0644)
			if err != nil {
				return changed, err
			}
		}
	} else {
		fmt.Printf("Unit file not changed: %s\n", unitPath)
	}

	return changed, nil
}

// Configure all of the systemd drop-ins based on a non-cyclic list of processes.
// If dry run flag is set, don't write/modify any files or make any systemd calls.
func ConfigureSystemd(systemdPath string, appName string, dryRun bool, processes *Processes) error {
	// Ensure path format
	if !strings.HasSuffix(systemdPath, "/") {
		systemdPath = systemdPath + "/"
	}

	// Open systemd connection
	userSystemd := strings.Contains(systemdPath, "user")
	var conn *dbus.Conn
	var err error
	if !dryRun {
		if userSystemd {
			conn, err = dbus.NewUserConnection()
			if err != nil {
				return err
			}
		} else {
			conn, err = dbus.NewSystemConnection()
			if err != nil {
				return err
			}
		}
		defer conn.Close()
	}

	// Write all new unit files, determine if systemd needs to be reloaded.
	var reloadNeeded []string
	var appFile *UnitFile
	appUnitName := ""
	if len(appName) != 0 {
		appFile, appUnitName = createApplication(appName, dryRun)
		changed, err := updateUnitFile(systemdPath, formatAppName(appUnitName), appFile, dryRun)
		if err != nil {
			return err
		}

		if changed {
			reloadNeeded = append(reloadNeeded, systemdPath+formatAppName(appUnitName))
		}
	}

	for _, p := range processes.Processes {
		depFile := createDropin(formatAppName(appUnitName), p)
		changed, err := updateUnitFile(systemdPath+p.Name+".service.d/", "dependencies.conf", depFile, dryRun)
		if err != nil {
			return err
		}

		if changed {
			reloadNeeded = append(reloadNeeded, systemdPath+p.Name+".service")
		}
	}

	// Reload all units if changed
	if len(reloadNeeded) > 0 && !dryRun {
		fmt.Println("Unit files update, reloading systemd")
		err := conn.Reload()
		if err != nil {
			return err
		}

		// Enable all changed units
		_, _, err = conn.EnableUnitFiles(reloadNeeded, false, false)
		if err != nil {
			return err
		}
	}

	return nil
}
