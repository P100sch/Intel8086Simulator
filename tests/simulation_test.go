package tests

import (
  "fmt"
  "log"
  "math"
  "path"
  "strconv"
  "strings"
  "testing"

  "github.com/P100sch/Intel8086Simulator/Simulation"
)

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
    logger.SetOutput(&builder)
    err = Simulation.Simulate(data, &logger)
    output := builder.String()
    if err != nil {
      print(output)
      t.Error(fileInfo.Name(), " ", err)
      continue
    }
    expected := strings.ReplaceAll(string(expectedOutput), "\r", "")

    if expected != output {
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
        fmt.Printf("%"+strconv.FormatFloat(digits, 'f', 0, 64)+"d: %s\n\n", i, outputLines[i])
      }
    }
    Simulation.Rest()
  }
}
