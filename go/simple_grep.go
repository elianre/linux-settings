// Linux grep golang version!
// A simple product while I was learning how to program with golang. Just for fun!
// Author: renliang87@gmail.com
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
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

	for i := src.start; ; {
		rb.AddLine(src.content[i], src.index[i])
		i++
		if i == src.end || i == src.count {
			break
		}

		if i == src.size {
			i = 0
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
		ret, _ := regexp.MatchString(matchRegx, s)
		// handle the color of matched part
		if ret == true && options.colorize == true {
			re := regexp.MustCompile(matchRegx)
			s = re.ReplaceAllStringFunc(s, ColorizeMatched)
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
	options := GrepOptions{false, false, true, false, 8, 8}
	ret, _ := Grep("Initializing", "/var/log/dmesg", options)
	if ret != true {
		os.Exit(1)
	}
}
