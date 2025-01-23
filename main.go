package main

import (
	"fmt"
	"os"
	"strings"
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
	for i, dir := range directories {
		splitName := strings.Split(dir, "/")
		cleanedName := splitName[len(splitName)-1]

		fmt.Printf("%d: %v\n", i, cleanedName)
	}

	var selection int
	fmt.Println("What project are you working on today?")
	for {
		fmt.Scanf("%d", &selection)

		if selection >= len(directories) {
			fmt.Printf("Specify a value from 0 to %d\n", len(directories))
		} else {
			return directories[selection]
		}
	}
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
