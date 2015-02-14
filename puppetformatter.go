package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

func main() {

	if len(os.Args) != 2 {
		fmt.Println("Usage: puppetformatter <file|directory>")
		return
	}

	path, err := filepath.Abs(os.Args[1])
	chkerr(err)

	info, err := os.Stat(path)
	chkerr(err)

	if info.IsDir() {
		processDirectory(path)
	} else {
		processFile(path, info)
	}
}

//recursively walk all directories below root and process any files that are found
func processDirectory(root string) {

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		chkerr(err)
		if !info.IsDir() {
			processFile(path, info)
		}
		return nil
	})
	chkerr(err)
}

//format individual file as needed
func processFile(path string, info os.FileInfo) {

	if !isPuppetFile(info) {
		return
	}

	fmt.Println(path)

	//these files aren't typically that large so just read it in all at once
	file, err := ioutil.ReadFile(path)
	chkerr(err)

	indent := 0

	//handle any formatting that can be done without the context of surrounding lines
	lines := strings.Split(string(file), "\n")
	for i, line := range lines {
		if (strings.Count(line, "}") > strings.Count(line, "{") || (strings.Count(line, "}") == strings.Count(line, "{") && strings.HasPrefix(strings.TrimLeftFunc(line, unicode.IsSpace), "}"))) && indent > 0 {
			indent -= 1
		}

		lines[i] = processLine(line, indent)

		//ensure resources have an empty line above them
		if indent == 1 && i > 0 && strings.Count(line, "{") > strings.Count(line, "}") && len(lines[i-1]) > 1 {
			lines[i-1] += "\n"
		}

		if strings.Count(line, "{") > strings.Count(line, "}") || (strings.Count(line, "}") == strings.Count(line, "{") && strings.HasPrefix(strings.TrimLeftFunc(line, unicode.IsSpace), "}")) {
			indent += 1
		}
	}

	//align all of the rockets (=>) within each resource
	formatRockets(lines)

	//join the lines back up and overwrite the original filen
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(path, []byte(output), 0644)
	chkerr(err)
}

func processLine(line string, indent int) string {

	if len(line) > 0 {
		line = formatWhiteSpace(line, indent)
	}
	if len(line) > 0 {
		line = formatComments(line)
	}
	if !strings.Contains(line, "#") {
		line = formatQuotes(line)
	}

	return line
}

func formatWhiteSpace(line string, indent int) string {

	//trim any whitespace
	line = strings.TrimFunc(line, unicode.IsSpace)

	//replace tabs with two spaces
	line = strings.Replace(line, "\t", "  ", -1)

	//indent line based on current nesting level
	if indent == 0 && strings.HasPrefix(line, "$") {
		indent = 1
	}
	return strings.Repeat(" ", indent*2) + line
}

func formatComments(line string) string {

	//convert // to # when not in quotes
	re := regexp.MustCompile("(['\"]).*//.*(['\"])")
	if strings.Contains(line, "//") && !strings.Contains(line, "#") && !re.MatchString(line) {
		line = strings.Replace(line, "//", "#", 1)
	}

	//convert /*...*/ to # when on single line
	re = regexp.MustCompile("/\\*(.*)\\*/")
	line = re.ReplaceAllString(line, "#${1}")

	return line
}

func formatQuotes(line string) string {

	//remove quotes from standalone variables as long as they aren't passwords
	if !strings.Contains(line, "password") {
		re := regexp.MustCompile("['\"]\\$([[:alnum:]_\\{\\}]*)['\"]")
		line = re.ReplaceAllString(line, "$$${1}")
	}

	//replace double quotes with single quotes when not wrapping variable or single quote
	//TODO this is currently restricted to lines where there are only two double quotes and no single quotes since there are so many special cases to handle
	if !strings.Contains(line, "'") && strings.Count(line, "\"") == 2 {
		re := regexp.MustCompile("\"([^'\"$]*)\"")
		line = re.ReplaceAllString(line, "'${1}'")
	}

	return line
}

func formatRockets(lines []string) {

	start := 0
	count := 0

	for i, line := range lines {
		if strings.Contains(line, "=>") {
			//in a block of rockets
			if start == 0 {
				start = i
				count = 0
			} else {
				count += 1
			}
		} else if start > 0 {
			//finished block of rockets
			alignRockets(lines, start, count+1)
			start = 0
			count = 0
		}
	}
}

func alignRockets(lines []string, start int, count int) {

	max := 0

	block := make([][]string, count)

	//split each line on rocket and trim surrounding whitespace to determine longest left side
	for i := 0; i < count; i++ {
		sides := strings.SplitN(lines[start+i], "=>", 2)
		sides[0] = strings.TrimRightFunc(sides[0], unicode.IsSpace)
		sides[1] = strings.TrimLeftFunc(sides[1], unicode.IsSpace)
		if len(sides[0]) > max {
			max = len(sides[0])
		}
		block[i] = sides
	}

	//align the left side lengths and reassemble sides
	for i := 0; i < count; i++ {
		size := len(block[i][0])
		if size < max {
			block[i][0] += strings.Repeat(" ", max-size)
		}
		lines[start+i] = block[i][0] + " => " + block[i][1]
	}
}

func isPuppetFile(info os.FileInfo) bool {
	//TODO this could be more configurable
	return strings.HasSuffix(strings.ToLower(info.Name()), ".pp")
}

func chkerr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
