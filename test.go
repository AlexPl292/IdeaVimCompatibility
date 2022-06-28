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
	checkEasyMotion("easymotion")
	checkSneak("vim-sneak")
}

func checkEasyMotion(dirName string) {
	recreateDir(dirName)

	cloneRepo(dirName, "https://github.com/AlexPl292/IdeaVim-EasyMotion.git")
	cloneRepo(dirName+"/IdeaVIM", "https://github.com/JetBrains/ideavim.git")
	cloneRepo(dirName+"/AceJump", "https://github.com/acejump/AceJump.git")

	currDir := getWorkingDir()

	updateFile(dirName+"/gradle.properties", func(s []byte) []byte {
		output := bytes.Replace(s, []byte("ideaVimFromMarketplace=true"), []byte("ideaVimFromMarketplace=false"), -1)
		output = bytes.Replace(output, []byte("aceJumpFromMarketplace=true"), []byte("aceJumpFromMarketplace=false"), -1)
		return output
	})

	updateFile(dirName+"/AceJump/build.gradle.kts", func(s []byte) []byte {
		output := bytes.Replace(s, []byte("kotlin(\"jvm\") version \"1.7.0\""), []byte("kotlin(\"jvm\")"), -1)
		output = bytes.Replace(output, []byte("id(\"org.jetbrains.intellij\") version \"1.6.0\""), []byte("id(\"org.jetbrains.intellij\")"), -1)

		output = bytes.Replace(output, []byte("kotlin.jvmToolchain {\n  run {\n    languageVersion.set(JavaLanguageVersion.of(11))\n  }\n}"), []byte("//kotlin.jvmToolchain {\n//  run {\n//    languageVersion.set(JavaLanguageVersion.of(11))\n//  }\n//}"), -1)
		output = bytes.Replace(output, []byte("version.set(\"2022.1.1\")"), []byte("version.set(\"LATEST-EAP-SNAPSHOT\")"), -1)

		return output
	})

	updateFile(dirName+"/IdeaVIM/build.gradle.kts", func(s []byte) []byte {
		output := bytes.Replace(s, []byte("kotlin(\"jvm\") version \"1.6.21\""), []byte("kotlin(\"jvm\")"), -1)
		output = bytes.Replace(output, []byte("id(\"org.jetbrains.intellij\") version \"1.6.0\""), []byte("id(\"org.jetbrains.intellij\")"), -1)

		output = bytes.Replace(output, []byte("implementation(project(\":vim-engine\"))"), []byte("api(project(\":IdeaVIM:vim-engine\"))"), -1)

		return output
	})

	updateFile(dirName+"/settings.gradle.kts", func(s []byte) []byte {
		output := bytes.Replace(s, []byte("include(\"IdeaVIM\", \"AceJump\")"), []byte("include(\"IdeaVIM\", \"AceJump\", \"IdeaVIM:vim-engine\")"), -1)

		return output
	})

	updateFile(dirName+"/build.gradle", func(s []byte) []byte {
		output := bytes.Replace(s, []byte("version = \"2021.3.1\""), []byte("version = \"LATEST-EAP-SNAPSHOT\""), -1)
		output = bytes.Replace(output, []byte("id \"org.jetbrains.kotlin.jvm\" version \"1.5.0\""), []byte("id \"org.jetbrains.kotlin.jvm\" version \"1.6.0\""), -1)

		return output
	})

	// It's needed to use java 11 or something (well, definitely NOT 18)
	runCmd("./gradlew build -x test -x buildSearchableOptions" /*+" -Dorg.gradle.java.home=/Users/Alex.Plate/Library/Java/JavaVirtualMachines/corretto-11.0.11/Contents/Home\n"*/, filepath.Join(currDir, dirName))
}

func checkSneak(dirName string) {
	recreateDir(dirName)

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

	done := appendToFile(dirName+"/settings.gradle.kts", "include(\"IdeaVIM\", \"IdeaVIM:vim-engine\")")
	if done {
		return
	}

	runCmd("./gradlew build -x test -x buildSearchableOptions", filepath.Join(currDir, dirName))
}

func recreateDir(dirName string) {
	log.Printf("Removing old directory")
	if err := os.RemoveAll(dirName); err != nil {
		log.Fatal(err)
	}
	log.Printf("Creating old directory")
	if err := os.Mkdir(dirName, os.ModePerm); err != nil {
		log.Fatal(err)
	}
}

func appendToFile(fileName string, str string) bool {
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
