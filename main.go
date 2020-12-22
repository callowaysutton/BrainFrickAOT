package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/novalagung/golpal"
)

const (
	usageMessage = "Usage:\n\t<bf source> --run to interpret\n\t<bf source> --export <outfile> to transpile to Go\n\t<bf source> --compile <outfile> to compile Go to a binary\n\t<bf source> --runFS to compile to FS then run"
	dataSize     = int(^uint16(0) >> 1)
	tab          = "\t"
	opIncr       = iota
	opDecr
	opIncrVal
	opDecrVal
	opOut
	opIn
	beginLoop
	endLoop
	zero
)

type opCode struct {
	Type  int
	Value int
}

func main() {
	testArgs(2)
	inputFile := os.Args[1]
	flag := os.Args[2]
	var doOptimizations = true

	buf, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatal(usageMessage)
	}

	if len(buf) == 0 {
		log.Fatal("Input file cannot be empty")
		os.Exit(1)
	}
	program := string(buf)
	opCodes := newQuFunc()
	last := &opCode{
		Type: -1,
	}
	hasopIn := false
	hasOutput := false

	fmt.Println("Converting...")
	for i := 0; i < len(program); i++ {
		switch program[i] {
		case '>':
			if last.Type == opIncr {
				last.Value++
			} else {
				if last.Type != -1 {
					opCodes.Push(last)
				}
				last = &opCode{
					Type:  opIncr,
					Value: 1,
				}
			}
		case '<':
			if last.Type == opDecr {
				last.Value++
			} else {
				if last.Type != -1 {
					opCodes.Push(last)
				}
				last = &opCode{
					Type:  opDecr,
					Value: 1,
				}
			}
		case '+':
			if last.Type == opIncrVal {
				last.Value++
			} else {
				if last.Type != -1 {
					opCodes.Push(last)
				}
				last = &opCode{
					Type:  opIncrVal,
					Value: 1,
				}
			}
		case '-':
			if last.Type == opDecrVal {
				last.Value++
			} else {
				if last.Type != -1 {
					opCodes.Push(last)
				}
				last = &opCode{
					Type:  opDecrVal,
					Value: 1,
				}
			}
		case '.':
			if last.Type != -1 {
				opCodes.Push(last)
			}
			last = &opCode{
				Type: opOut,
			}
			hasOutput = true
		case ',':
			if last.Type != -1 {
				opCodes.Push(last)
			}
			last = &opCode{
				Type: opIn,
			}
			hasopIn = true
		case '[':
			if last.Type != -1 {
				opCodes.Push(last)
			}
			last = &opCode{
				Type: beginLoop,
			}
		case ']':
			if last.Type != -1 {
				opCodes.Push(last)
			}
			last = &opCode{
				Type: endLoop,
			}
		default:
			continue
		}
	}
	if last.Type != -1 {
		opCodes.Push(last)
	}
	fmt.Println("Optimizing...")
	var processed *quFunc
	if doOptimizations {
		processed = optimizer(opCodes)
	} else {
		processed = opCodes
	}

	fmt.Println("Initializing...")
	sb := &strings.Builder{}
	tabs := 0
	_, _ = hasOutput, hasopIn
	addLine(sb, "package main", tabs)
	addLine(sb, "import \"fmt\"", tabs)
	//addLine(sb, "import \"os\"", tabs)
	addLine(sb, "func main() {", tabs)
	tabs++
	addLine(sb, fmt.Sprintf("buffer := make([]byte, %d)", dataSize), tabs)
	addLine(sb, "ptr := 0", tabs)
	if hasopIn {
		addLine(sb, "b := make([]byte, 1)", tabs)
	}
	for opt := processed.Pop(); opt != nil; opt = processed.Pop() {
		switch opt.Type {
		case opIncr:
			addLine(sb, fmt.Sprintf("ptr += %d", opt.Value), tabs)

		case opDecr:
			addLine(sb, fmt.Sprintf("ptr -= %d", opt.Value), tabs)

		case opIncrVal:
			addLine(sb, fmt.Sprintf("buffer[ptr] += byte(%d)", opt.Value), tabs)

		case opDecrVal:
			addLine(sb, fmt.Sprintf("buffer[ptr] -= byte(%d)", opt.Value), tabs)

		case opOut:
			addLine(sb, "fmt.Print(string(buffer[ptr]))", tabs)

		case opIn:
			addLine(sb, "os.Stdin.Read(b)", tabs)
			addLine(sb, "buffer[ptr] = b[0]", tabs)

		case zero:
			addLine(sb, "buffer[ptr] = 0", tabs)

		case beginLoop:
			addLine(sb, "for buffer[ptr] != 0 {", tabs)
			tabs++

		case endLoop:
			tabs--
			addLine(sb, "}", tabs)
		}
	}
	addLine(sb, "_ = buffer", tabs)
	addLine(sb, "_ = ptr", tabs)
	tabs--
	addLine(sb, "}", tabs)

	fmt.Println("Ready.")

	if flag == "--export" {
		testArgs(4)
		exportName := string(os.Args[3])
		result := exportName + ".go"
		os.Remove(result)
		f, err := os.Create(result)
		_, _ = f, err
		err = ioutil.WriteFile(result, []byte(sb.String()), 7777)
		if err != nil {
			panic(err)
		}
		fmt.Println("Exported!")
	}
	if flag == "--compile" {
		testArgs(4)
		exportName := string(os.Args[3]) + ".go"
		result := string(os.Args[3]) + ".exe"
		os.Remove(exportName)
		f, err := os.Create(exportName)
		_, _ = f, err
		err = ioutil.WriteFile(exportName, []byte(sb.String()), 7777)
		if err != nil {
			panic(err)
		}
		fmt.Println("Compiling...")
		cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", result, exportName)
		err = cmd.Run()
		cmd = exec.Command("upx", "--brute", result)
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		fmt.Println("Done.")
	}
	if flag == "--runFS" {
		testArgs(3)
		fmt.Println("Initializing...")
		os.Remove("tmp.t")
		f, err := os.Create("tmp.go")
		_, _ = f, err
		err = ioutil.WriteFile("tmp.go", []byte(sb.String()), 7777)
		fmt.Println("Interpreting...")
		fmt.Print("\033[H\033[2J")
		cmd := exec.Command("go", "run", "tmp.go")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
		os.Remove("tmp.go")
	}
	if flag == "--run" {
		testArgs(3)
		output, err := golpal.New().ExecuteRaw(sb.String())
		if err != nil {
			log.Fatal(usageMessage)
		}
		fmt.Println("Interpreting...")
		fmt.Print("\033[H\033[2J")
		fmt.Print(output)
	}
}

func optimizer(opCodes *quFunc) *quFunc {
	output := newQuFunc()
	for opt := opCodes.Pop(); opt != nil; opt = opCodes.Pop() {
		// Looking for what is essentially [-]
		if opt.Type == beginLoop {
			opt1 := opCodes.Pop()
			opt2 := opCodes.Pop()
			if opt1.Type == opDecrVal && opt2.Type == endLoop {
				output.Push(&opCode{Type: zero})
				continue
			} else {
				// Not a reset, put the values back onto the queue like it never happened
				opCodes.PushHead(opt2)
				opCodes.PushHead(opt1)
			}
		}
		output.Push(opt)
	}
	return output
}

func addLine(sb *strings.Builder, value string, tabs int) {
	for i := 0; i < tabs; i++ {
		sb.WriteString(tab)
	}
	sb.WriteString(value)
	sb.WriteString("\n")
}

type quFunc struct {
	queue []*opCode
}

func newQuFunc() *quFunc {
	return &quFunc{
		queue: make([]*opCode, 0),
	}
}

func (f *quFunc) Push(v *opCode) {
	f.queue = append(f.queue, v)
}

func (f *quFunc) PushHead(v *opCode) {
	f.queue = append([]*opCode{v}, f.queue...)
}

func (f *quFunc) Pop() *opCode {
	if len(f.queue) < 1 {
		return nil
	}
	v := f.queue[0]
	if len(f.queue) == 1 {
		f.queue = make([]*opCode, 0)
	} else {
		f.queue = f.queue[1:]
	}
	return v
}

func (f *quFunc) HasNext() bool {
	return len(f.queue) >= 0
}

func (f *quFunc) Peak() *opCode {
	return f.queue[0]
}

func (f *quFunc) PeakOffset(off int) *opCode {
	if off > len(f.queue) {
		return nil
	}
	return f.queue[off]
}

func (f *quFunc) PeakLast() *opCode {
	return f.queue[len(f.queue)-1]
}

func (f *quFunc) PopLast() *opCode {
	v := f.queue[len(f.queue)-1]
	f.queue = f.queue[:len(f.queue)-1]
	return v
}

func (f *quFunc) Length() int {
	return len(f.queue)
}

func testArgs(n int) {
	if len(os.Args) < n && len(os.Args) < 5 {
		fmt.Print(usageMessage)
		os.Exit(1)
	}
}
