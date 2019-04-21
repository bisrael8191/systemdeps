package systemdeps

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// List of system processes.
type Processes struct {
	Processes []Process `json:"processes"`
}

// Representation of a process and it's dependencies.
type Process struct {
	Name         string   `json:"name"`
	Dependencies []string `json:"dependencies"`
}

// Read a JSON file and return the Processes struct.
func ReadDependencyFile(filename string) (*Processes, error) {
	jsonFile, err := os.Open(filepath.Clean(filename))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var processes Processes
	json.Unmarshal(byteValue, &processes)
	return &processes, nil
}
