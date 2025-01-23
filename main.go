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

	for i, dir := range directories {
		splitName := strings.Split(dir, "/")
		cleanedName := splitName[len(splitName)-1]

		display.DisplayAt(cleanedName, 2, 2+i)
	}

	display.DisplayAt("What project are you working on today? ", 2, display.Height-1)
	for {
		input, bytes := input.Read()
		display.DisplayAt(input+strings.Repeat(" ", 80-len(input)), 2, display.Height)
		display.DisplayAt(fmt.Sprintf("%v", bytes), 80, display.Height)
	}

	return "string"
}

type Display struct {
	tty    *os.File
	Width  int
	Height int
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

func (d *Display) DisplayAt(data string, x, y int) {
	command := fmt.Sprintf("\x1b[%d;%dH%v", y, x, data)
	d.tty.WriteString(command)
}

type Input struct {
	oldFdState *term.State
	readBuffer []byte
	value      []byte
}

func NewInput() *Input {
	oldState, _ := term.MakeRaw(int(os.Stdin.Fd()))

	return &Input{ oldFdState: oldState, readBuffer: make([]byte, 3), value: make([]byte, 0, 80)}
}

func (i *Input) Read() (string, []byte) {
	os.Stdin.Read(i.readBuffer)

	switch i.readBuffer[0] {
	case '\x7F':
		if len(i.value) == 0 {
			break
		}

		i.value = i.value[0 : len(i.value)-1]
		break
	default:
		i.value = append(i.value, i.readBuffer...)
	}

	return i.GetValue(), i.readBuffer
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
