package main

import (
	"bytes"
	"fmt"
	"github.com/go-git/go-git/v5"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	checkSneak("vim-sneak")
}

func checkSneak(dirName string) {
	log.Printf("Removing old directory")
	if err := os.RemoveAll(dirName); err != nil {
		log.Fatal(err)
	}
	log.Printf("Creating old directory")
	if err := os.Mkdir(dirName, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	cloneRepo(dirName, "https://github.com/Mishkun/ideavim-sneak.git")
	cloneRepo(dirName+"/IdeaVIM", "https://github.com/JetBrains/ideavim.git")

	currDir := getWorkingDir()

	runCmd("./gradlew wrapper --gradle-version 7.4.2", filepath.Join(currDir, dirName))

	updateFile(dirName+"/build.gradle.kts", func(s []byte) []byte {
		output := bytes.Replace(s, []byte("id(\"org.jetbrains.intellij\") version \"1.0\""), []byte("id(\"org.jetbrains.intellij\") version \"1.6.0\""), -1)
		output = bytes.Replace(output, []byte("kotlin(\"jvm\") version \"1.4.10\""), []byte("kotlin(\"jvm\") version \"1.6.21\""), -1)
		output = bytes.Replace(output, []byte("version.set(\"2020.1\")"), []byte("version.set(\"LATEST-EAP-SNAPSHOT\")"), -1)
		output = bytes.Replace(output, []byte("plugins.set(listOf(\"IdeaVIM:0.61\"))"), []byte("plugins.set(listOf(project(\":IdeaVIM\")))"), -1)
		return output
	})

	updateFile(dirName+"/IdeaVIM/build.gradle.kts", func(s []byte) []byte {
		return bytes.Replace(s, []byte("implementation(project(\":vim-engine\"))"), []byte("implementation(project(\":IdeaVIM:vim-engine\"))"), -1)
	})

	done := funcName(dirName+"/settings.gradle.kts", "include(\"IdeaVIM\", \"IdeaVIM:vim-engine\")")
	if done {
		return
	}

	runCmd("./gradlew build -x test -x buildSearchableOptions", filepath.Join(currDir, dirName))
}

func funcName(fileName string, str string) bool {
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)

	_, err = f.WriteString(str)

	err = f.Close()
	if err != nil {
		return true
	}
	return false
}

func updateFile(filePath string, modification func([]byte) []byte) {
	log.Printf("Update " + filePath)
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	output := modification(input)

	if err = ioutil.WriteFile(filePath, output, 0666); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runCmd(command string, dirName string) {
	log.Printf("run command: " + command)
	fields := strings.Fields(command)
	cmd := exec.Command(fields[0], fields[1:]...)
	cmd.Dir = dirName
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(string(out))
		log.Fatal(err.Error())
	}
	fmt.Printf("%s\n", out)
}

func getWorkingDir() string {
	currDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return currDir
}

func cloneRepo(dirName string, url string) {
	log.Printf("Clone " + url)
	if _, err := git.PlainClone(dirName, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	}); err != nil {
		log.Fatal(err)
	}
}
