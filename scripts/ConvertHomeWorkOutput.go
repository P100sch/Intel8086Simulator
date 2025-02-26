package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
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

type state struct {
	clocks            string
	flags             [9]byte
	modifiedRegisters []modifiedRegister
}

var clockShift = 15

func main() {
	var err error
	var inFile = os.Stdin
	var outWriter = os.Stdout
	var skipClock = false

	dir, err := os.MkdirTemp("", "Intel8086SimulatorConversion")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = os.RemoveAll(dir); err != nil {
			print(err)
		}
	}()

	if len(os.Args) > 1 {
		if index := strings.LastIndex(os.Args[1], "listing_"); index != -1 {
			parts := strings.Split(os.Args[1][index:], "_")
			if number, err := strconv.Atoi(parts[1]); err == nil {
				skipClock = number < 56
			}
		}

		if os.Args[1][0] == '.' || os.Args[1][0] == os.PathSeparator {
			inFile, err = os.Open(os.Args[1])
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				_ = inFile.Close()
			}()
		} else {
			inFile, err = os.Create(path.Join(dir, "input.txt"))
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				_ = inFile.Close()
			}()
			response, err := http.DefaultClient.Get(os.Args[1])
			if err != nil {
				log.Fatal(err)
			}
			defer func() {
				_ = response.Body.Close()
			}()
			if response.StatusCode != 200 {
				log.Fatal(response.Status, " (", response.StatusCode, ")")
			}
			_, err = inFile.ReadFrom(response.Body)
			if err != nil {
				log.Fatal(err)
			}
			_, err = inFile.Seek(0, 0)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	if len(os.Args) > 2 {
		outWriter, err = os.Create(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
	}
	asmFileName := filepath.Join(dir, "instructions.asm")
	asmFile, err := os.Create(asmFileName)
	if err != nil {
		log.Fatal(err)
	}

	_, err = fmt.Fprint(outWriter, "AX:0x0000 BX:0x0000 CX:0x0000 DX:0x0000 SP:0x0000 BP:0x0000 SI:0x0000 DI:0x0000 IP:0x0000 CS:0x0000 DS:0x0000 SS:0x0000 ES:0x0000 F:          ; JMP 0:0 ; 5bytes")
	if err != nil {
		log.Fatal(err)
	}
	if skipClock {
		_, err = fmt.Fprintln(outWriter)
	} else {
		_, err = fmt.Fprintln(outWriter, " +15 = 15")
	}
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(inFile)
	var flags = [9]byte{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}
	var states = make([]state, 0, 500)
	for line, err := reader.ReadString('-'); err == nil; line, err = reader.ReadString('\n') {
		line = strings.ReplaceAll(line, "\n", "")
		line = strings.ReplaceAll(line, "\r", "")
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if line[0] == '-' || line[0] == '*' {
			continue
		}

		segments := strings.Split(line, " ; ")
		_, err = asmFile.WriteString(segments[0] + "\n")
		if err != nil {
			log.Fatal(err)
		}

		var modifiedRegisters []modifiedRegister
		var clocks, rest string
		if skipClock {
			rest = segments[1]
		} else {
			clocks, rest = separateClocks(segments[1], segments[0])
		}
		modifiedRegisters, flags = separateRegistersAndFlags(rest, flags)

		states = append(states, state{clocks, flags, modifiedRegisters})
	}

	outputFileName := filepath.Join(dir, "instructions.bin")
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
	disassemblyAsm, err := Disassembly.Disassemble(outputData[:])
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(disassemblyAsm, "\n")

	var lastClock string
	for i, currentState := range states {
		skipIP := false
		for _, register := range currentState.modifiedRegisters {
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
		line := lines[i]
		if !skipIP {
			offset := getInstructionLength(line)
			registers.IP = (registers.IP + offset) & math.MaxUint16
		}
		_, err = fmt.Fprintf(outWriter, "AX:0x%04x BX:0x%04x CX:0x%04x DX:0x%04x SP:0x%04x BP:0x%04x SI:0x%04x DI:0x%04x IP:0x%04x CS:0x%04x DS:0x%04x SS:0x%04x ES:0x%04x F:%9s ; %s%s\n",
			registers.AX, registers.BX, registers.CX, registers.DX, registers.SP, registers.BP, registers.SI, registers.DI, registers.IP, registers.CS, registers.DS, registers.SS, registers.ES,
			string(currentState.flags[:]), line, currentState.clocks)
		if err != nil {
			log.Fatal(err)
		}
		lastClock = currentState.clocks
	}

	_, err = fmt.Fprintf(outWriter, "AX:0x%04x BX:0x%04x CX:0x%04x DX:0x%04x SP:0x%04x BP:0x%04x SI:0x%04x DI:0x%04x IP:0x%04x CS:0x%04x DS:0x%04x SS:0x%04x ES:0x%04x F:%9s ; %s%s\n",
		registers.AX, registers.BX, registers.CX, registers.DX, registers.SP, registers.BP, registers.SI, registers.DI, registers.IP, registers.CS, registers.DS, registers.SS, registers.ES,
		string(flags[:]), "HLT ; 1bytes", modifyHLTClocks(lastClock))
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}

func separateClocks(s, asm string) (clocks, rest string) {
	portions := strings.Split(s, " | ")
	if len(portions) != 2 {
		clocks = ""
		rest = s
		return
	}
	clocks = strings.TrimPrefix(portions[0], "Clocks:")
	override := strings.Split(clocks, " = ")[0]
	if strings.Contains(asm, "[bp]") {
		clock, err := strconv.Atoi(override[2:])
		if err != nil {
			log.Fatal(err)
		}
		clock += 4
		override = " +" + strconv.Itoa(clock)
		clocks = strings.Replace(clocks, "5ea", "9ea", 1)
		clockShift += 4
	}
	clocks = incrementClockTotal(clocks, override, clockShift)
	rest = portions[1]
	return
}

func separateRegistersAndFlags(s string, defaultFlags [9]byte) (registers []modifiedRegister, flags [9]byte) {
	portions := strings.Split(s, "flags:")
	registers = make([]modifiedRegister, 0, 1)
	flags = defaultFlags

	if len(portions[0]) > 10 {
		for _, register := range strings.Split(portions[0], " ") {
			if register != "" {
				newValue, err := strconv.ParseUint(strings.Split(register, "->")[1][2:], 16, 16)
				if err != nil {
					print(err.Error())
					os.Exit(1)
				}
				registers = append(registers, modifiedRegister{
					name:     register[:2],
					newValue: newValue,
				})
			}
		}
	}

	if len(portions) < 2 {
		return
	}

	if len(portions[1]) > 1 {
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
	return
}

func getInstructionLength(s string) uint64 {
	portions := strings.Split(s, " ; ")

	offset, err := strconv.ParseUint(strings.TrimSuffix(portions[1], "bytes"), 10, 64)
	if err != nil {
		log.Fatal(err)
	}

	return offset
}

func incrementClockTotal(s, clockOverride string, increment int) string {
	if len(s) < 6 {
		return ""
	}
	segments := strings.Split(s, " = ")
	portions := strings.Split(segments[1], " (")
	count, err := strconv.Atoi(portions[0])
	if err != nil {
		log.Fatal(err)
	}
	clock := clockOverride
	clock += " = " + strconv.Itoa(count+increment)
	if len(portions) > 1 {
		clock += " (" + portions[1]
	}
	return clock
}

func modifyHLTClocks(clocks string) string {
	if len(clocks) < 6 {
		return ""
	}
	segments := strings.Split(clocks, " = ")
	portions := strings.Split(segments[1], " (")
	count, err := strconv.Atoi(portions[0])
	if err != nil {
		log.Fatal(err)
	}
	clock := " +2"
	clock += " = " + strconv.Itoa(count+2)
	return clock
}
