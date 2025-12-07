package executor

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Executor handles command execution
type Executor struct {
}

// NewExecutor creates a new command executor
func NewExecutor() *Executor {
	return &Executor{}
}

// Execute executes a command and returns the result
func (e *Executor) Execute(ctx context.Context, cmd string, args []string, workDir string, env map[string]string) (*Result, error) {
	startTime := time.Now()

	// Prepare command
	command := exec.CommandContext(ctx, cmd, args...)
	if workDir != "" {
		command.Dir = workDir
	}

	// Set environment variables
	if env != nil {
		envVars := os.Environ()
		for k, v := range env {
			envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
		}
		command.Env = envVars
	}

	// Capture output
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return nil, err
	}

	// Start command
	if err := command.Start(); err != nil {
		return nil, err
	}

	// Read output
	var stdoutBuf, stderrBuf strings.Builder
	stdoutDone := make(chan bool)
	stderrDone := make(chan bool)

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			stdoutBuf.WriteString(scanner.Text() + "\n")
		}
		stdoutDone <- true
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			stderrBuf.WriteString(scanner.Text() + "\n")
		}
		stderrDone <- true
	}()

	// Wait for output to finish
	<-stdoutDone
	<-stderrDone

	// Wait for command to finish
	err = command.Wait()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return nil, err
		}
	}

	executionTime := time.Since(startTime)

	return &Result{
		ExitCode:      exitCode,
		Stdout:        stdoutBuf.String(),
		Stderr:        stderrBuf.String(),
		ExecutionTime: executionTime,
	}, nil
}

// ExecuteStream executes a command and streams output
func (e *Executor) ExecuteStream(ctx context.Context, cmd string, args []string, workDir string, env map[string]string, outputChan chan<- *Output) error {
	defer close(outputChan)

	// Prepare command
	command := exec.CommandContext(ctx, cmd, args...)
	if workDir != "" {
		command.Dir = workDir
	}

	// Set environment variables
	if env != nil {
		envVars := os.Environ()
		for k, v := range env {
			envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
		}
		command.Env = envVars
	}

	// Capture output
	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return err
	}

	// Start command
	if err := command.Start(); err != nil {
		return err
	}

	// Stream stdout
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			outputChan <- &Output{
				Data:    scanner.Text() + "\n",
				IsStderr: false,
				IsEOF:   false,
			}
		}
	}()

	// Stream stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			outputChan <- &Output{
				Data:    scanner.Text() + "\n",
				IsStderr: true,
				IsEOF:   false,
			}
		}
	}()

	// Wait for command to finish
	err = command.Wait()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		}
	}

	// Send EOF
	outputChan <- &Output{
		IsEOF:    true,
		ExitCode: exitCode,
	}

	return nil
}

// ExecuteInteractive executes a command interactively
func (e *Executor) ExecuteInteractive(ctx context.Context, cmd string, args []string, workDir string, env map[string]string, input io.Reader, output io.Writer, stderr io.Writer) error {
	// Prepare command
	command := exec.CommandContext(ctx, cmd, args...)
	if workDir != "" {
		command.Dir = workDir
	}

	// Set environment variables
	if env != nil {
		envVars := os.Environ()
		for k, v := range env {
			envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
		}
		command.Env = envVars
	}

	// Connect pipes
	command.Stdin = input
	command.Stdout = output
	command.Stderr = stderr

	// Start and wait
	return command.Run()
}

// Result contains the result of command execution
type Result struct {
	ExitCode      int
	Stdout        string
	Stderr        string
	ExecutionTime time.Duration
}

// Output represents streaming output
type Output struct {
	Data     string
	IsStderr bool
	IsEOF    bool
	ExitCode int
}

