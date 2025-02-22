package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/P100sch/Intel8086Simulator/Simulation/Disassembly"
)

var registers struct {
	AX, BX, CX, DX, SP, BP, SI, DI, IP uint64
	CS, DS, SS, ES                     uint64
}

type modifiedRegister struct {
	name     string
	newValue uint64
}

func main() {
	dir, err := os.MkdirTemp("", "Intel8086SimulatorConversion")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = os.RemoveAll(dir); err != nil {
			print(err)
		}
	}()

	fmt.Println("AX:0x0000 BX:0x0000 CX:0x0000 DX:0x0000 SP:0x0000 BP:0x0000 SI:0x0000 DI:0x0000 IP:0x0000 CS:0x0000 DS:0x0000 SS:0x0000 ES:0x0000 F:          ; JMP 0:0")

	reader := bufio.NewReader(os.Stdin)
	var flags = [9]byte{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}
	for line, err := reader.ReadString('\n'); err == nil; line, err = reader.ReadString('\n') {
		line = strings.ReplaceAll(line, "\n", "")
		line = strings.ReplaceAll(line, "\r", "")
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if line[0] == '-' {
			continue
		}
		segments := strings.Split(line, " ; ")
		asm := strings.ToUpper(segments[0])
		portions := strings.Split(segments[1], "flags:")
		var modifiedRegisters []modifiedRegister
		skipIP := false
		if len(portions[0]) > 10 {
			for _, register := range strings.Split(portions[0], " ") {
				if register != "" {
					newValue, err := strconv.ParseUint(strings.Split(register[3:], "->")[1][2:], 16, 16)
					if err != nil {
						print(err.Error())
						os.Exit(1)
					}
					modifiedRegisters = append(modifiedRegisters, modifiedRegister{
						name:     register[:2],
						newValue: newValue,
					})
				}
			}
			for _, register := range modifiedRegisters {
				switch register.name {
				case "ax":
					registers.AX = register.newValue
				case "bx":
					registers.BX = register.newValue
				case "cx":
					registers.CX = register.newValue
				case "dx":
					registers.DX = register.newValue
				case "sp":
					registers.SP = register.newValue
				case "bp":
					registers.BP = register.newValue
				case "si":
					registers.SI = register.newValue
				case "di":
					registers.DI = register.newValue
				case "ip":
					registers.IP = register.newValue
					skipIP = true
				case "cs":
					registers.CS = register.newValue
				case "ds":
					registers.DS = register.newValue
				case "ss":
					registers.SS = register.newValue
				case "es":
					registers.ES = register.newValue
				}
			}
		}
		if len(portions) > 1 && len(portions[1]) > 1 {
			flags = [9]byte{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}
			for _, flag := range strings.Split(portions[1], "->")[1] {
				switch flag {
				case 'T':
					flags[0] = 'T'
				case 'D':
					flags[1] = 'D'
				case 'I':
					flags[2] = 'I'
				case 'O':
					flags[3] = 'O'
				case 'S':
					flags[4] = 'S'
				case 'Z':
					flags[5] = 'Z'
				case 'A':
					flags[6] = 'A'
				case 'P':
					flags[7] = 'P'
				case 'C':
					flags[8] = 'C'
				}
			}
		}
		offset, disassemblyAsm := getIPOffsetAndASM(asm, dir)
		if !skipIP {
			registers.IP = (registers.IP + offset) & math.MaxUint16
		}
		fmt.Printf("AX:0x%04x BX:0x%04x CX:0x%04x DX:0x%04x SP:0x%04x BP:0x%04x SI:0x%04x DI:0x%04x IP:0x%04x CS:0x%04x DS:0x%04x SS:0x%04x ES:0x%04x F:%9s ; %s\n",
			registers.AX, registers.BX, registers.CX, registers.DX, registers.SP, registers.BP, registers.SI, registers.DI, registers.IP, registers.CS, registers.DS, registers.SS, registers.ES,
			string(flags[:]), disassemblyAsm)
	}

	fmt.Printf("AX:0x%04x BX:0x%04x CX:0x%04x DX:0x%04x SP:0x%04x BP:0x%04x SI:0x%04x DI:0x%04x IP:0x%04x CS:0x%04x DS:0x%04x SS:0x%04x ES:0x%04x F:%9s ; %s\n",
		registers.AX, registers.BX, registers.CX, registers.DX, registers.SP, registers.BP, registers.SI, registers.DI, registers.IP+1, registers.CS, registers.DS, registers.SS, registers.ES,
		string(flags[:]), "HLT")

	os.Exit(0)
}

func getIPOffsetAndASM(asm, tempDir string) (uint64, string) {
	asmFileName := filepath.Join(tempDir, "instruction.asm")
	err := os.WriteFile(asmFileName, []byte(asm), 0666)
	if err != nil {
		log.Fatal(err)
	}

	outputFileName := filepath.Join(tempDir, "instruction.bin")
	var outputData []byte
	command := exec.Command("nasm", "-f", "bin", asmFileName, "-o", outputFileName)
	builder := strings.Builder{}
	command.Stdout = &builder
	command.Stderr = &builder
	err = command.Run()
	if err != nil {
		log.Fatal("assembly error\n", builder.String())
	}
	outputData, err = os.ReadFile(outputFileName)
	if err != nil {
		log.Fatal(err)
	}
	asm, err = Disassembly.Disassemble(outputData[:])
	if err != nil {
		log.Fatal(err)
	}
	return uint64(len(outputData)), asm
}
