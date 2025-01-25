package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	term "golang.org/x/term"
)

const DIR_CONFIG = "$HOME/.config/wcode"
const FILENAME_SELECTION = "selection"
const ENV_PROJECT_PATHS = "$WCODE_PATHS"

const (
	EXIT_OK          = 0
	EXIT_NO_PROJECTS = 1
	EXIT_BAD_PATH    = 2
	EXIT_TERMINATED  = 9
)

func main() {
	err := setupFiles()
	if err != nil {
		fmt.Println("An unexpected error occured while initializing the config directory:", err.Error())
		os.Exit(EXIT_BAD_PATH)
	}

	projectRoots := gatherProjectPaths()

	directories, err := gatherProjects(projectRoots)
	if err != nil {
		fmt.Printf("There was a problem while collecting the projects: %v\n", err.Error())
		os.Exit(EXIT_BAD_PATH)
	}

	if len(directories) == 0 {
		fmt.Println("There don't exist any projects in the directories: ", projectRoots)
		os.Exit(EXIT_NO_PROJECTS)
	}

	selection := display(directories)
	fmt.Printf("Opening: %v\n", selection)

	saveSelectionToDisk(selection)
	if err != nil {
		fmt.Println("An unexpected error occured while saving the selection:", err.Error())
		os.Exit(EXIT_BAD_PATH)
	}
}

func gatherProjects(roots []string) ([]string, error) {
	directories := []string{}
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, err
		}

		for _, v := range entries {
			if v.Type().IsDir() == false {
				continue
			}

			projectPath := strings.ReplaceAll(root+string(os.PathSeparator)+v.Name(), "//", "/")
			directories = append(directories, projectPath)
		}
	}

	return directories, nil
}

func display(directories []string) string {
	input := NewInput()
	defer input.Close()

	display := NewDisplay()
	display.Clear()

	displayInputGraphic(display)

	for {
		queriedDirectories := directories
		if len(input.GetValue()) != 0 {
			queriedDirectories = getProjectMatches(directories, input.GetValue(), false)
		}

		for i, dir := range queriedDirectories {
			splitName := strings.Split(dir, "/")
			path := strings.Join(splitName[0:len(splitName)-1], "/") + "/"
			project := splitName[len(splitName)-1]

			display.AddModifier("\x1b[2m")
			display.DisplayAt(path, 2, 2+i)
			display.DisplayAt(project, len(path)+2, 2+i)
		}

		display.MoveCursorAt(3+len(input.GetValue()), display.Height-1)
		input, bytes, finished := input.Read()

		display.Clear()

		displayInputGraphic(display)
		display.AddModifier("\x1b[1m")
		display.DisplayAt(input, 3, display.Height-1)
		display.DisplayAt(fmt.Sprintf("%v", bytes), 85, display.Height-1)
		if finished {
			break
		}
	}

	return input.GetValue()
}

type Display struct {
	tty    *os.File
	Width  int
	Height int

	modifier string
}

func NewDisplay() *Display {
	f, err := os.OpenFile("/dev/tty", os.O_RDWR|os.O_APPEND|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Println("Error:", err)
	}

	output := &bytes.Buffer{}

	cmd := exec.Command("tput", "lines")
	cmd.Stdin = os.Stdin
	cmd.Stdout = output
	cmd.Run()
	height, _ := strconv.Atoi(strings.Trim(output.String(), " \n\t"))
	output.Reset()

	if height == 0 {
		height = 24
	}

	cmd = exec.Command("tput", "cols")
	cmd.Stdin = os.Stdin
	cmd.Stdout = output
	cmd.Run()
	width, _ := strconv.Atoi(strings.Trim(output.String(), " \n\t"))
	output.Reset()

	if width == 0 {
		width = 80
	}

	return &Display{
		tty:    f,
		Width:  0,
		Height: height,
	}
}

func (d *Display) Clear() {
	d.tty.WriteString("\x1b[H\x1b[J")
}

func (d *Display) AddModifier(modifier string) {
	d.modifier = modifier
}

func (d *Display) MoveCursorAt(x, y int) {
	d.tty.WriteString(fmt.Sprintf("\x1b[%d;%dH", y, x))
}

func (d *Display) DisplayAt(data string, x, y int) {
	d.MoveCursorAt(x, y)
	command := data
	if len(d.modifier) != 0 {
		command = d.modifier + data + "\x1b[0m"
		d.modifier = ""
	}

	d.tty.WriteString(command)
}

type Input struct {
	oldFdState *term.State
	readBuffer []byte
	value      []byte
}

func NewInput() *Input {
	oldState, _ := term.MakeRaw(int(os.Stdin.Fd()))

	return &Input{oldFdState: oldState, readBuffer: make([]byte, 3), value: make([]byte, 0, 80)}
}

func (in *Input) Read() (string, []byte, bool) {
	finished := false
	in.readBuffer[1] = 0
	in.readBuffer[2] = 0
	os.Stdin.Read(in.readBuffer)

	switch in.readBuffer[0] {
	case '\x7F':
		if len(in.value) == 0 {
			break
		}

		in.value = in.value[0 : len(in.value)-1]
		break
	case '\x03', '\x18':
		in.Close()
		os.Exit(EXIT_TERMINATED)
		break
	case '\x0D':
		finished = true
		break
	case '\x1B':
		if in.readBuffer[1] == '\x7F' {
			foundSpace := false
			foundWord := false
			i := len(in.value) - 1
			for i >= 0 && !foundSpace {
				if foundWord && in.value[i] == '\x20' {
					foundSpace = true
				} else {
					foundWord = true
					i--
				}
			}

			in.value = in.value[0 : i+1]
		}
		break
	default:
		in.value = append(in.value, in.readBuffer[0])
	}

	return in.GetValue(), in.readBuffer, finished
}

func (i *Input) GetValue() string {
	return string(i.value)
}

func (i *Input) Close() {
	term.Restore(int(os.Stdin.Fd()), i.oldFdState)
}

func setupFiles() error {
	baseDir := os.ExpandEnv(DIR_CONFIG)
	return os.MkdirAll(baseDir, 0751)
}

func saveSelectionToDisk(selection string) error {
	baseDir := os.ExpandEnv(DIR_CONFIG)
	file, err := os.Create(baseDir + string(os.PathSeparator) + FILENAME_SELECTION)
	if err == nil {
		file.Write([]byte(selection))
	}

	return err
}

func gatherProjectPaths() []string {
	pathsString := os.ExpandEnv(ENV_PROJECT_PATHS)
	return strings.Split(pathsString, " ")
}

func displayInputGraphic(display *Display) {
	display.AddModifier("\x1b[2;36m")
	display.DisplayAt("┌─ What project are you working on today? ────────────────────────────────────────┐", 1, display.Height-3)
	display.AddModifier("\x1b[2;36m")
	display.DisplayAt("│                                                                                 │", 1, display.Height-2)
	display.AddModifier("\x1b[2;36m")
	display.DisplayAt("│                                                                                 │", 1, display.Height-1)
	display.AddModifier("\x1b[2;36m")
	display.DisplayAt("└─────────────────────────────────────────────────────────────────────────────────┘", 1, display.Height)
}

func getProjectMatches(dirs []string, needle string, matchPath bool) []string {
	res := make([]string, 0, len(dirs))

	for _, hay := range dirs {
		if !isMatch(strings.ToLower(hay), strings.ToLower(needle), matchPath) {
			continue
		}

		res = append(res, hay)
	}

	return res
}

func isMatch(haystack string, needle string, matchPath bool) bool {
	for i := len(haystack) - 1; i >= len(needle)-1; i-- {
		if !matchPath && haystack[i] == '/' {
			return false
		}

		j := 0
		for j < len(needle) && haystack[i-j] == needle[len(needle)-j-1] {
			j++
		}

		if j == len(needle) {
			return true
		}
	}

	return false
}
