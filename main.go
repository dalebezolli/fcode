package main

import (
	"fmt"
	"os"
	"strings"
)

const DIR_CONFIG = "$HOME/.config/wcode"
const FILENAME_SELECTION = "selection"
const ENV_PROJECT_PATHS = "$WCODE_PATHS"

func main() {
	setupFiles()

	projectRoots := gatherProjectPaths()

	directories := gatherProjects(projectRoots)
	if len(directories) == 0 {
		fmt.Println("There don't exist any projects in the following roots: ", projectRoots)
		return
	}

	selection := display(directories)
	fmt.Printf("Opening: %v\n", selection)

	saveSelectionToDisk(selection)
}

func gatherProjects(roots []string) []string {
	directories := []string{}
	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			fmt.Printf("There was a problem while reading %q: %v\n", root, err.Error())
			return []string{}
		}

		for _, v := range entries {
			if v.Type().IsDir() == false {
				continue
			}

			projectPath := strings.ReplaceAll(root+string(os.PathSeparator)+v.Name(), "//", "/")
			directories = append(directories, projectPath)
		}
	}

	return directories
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

func setupFiles() {
	baseDir := os.ExpandEnv(DIR_CONFIG)

	err := os.MkdirAll(baseDir, 0751)
	if err != nil {
		fmt.Println("An unexpected error occured while initializing the config directory:", err.Error())
	}
}

func saveSelectionToDisk(selection string) {
	baseDir := os.ExpandEnv(DIR_CONFIG)

	file, err := os.Create(baseDir + string(os.PathSeparator) + FILENAME_SELECTION)
	if err != nil {
		fmt.Println("An unexpected error occured while saving the selection:", err.Error())
		return
	}

	file.Write([]byte(selection))
}

func gatherProjectPaths() []string {
	pathsString := os.ExpandEnv(ENV_PROJECT_PATHS)
	return strings.Split(pathsString, " ")
}
