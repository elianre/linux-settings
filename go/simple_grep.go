// Linux grep golang version!
// A simple product while I was learning how to program with golang. Just for fun!
// Author: renliang87@gmail.com

//How to build: go tool cgo simple_grep.go
//              go build simple_grep.go
package main

//#include <unistd.h>
import "C"

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

type GrepOptions struct {
	ignoreCase    bool
	fixedStrings  bool
	lineNumber    bool
	colorize      bool
	beforeContext int
	afterContext  int
}

type RingBuffer struct {
	size    int
	content []string
	index   []int //for line index
	start   int
	now     int
	end     int
	count   int
}

func (rb *RingBuffer) Init(size int) {
	rb.size = size
	rb.content = make([]string, size)
	rb.index = make([]int, size)
	rb.start = 0
	rb.end = 0
	rb.now = -1
	rb.count = 0
}

func (rb *RingBuffer) AddLine(line string, index int) {
	//fmt.Println(s)
	if rb.now != -1 && rb.index[rb.now] >= index {
		return
	}

	if rb.size > 0 {
		rb.content[rb.end] = line
		rb.index[rb.end] = index
		rb.now = rb.end
		rb.end++
		rb.count++
		if rb.end == rb.size {
			rb.end = 0
		}
		if rb.count > rb.size {
			rb.start++
		}
		if rb.start == rb.size {
			rb.start = 0
		}
	} else {
		rb.content = append(rb.content, line)
		rb.index = append(rb.index, index)
		rb.now = rb.end
		rb.end++
		rb.count++
	}
}

func (rb *RingBuffer) Extend(src *RingBuffer) {
	if src.now == -1 {
		return
	}

	j := 0
	for i := src.start; ; {
		rb.AddLine(src.content[i], src.index[i])
		i++
		j++
		if i == src.end || i == src.count {
			break
		}

		if i == src.size {
			i = 0
		}

		if src.size != 0 && j >= src.size {
			break
		}
	}
}

func (rb RingBuffer) String() string {
	buffer := new(bytes.Buffer)
	j := 0
	for i := rb.start; ; j++ {
		if j == rb.count {
			break
		}
		//buffer.WriteString(fmt.Sprintf("\t%d\t", rb.index[i]))
		buffer.WriteString(rb.content[i])
		buffer.WriteString("\n")
		i++
		if i == rb.end {
			break
		}

		if i == rb.size {
			i = 0
		}
	}
	return buffer.String()
}

func Readln(r *bufio.Reader) (string, error) {
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}
	return string(ln), err
}

func Grep(matchRegx string, file string, options GrepOptions) (status bool, matchedLines []string) {
	status = false
	fh, err := os.Open(file)
	if err != nil {
		fmt.Printf("ERROR: Failed to open file %s: %v\n", file, err)
		return
	}
	defer fh.Close()

	if options.ignoreCase == true && options.fixedStrings == false {
		matchRegx = fmt.Sprintf("(?i)%s", matchRegx)
	}

	matched := new(RingBuffer)
	matched.Init(0)

	buffBefore := new(RingBuffer)
	buffBefore.Init(options.beforeContext)

	buffAfter := new(RingBuffer)
	buffAfter = nil

	r := bufio.NewReader(fh)
	i := 0
	lastMatchedIndex := 0
	s, e := Readln(r)
	for e == nil {
		i++

		ret := false
		if options.fixedStrings == true {
			ret = strings.Contains(s, matchRegx)
			if ret == true && options.colorize == true {
				s = strings.Replace(s, matchRegx, fmt.Sprintf("%c[0;31m%s%c[0m", 27, matchRegx, 27), -1)
			}
		} else {
			re := regexp.MustCompile(matchRegx)
			ret = re.MatchString(s)
			// handle the color of matched part
			if ret == true && options.colorize == true {
				s = re.ReplaceAllStringFunc(s, ColorizeMatched)
			}
		}

		// handle the line number
		if options.lineNumber == true {
			if ret == true {
				if options.colorize == true {
					s = fmt.Sprintf("%c[0;32m%d%c[0m-%s", 27, i, 27, s)
				} else {
					s = fmt.Sprintf("%d-%s", i, s)
				}
			} else {
				if options.colorize == true {
					s = fmt.Sprintf("%c[0;32m%d%c[0m:%s", 27, i, 27, s)
				} else {
					s = fmt.Sprintf("%d:%s", i, s)
				}
			}
		}

		// handle the context before match
		if options.beforeContext > 0 {
			if ret == true {
				matched.Extend(buffBefore)
			} else {
				buffBefore.AddLine(s, i)
			}
		}

		if ret == true {
			matched.AddLine(s, i)
		}

		// handle the context after match
		if options.afterContext > 0 {
			if ret == true {
				lastMatchedIndex = i

				if buffAfter != nil {
					matched.Extend(buffAfter)
					buffAfter = nil
				}
				buffAfter = new(RingBuffer)
				buffAfter.Init(options.afterContext)
			}

			if buffAfter != nil {
				if lastMatchedIndex < i && i-lastMatchedIndex <= options.afterContext {
					buffAfter.AddLine(s, i)
				} else if i-lastMatchedIndex-1 == options.afterContext {
					matched.Extend(buffAfter)
					buffAfter = nil
				}
			}
		}

		s, e = Readln(r)
	}

	fmt.Println(matched)

	status = true
	return
}

func ColorizeMatched(str string) (result string) {
	return fmt.Sprintf("%c[0;31m%s%c[0m", 27, str, 27)
}

func main() {
	ignoreCase := flag.Bool("i", false, "ignore case distinctions")
	fixedStrings := flag.Bool("F", false, "PATTERN is a set of newline-separated fixed strings")
	lineNumber := flag.Bool("n", false, "print line number with output lines")
	color := flag.String("color", "auto", "use markers to highlight the matching strings. value could be 'always', 'never', or 'auto'")
	beforeContext := flag.Int("B", 0, "print NUM lines of leading context")
	afterContext := flag.Int("A", 0, "print NUM lines of trailing context")
	flag.Parse()
	//fmt.Println(*ignoreCase, " ", *fixedStrings, " ", *lineNumber, " ", *color, " ", *beforeContext, " ", *afterContext)
	help := func(f *flag.Flag) {
		fn := *f
		fmt.Printf("  -%s=%s: %s\n", fn.Name, fn.DefValue, fn.Usage)
	}
	if len(flag.Args()) < 2 {
		fmt.Println("ERROR: Wrong options!")
		fmt.Printf("Usage: %s [OPTION] [...] REGX FILE\n", filepath.Base(os.Args[0]))
		flag.VisitAll(help)
		os.Exit(1)
	}
	regx := flag.Args()[0]
	file := flag.Args()[1]

	colorize := true
	if strings.EqualFold(*color, "never") {
		colorize = false
	} else if strings.EqualFold(*color, "auto") {
		if int(C.isatty(C.int(syscall.Stdout))) == 0 {
			colorize = false
		}
	}

	options := GrepOptions{*ignoreCase, *fixedStrings, *lineNumber, colorize, *beforeContext, *afterContext}
	//ret, _ := Grep("3.13.0", "/var/log/dmesg", options)
	ret, _ := Grep(regx, file, options)
	if ret != true {
		os.Exit(1)
	}
}
