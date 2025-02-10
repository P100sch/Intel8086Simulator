package main

import (
  "os"

  "github.com/P100sch/Intel8086Simulator/Simulation/Disassembly"
)

func main() {
  if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
    printHelp()
    os.Exit(0)
  }

  filePath := os.Args[1]
  outputFilePath := filePath + ".asm"

  if len(os.Args) > 2 {
    if len(os.Args) == 3 || len(os.Args) > 4 {
      println("Wrong number of arguments!")
      printHelp()
      os.Exit(1)
    }
    if os.Args[2] != "-o" {
      println("Wrong argument!")
      printHelp()
      os.Exit(1)
    }
    outputFilePath = os.Args[3]
  }

  data, err := os.ReadFile(filePath)
  if err != nil {
    println("Error reading file!")
    println(err.Error())
    os.Exit(2)
  }

  var output string
  output, err = Disassembly.Disassemble(data)
  if err != nil {
    println("Error decoding instructions!")
    println(err.Error())
    os.Exit(3)
  }

  err = os.WriteFile(outputFilePath, []byte(output), 0644)
  if err != nil {
    println("Error writing file!")
    println(err.Error())
    os.Exit(4)
  }
}

func printHelp() {
  println("Intel8086Simulator instructions.bin [-o out.asm]")
  println("Outputs a disassembly of the instruction stream to the console or to the file specified by the `-o` flag.")
}
