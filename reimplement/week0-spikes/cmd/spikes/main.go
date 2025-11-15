package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	cmd := flag.String("spike", "", "Which spike to run: spike1, spike2, spike3, or all")
	environment := flag.String("env", "mock", "Environment: mock or real")
	outputDir := flag.String("output", "./results", "Output directory for results")
	verbose := flag.Bool("verbose", false, "Verbose logging")

	flag.Parse()

	if *cmd == "" || (*cmd != "spike1" && *cmd != "spike2" && *cmd != "spike3" && *cmd != "all") {
		fmt.Println("Usage: spikes -spike [spike1|spike2|spike3|all] -env [mock|real] -output ./results")
		os.Exit(1)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	startTime := time.Now()
	resultsDir := filepath.Join(wd, *outputDir)

	if *verbose {
		log.Printf("Week 0 Dependency Validation Spikes")
		log.Printf("Working directory: %s", wd)
		log.Printf("Environment: %s", *environment)
		log.Printf("Output directory: %s", resultsDir)
		log.Printf("Running: %s", *cmd)
	}

	switch *cmd {
	case "spike1":
		runSpike1(wd, *environment, resultsDir, *verbose)
	case "spike2":
		runSpike2(wd, *environment, resultsDir, *verbose)
	case "spike3":
		runSpike3(wd, *environment, resultsDir, *verbose)
	case "all":
		runAllSpikes(wd, *environment, resultsDir, *verbose)
	}

	elapsed := time.Since(startTime)
	if *verbose {
		log.Printf("All spikes completed in %v", elapsed)
	}

	fmt.Printf("Results written to: %s\n", resultsDir)
}

func runSpike1(wd, environment, outputDir string, verbose bool) {
	spikeDir := filepath.Join(wd, "spike1")
	if err := buildAndRun(spikeDir, spikeDir, environment, outputDir, "spike1", verbose); err != nil {
		log.Fatalf("Spike 1 failed: %v", err)
	}
}

func runSpike2(wd, environment, outputDir string, verbose bool) {
	spikeDir := filepath.Join(wd, "spike2")
	if err := buildAndRun(spikeDir, spikeDir, environment, outputDir, "spike2", verbose); err != nil {
		log.Fatalf("Spike 2 failed: %v", err)
	}
}

func runSpike3(wd, environment, outputDir string, verbose bool) {
	spikeDir := filepath.Join(wd, "spike3")
	if err := buildAndRun(spikeDir, spikeDir, environment, outputDir, "spike3", verbose); err != nil {
		log.Fatalf("Spike 3 failed: %v", err)
	}
}

func runAllSpikes(wd, environment, outputDir string, verbose bool) {
	spikes := []string{"spike1", "spike2", "spike3"}
	for _, spike := range spikes {
		spikeDir := filepath.Join(wd, spike)
		if err := buildAndRun(spikeDir, spikeDir, environment, outputDir, spike, verbose); err != nil {
			log.Printf("Warning: %s failed: %v", spike, err)
			// Continue with next spike instead of failing completely
		}
	}
}

func buildAndRun(projectDir, spikeDir, environment, outputDir, spikeName string, verbose bool) error {
	if verbose {
		log.Printf("[%s] Building...", spikeName)
	}

	// Build the spike binary
	binaryPath := filepath.Join(spikeDir, spikeName)
	buildCmd := exec.Command("go", "build", "-o", binaryPath, projectDir)
	buildCmd.Dir = projectDir

	if output, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("build failed: %v\n%s", err, string(output))
	}

	if verbose {
		log.Printf("[%s] Running spike...", spikeName)
	}

	// Run the spike
	runCmd := exec.Command(binaryPath,
		"-environment", environment,
		"-output", outputDir,
		"-verbose",
	)

	if output, err := runCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("execution failed: %v\n%s", err, string(output))
	}

	if verbose {
		log.Printf("[%s] Completed successfully", spikeName)
	}

	return nil
}
