package tests

import (
	"fmt"
	"log"
	"math"
	"path"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/P100sch/Intel8086Simulator/Simulation"
)

type stripClockWriter strings.Builder

func (s *stripClockWriter) Write(p []byte) (n int, err error) {
	position := strings.LastIndex(string(p), "bytes")
	return (*strings.Builder)(s).Write(slices.Concat(p[:position+5], []byte{'\n'}[:]))
}

func TestSimulate(t *testing.T) {
	listing, err := testFiles.ReadDir("data/simulation")
	if err != nil {
		t.Fatal(err)
	}
	for _, fileInfo := range listing {
		if fileInfo.IsDir() {
			continue
		}
		var data []byte
		data, err = testFiles.ReadFile(path.Join("data/simulation", fileInfo.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var expectedOutput []byte
		expectedOutput, err = testFiles.ReadFile(path.Join("data/simulation/outputs", fileInfo.Name()+".txt"))
		if err != nil {
			t.Fatal(err)
		}

		builder := strings.Builder{}
		logger := log.Logger{}
		if isNotClocked(fileInfo.Name(), t) {
			logger.SetOutput((*stripClockWriter)(&builder))
		} else {
			logger.SetOutput(&builder)
		}
		err = Simulation.LoadProgram(data, true)
		if err != nil {
			log.Fatal(err)
		}
		err = Simulation.Simulate(&logger)
		output := builder.String()
		if err != nil {
			print(output)
			t.Error(fileInfo.Name(), " ", err)
			continue
		}
		expected := strings.ReplaceAll(string(expectedOutput), "\r", "")

		if expected != output {
			firstError := true
			t.Error("output does not match for " + fileInfo.Name())
			outputLines := strings.Split(output, "\n")
			outputLen := len(outputLines)
			expectedOutputLines := strings.Split(expected, "\n")
			digits := math.Floor(math.Log10(math.Max(float64(outputLen), float64(len(expectedOutputLines))))) + 1
			for i := 0; i < len(expectedOutputLines); i++ {
				if expectedOutputLines[i] != outputLines[i] {
					fmt.Printf("%"+strconv.FormatFloat(digits, 'f', 0, 64)+"d: %s\n", i, expectedOutputLines[i])
					fmt.Printf("%"+strconv.FormatFloat(digits, 'f', 0, 64)+"d: %s\n\n", i, outputLines[i])
				}
			}
			for i := len(expectedOutputLines); i < outputLen; i++ {
				if firstError {
					firstError = false
				}
				fmt.Printf("%"+strconv.FormatFloat(digits, 'f', 0, 64)+"d: %s\n\n", i, outputLines[i])
			}
		}
		Simulation.Rest()
	}
}

func isNotClocked(fileName string, t *testing.T) bool {
	splitName := strings.Split(fileName, "_")
	if len(splitName) < 2 {
		t.Fatal("invalid file name \"" + fileName + "\"")
	}
	value, err := strconv.Atoi(splitName[1])
	if err != nil {
		t.Fatal(err)
	}
	return value < 56
}
