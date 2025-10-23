package main_test

import (
	"os/exec"
	"testing"
)

func TestMainExecution(t *testing.T) {
	cmd := exec.Command("go", "build", "-o", "/tmp/test-metrics-sidecard", "./cmd/server")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build main program: %v", err)
	}
}
