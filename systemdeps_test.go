package systemdeps

import (
	"testing"
)

func TestDAG(t *testing.T) {
	processes := Processes{}
	processes.Processes = append(processes.Processes, Process{Name: "postgresql", Dependencies: []string{"docker"}})
	processes.Processes = append(processes.Processes, Process{Name: "influxdb", Dependencies: []string{"docker"}})
	processes.Processes = append(processes.Processes, Process{Name: "process1", Dependencies: []string{"postgres", "influxdb"}})
	processes.Processes = append(processes.Processes, Process{Name: "process2", Dependencies: []string{"influxdb"}})
	processes.Processes = append(processes.Processes, Process{Name: "process3", Dependencies: []string{"process1"}})
	processes.Processes = append(processes.Processes, Process{Name: "process4", Dependencies: []string{"process3"}})
	processes.Processes = append(processes.Processes, Process{Name: "process5", Dependencies: []string{"process4"}})
	processes.Processes = append(processes.Processes, Process{Name: "process6", Dependencies: []string{"process5", "postgres", "influxdb"}})

	cycle, graph := HasCycle(&processes)
	if cycle {
		t.Errorf("Cycle found between %s and %s, no cycle expected", graph.CycleStart, graph.CycleEnd)
	}
}

func TestCycle(t *testing.T) {
	processes := Processes{}
	processes.Processes = append(processes.Processes, Process{Name: "postgresql", Dependencies: []string{"docker"}})
	processes.Processes = append(processes.Processes, Process{Name: "influxdb", Dependencies: []string{"docker"}})
	processes.Processes = append(processes.Processes, Process{Name: "process1", Dependencies: []string{"postgres", "influxdb"}})
	processes.Processes = append(processes.Processes, Process{Name: "process2", Dependencies: []string{"influxdb"}})
	processes.Processes = append(processes.Processes, Process{Name: "process3", Dependencies: []string{"process1"}})
	processes.Processes = append(processes.Processes, Process{Name: "process4", Dependencies: []string{"process3", "process6"}})
	processes.Processes = append(processes.Processes, Process{Name: "process5", Dependencies: []string{"process4"}})
	processes.Processes = append(processes.Processes, Process{Name: "process6", Dependencies: []string{"process5", "postgres", "influxdb", "process3"}})

	cycle, _ := HasCycle(&processes)
	if !cycle {
		t.Errorf("Cycle not found when expected")
	}
}

func TestSelfCycle(t *testing.T) {
	processes := Processes{}
	processes.Processes = append(processes.Processes, Process{Name: "postgresql", Dependencies: []string{"postgresql"}})

	cycle, _ := HasCycle(&processes)
	if !cycle {
		t.Errorf("Cycle not found when expected")
	}
}
