package main

import (
  "log"
  "os"
  "strings"

  "github.com/P100sch/Intel8086Simulator/Simulation"
  "github.com/P100sch/Intel8086Simulator/Simulation/Disassembly"
)

func main() {
  if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
    printHelp()
    os.Exit(0)
  }

  var filePath, outputFilePath string
  var disassemble, verbose bool

  if len(os.Args) > 2 {
    outputFlag := false
    for _, arg := range os.Args[2:] {
      if arg[0] == '-' {
        if strings.ContainsAny(arg, "v") {
          if filePath == "" {
            verbose = true
          } else {
            println("flags need to be passed before the assembly file path")
            printHelp()
            os.Exit(1)
          }
        }
        if strings.ContainsAny(arg, "d") {
          if filePath == "" {
            disassemble = true
          } else {
            println("flags need to be passed before the assembly file path")
            printHelp()
            os.Exit(1)
          }
        }
        if strings.ContainsAny(arg, "o") {
          if filePath != "" {
            outputFlag = true
          } else {
            println("output flags needs to be passed after the assembly file path")
            printHelp()
            os.Exit(1)
          }
        }
      } else {
        if outputFilePath != "" {
          println("too many arguments")
          printHelp()
          os.Exit(1)
        }
        if outputFlag {
          outputFilePath = arg
        } else {
          filePath = arg
        }
      }
    }
    outputFilePath = os.Args[3]
  }

  data, err := os.ReadFile(filePath)
  if err != nil {
    println("Error reading file!")
    println(err.Error())
    os.Exit(2)
  }

  if disassemble {
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
  } else {
    var logger *log.Logger
    if verbose {
      logger = log.Default()
    }
    err = Simulation.Simulate(data, logger)
    if err != nil {
      println("Error running simulation!")
      println(err.Error())
      os.Exit(5)
    }
  }
}

func printHelp() {
  println("Intel8086Simulator [-v|d] instructions.bin [-o out.asm]")
  println("Simulates the execution of the instruction stream.")
  println("-v Outputs disassembly and the state of the registers after each instruction.")
  println("-d Only outputs a disassembly of the instruction stream to the console or to the file specified by the `-o` flag.")
}
