package tests

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/P100sch/Intel8086Simulator/Simulator"
)

//go:embed data/*
var testFiles embed.FS

func TestDecode(t *testing.T) {
	dir, err := os.MkdirTemp("", "Intel8086SimulatorDecodingTest")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { t.Log(os.RemoveAll(dir)) }()

	var listing []fs.DirEntry
	listing, err = testFiles.ReadDir("data")
	if err != nil {
		t.Fatal(err)
	}
	for _, fileInfo := range listing {
		var data []byte
		data, err = testFiles.ReadFile(path.Join("data", fileInfo.Name()))
		if err != nil {
			t.Fatal(err)
		}

		var asm string
		asm, err = Simulator.Decode(data)
		if err != nil {
			t.Error(fileInfo.Name(), " ", err)
			continue
		}
		asmFileName := filepath.Join(dir, fileInfo.Name()+".asm")
		err = os.WriteFile(asmFileName, []byte(asm), 0666)
		if err != nil {
			t.Fatal(err)
		}

		outputFileName := filepath.Join(dir, fileInfo.Name())
		var outputData []byte
		command := exec.Command("nasm", "-f", "bin", asmFileName, "-o", outputFileName)
		builder := strings.Builder{}
		command.Stdout = &builder
		command.Stderr = &builder
		err = command.Run()
		if err != nil {
			t.Error(fileInfo.Name(), " assembly error\n", builder.String())
			continue
		}
		outputData, err = os.ReadFile(outputFileName)
		if err != nil {
			t.Fatal(err)
		}
		//workaround XCHG register swap by NASM
		if fileInfo.Name() == "listing_0042_completionist_decode" {
			outputData[120] = outputData[120]&0b11000000 | outputData[120]&0b00111000>>3 | outputData[120]&0b00000111<<3
			outputData[122] = outputData[122]&0b11000000 | outputData[122]&0b00111000>>3 | outputData[122]&0b00000111<<3
			outputData[124] = outputData[124]&0b11000000 | outputData[124]&0b00111000>>3 | outputData[124]&0b00000111<<3
		}

		if !bytes.Equal(data, outputData) {
			t.Error("output file does not match for " + fileInfo.Name())
			outputLen := len(outputData)
			digits := math.Floor(math.Log10(math.Max(float64(outputLen), float64(len(data))))) + 1
			for i := 0; i < len(data); i++ {
				fmt.Printf("%"+strconv.FormatFloat(digits, 'f', 0, 64)+"d: ", i)
				print(formatByte(data[i]))
				print("|")
				if i < outputLen {
					print(formatByte(outputData[i]))
					if data[i] != outputData[i] {
						print("<---")
					}
				} else {
					print("<---")
				}
				print("\n")
			}
			for i := len(data); i < outputLen; i++ {
				fmt.Printf("%03d: ", i)
				print("        |")
				print(formatByte(outputData[i]))
				print("<---\n")
			}
		}
	}
}

func formatByte(x byte) string {
	output := [8]byte{}
	for i := 0; i < 8; i++ {
		output[i] = x&(0b10000000>>i)>>(7-i) + '0'
	}
	return string(output[:])
}
