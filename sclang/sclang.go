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
package sclang

import (
	"errors"
	"io"
	"os/exec"
	"time"
)

type Sclang struct {
	PathToSclang string
	StdoutWriter *io.Writer
	StdinWriter  io.WriteCloser
	StdoutReader io.ReadCloser
	Recording    bool
}

func Start(pathToSclang string, stdoutWriter *io.Writer) (sclang *Sclang, err error) {
	sclang = &Sclang{}
	err = sclang.Init(pathToSclang, stdoutWriter)
	if err != nil {
		return nil, err
	}
	return sclang, nil
}

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

func (sclang *Sclang) Evaluate(code string, silent bool) error {
	if sclang.StdinWriter == nil {
		return errors.New("sclang#StdinWriter is nil.")
	}

	sclang.StdinWriter.Write([]byte(code))
	if silent {
		sclang.StdinWriter.Write([]byte("\x1b"))
	} else {
		sclang.StdinWriter.Write([]byte("\x0c"))
	}
	return nil
}

func (sclang *Sclang) StartServer() error {
	return sclang.Evaluate("Server.default.boot;", false)
}

func (sclang *Sclang) StopServer() error {
	return sclang.Evaluate("Server.default.quit;", false)
}

func (sclang *Sclang) StopSound() error {
	return sclang.Evaluate("thisProcess.stop;", false)
}

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
		time.Sleep(100 * time.Millisecond)
		err = sclang.Evaluate("s.record;", true)
		if err != nil {
			return err
		}
		sclang.Recording = true
	}
	return nil
}

func (sclang *Sclang) Restart() error {
	err := sclang.Dispose()
	if err != nil {
		return err
	}
	return sclang.Init(sclang.PathToSclang, sclang.StdoutWriter)
}
