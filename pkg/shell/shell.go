package shell

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	pty "github.com/kr/pty"
)

type Shell struct {
	BootCommand *exec.Cmd
	StoragePath string
	TTY         *os.File
	ExitSignal  chan string
}

func NewShell(bootCommand *exec.Cmd, storagePath string) (*Shell, error) {
	exitChannel := make(chan string, 1)

	return &Shell{
		BootCommand: bootCommand,
		StoragePath: storagePath,
		ExitSignal:  exitChannel,
	}, nil
}

func (s *Shell) Start() error {
	log.Printf("Starting stateful shell")

	tty, err := pty.Start(s.BootCommand)
	if err != nil {
		log.Printf("Failed to start stateful shell")
		return err
	}

	s.TTY = tty

	s.handleAbruptShellCloses()

	time.Sleep(1000)

	return s.silencePromptAndDisablePS1()
}

func (s *Shell) handleAbruptShellCloses() {
	//
	// If the Shell is abrupty closed, we are cleaning up, and sending out an
	// exit signal.
	//
	// Abrupt closes can be caused by:
	//
	//  - running exit 1 command
	//  - setting up set -e
	//  - setting up set -pipefail
	//  - killing the shell with kill <pid>
	//
	go func() {
		err := s.BootCommand.Wait()

		msg := "no exit message"
		if err != nil {
			msg = err.Error()
		}

		log.Printf("Shell unexpectedly closed with %s. Closing associated TTY.", msg)
		s.TTY.Close()

		log.Printf("Publishing an exit signal.")
		s.ExitSignal <- msg
	}()
}

func (s *Shell) Read(buffer *([]byte)) (int, error) {
	done := make(chan bool, 1)

	var n int
	var err error

	go func() {
		n, err = s.TTY.Read(*buffer)
		done <- true
	}()

	select {
	case <-done:
		return n, err
	case <-s.ExitSignal:
		return 0, fmt.Errorf("Shell Closed")
	}
}

func (s *Shell) Write(instruction string) (int, error) {
	log.Printf("Sending Instruction: %s", instruction)

	done := make(chan bool, 1)

	var n int
	var err error

	go func() {
		n, err = s.TTY.Write([]byte(instruction + "\n"))
		done <- true
	}()

	select {
	case <-done:
		return n, err
	case <-s.ExitSignal:
		return 0, fmt.Errorf("Shell Closed")
	}
}

func (s *Shell) silencePromptAndDisablePS1() error {
	everythingIsReadyMark := "87d140552e404df69f6472729d2b2c3"

	s.TTY.Write([]byte("export PS1=''\n"))
	s.TTY.Write([]byte("stty -echo\n"))
	s.TTY.Write([]byte("echo stty `stty -g` > /tmp/restore-tty\n"))
	s.TTY.Write([]byte("cd ~\n"))
	s.TTY.Write([]byte("echo '" + everythingIsReadyMark + "'\n"))

	stdoutScanner := bufio.NewScanner(s.TTY)

	//
	// At this point, the terminal is still echoing the output back to stdout
	// we ignore the entered command, and look for the magic mark in the output
	//
	// Example content of output before ready mark:
	//
	//   export PS1=''
	//   stty -echo
	//   echo + '87d140552e404df69f6472729d2b2c3'
	//   vagrant@boxbox:~/code/agent/pkg/executors/shell$ export PS1=''
	//   stty -echo
	//   echo '87d140552e404df69f6472729d2b2c3'
	//

	// We wait until marker is displayed in the output

	log.Println("Waiting for initialization")

	for stdoutScanner.Scan() {
		text := stdoutScanner.Text()

		log.Printf("(tty) %s\n", text)

		if strings.Contains(text, "executable file not found") {
			return fmt.Errorf(text)
		}

		if !strings.Contains(text, "echo") && strings.Contains(text, everythingIsReadyMark) {
			break
		}
	}

	return nil
}

func (s *Shell) NewProcess(command string) *Process {
	return NewProcess(command, s.StoragePath, s)
}

func (s *Shell) Close() error {
	err := s.TTY.Close()
	if err != nil {
		log.Printf("Closing the TTY returned an error")
		return err
	}

	err = s.BootCommand.Process.Kill()
	if err != nil {
		log.Printf("Process killing procedure returned an erorr %+v\n", err)
		return err
	}

	return nil
}
