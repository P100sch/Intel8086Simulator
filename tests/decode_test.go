package tests

import (
	"bytes"
	"embed"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/P100sch/Intel8086Simulator/Simulator"
)

//go:embed data/*
var testFiles embed.FS

func TestDecode(t *testing.T) {
	listing, err := testFiles.ReadDir("data")
	if err != nil {
		t.Fatal(err)
	}
	var dir string
	dir, err = os.MkdirTemp("", "Intel8086SimulatorDecodingTest")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { t.Log(os.RemoveAll(dir)) }()
	for _, fileInfo := range listing {
		var data []byte
		data, err = testFiles.ReadFile(path.Join("data", fileInfo.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var asm string
		asm, err = Simulator.Decode(data)
		asmFileName := filepath.Join(dir, fileInfo.Name()+".asm")
		err = os.WriteFile(asmFileName, []byte(asm), 0666)
		if err != nil {
			t.Fatal(err)
		}
		outputFileName := filepath.Join(dir, fileInfo.Name())
		err = exec.Command("nasm", "-f", "bin", asmFileName, "-o", outputFileName).Run()
		if err != nil {
			t.Error(err)
			continue
		}
		var outputData []byte
		outputData, err = os.ReadFile(outputFileName)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, outputData) {
			t.Error("output file does not match")
		}
	}
}
