package search

import (
	"fmt"
	"strings"
	"sync/atomic"
)

func (ss *SuperSearch) handleMatches(sf *searchFile) {
	var (
		output strings.Builder

		lineNo       = 1
		matchIndex   = 0
		printingLine = false
		lastIndex    = 0
		done         = false
	)

	atomic.AddUint64(&ss.numMatches, uint64(len(sf.matches)))

	fName := strings.Replace(sf.path, ss.workDir, "", -1)
	if fName[0] == '/' {
		fName = fName[1:]
	}

	output.WriteString(highlightFile.Sprintf("%v\n", fName))

	for i := 0; i < len(sf.buf); i++ {
		if sf.buf[i] == '\n' {

			if printingLine {
				output.Write(sf.buf[lastIndex:i])
				if done {
					break
				}
				output.WriteRune('\n')
				printingLine = false
			}

			lineNo++
			lastIndex = i + 1
		}

		if done {
			if printingLine {
				continue
			}
			break
		}

		if i == sf.matches[matchIndex] {
			matchIndex++

			// Print line number, followed by each match
			if !printingLine {
				output.WriteString(highlightNumber.Sprintf("%v:", lineNo))
			}

			output.Write(sf.buf[lastIndex:i])
			output.WriteString(highlightMatch.Sprint(string(sf.buf[i : i+len(ss.opts.Pattern)])))

			if matchIndex == len(sf.matches) {
				done = true
			}

			lastIndex = i + len(ss.opts.Pattern)
			printingLine = true
		}
	}

	output.WriteString("\n\n")
	fmt.Print(output.String())
}
