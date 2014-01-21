// go-sclang/sclang.go
//
// Copyright (C) 2014 Kenichi Kanai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

// Package sclang provides a proxy class for sclang (SuperCollider client application).
package sclang

// The code is a port of ScLang.py (SuperCollider mode for gedit).
import (
	"errors"
	"io"
	"os/exec"
	"time"
)

// Sclang represents a sclang process.
type Sclang struct {
	// PathToSclang holds the parent path of the sclang command.
	PathToSclang string

	// StdoutWriter specifies the destination stream to where the process's
	// standard output will be written.
	StdoutWriter *io.Writer

	// StdinWriter holds the pipe connected to the process's standard input.
	StdinWriter io.WriteCloser

	// StdoutReader holds the pipe connected to the process's standard output.
	StdoutReader io.ReadCloser

	// Recording holds whether the sclang process is in the recoring status.
	Recording bool
}

// Start starts a sclang process and returns the Sclang struct.
func Start(pathToSclang string, stdoutWriter io.Writer) (sclang *Sclang, err error) {
	sclang = &Sclang{}
	err = sclang.Init(pathToSclang, &stdoutWriter)
	if err != nil {
		return nil, err
	}
	return sclang, nil
}

// Init initializes the Sclang struct with the specified pathToSclang and stdoutWriter.
func (sclang *Sclang) Init(pathToSclang string, stdoutWriter *io.Writer) error {
	sclang.PathToSclang = pathToSclang
	sclang.StdoutWriter = stdoutWriter

	cmd := exec.Command(pathToSclang+"sclang", "-i", "go-sclang", "-d", pathToSclang)
	var err error = nil
	sclang.StdoutReader, err = cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go io.Copy(*stdoutWriter, sclang.StdoutReader)

	sclang.StdinWriter, err = cmd.StdinPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	sclang.Recording = false
	return nil
}

// Dispose ends the sclang process.
func (sclang *Sclang) Dispose() error {
	err := sclang.StopServer()
	if err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)
	err = sclang.StdinWriter.Close()
	if err != nil {
		return err
	}

	err = sclang.StdoutReader.Close()
	if err != nil {
		return err
	}

	sclang.StdinWriter = nil
	sclang.StdoutReader = nil
	sclang.Recording = false
	return nil
}

// Evaluate makes the sclang process to evaluate the specified code.
func (sclang *Sclang) Evaluate(code string, silent bool) error {
	if sclang.StdinWriter == nil {
		return errors.New("sclang#StdinWriter is nil.")
	}

	sclang.StdinWriter.Write([]byte(code))
	if silent {
		sclang.StdinWriter.Write([]byte{0x1b})
	} else {
		sclang.StdinWriter.Write([]byte{0x0c})
	}
	return nil
}

// StartServer starts the default server (scsynth).
func (sclang *Sclang) StartServer() error {
	return sclang.Evaluate("Server.default.boot;", false)
}

// StopServer stops the default server.
func (sclang *Sclang) StopServer() error {
	return sclang.Evaluate("Server.default.quit;", false)
}

// StopSound stops the sound.
func (sclang *Sclang) StopSound() error {
	return sclang.Evaluate("thisProcess.stop;", false)
}

// ToggleRecording starts/stops recording.
func (sclang *Sclang) ToggleRecording() error {
	if sclang.Recording {
		err := sclang.Evaluate("s.stopRecording;", true)
		if err != nil {
			return err
		}
		sclang.Recording = false
	} else {
		err := sclang.Evaluate("s.prepareForRecord;", true)
		if err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond) // WORKAROUND - give server some time to prepare
		err = sclang.Evaluate("s.record;", true)
		if err != nil {
			return err
		}
		sclang.Recording = true
	}
	return nil
}

// Restart restarts the sclang process.
func (sclang *Sclang) Restart() error {
	err := sclang.Dispose()
	if err != nil {
		return err
	}
	return sclang.Init(sclang.PathToSclang, sclang.StdoutWriter)
}
