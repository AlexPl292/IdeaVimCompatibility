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
)

func main() {
	log.Printf("Removing old directory")
	if err := os.RemoveAll("test"); err != nil {
		log.Fatal(err)
	}
	log.Printf("Creating old directory")
	if err := os.Mkdir("test", os.ModePerm); err != nil {
		log.Fatal(err)
	}

	log.Printf("Clone ideavim-sneak plugin")
	if _, err := git.PlainClone("test", false, &git.CloneOptions{
		URL:      "https://github.com/Mishkun/ideavim-sneak.git",
		Progress: os.Stdout,
	}); err != nil {
		log.Fatal(err)
	}

	log.Printf("Clone ideavim plugin")
	if _, err := git.PlainClone("test/IdeaVIM", false, &git.CloneOptions{
		URL:      "https://github.com/JetBrains/ideavim.git",
		Progress: os.Stdout,
	}); err != nil {
		log.Fatal(err)
	}

	currDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Update gradle wrapper")
	cmd := exec.Command("./gradlew", "wrapper", "--gradle-version", "7.4.2")
	cmd.Dir = filepath.Join(currDir, "test")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(string(out))
		log.Fatal(err.Error())
	}
	fmt.Printf("%s\n", out)

	log.Printf("Update files")
	input, err := ioutil.ReadFile("test/build.gradle.kts")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	output := bytes.Replace(input, []byte("id(\"org.jetbrains.intellij\") version \"1.0\""), []byte("id(\"org.jetbrains.intellij\") version \"1.6.0\""), -1)
	output = bytes.Replace(output, []byte("kotlin(\"jvm\") version \"1.4.10\""), []byte("kotlin(\"jvm\") version \"1.6.21\""), -1)
	output = bytes.Replace(output, []byte("version.set(\"2020.1\")"), []byte("version.set(\"LATEST-EAP-SNAPSHOT\")"), -1)
	output = bytes.Replace(output, []byte("plugins.set(listOf(\"IdeaVIM:0.61\"))"), []byte("plugins.set(listOf(project(\":IdeaVIM\")))"), -1)

	if err = ioutil.WriteFile("test/build.gradle.kts", output, 0666); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	input, err = ioutil.ReadFile("test/IdeaVIM/build.gradle.kts")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	output = bytes.Replace(input, []byte("implementation(project(\":vim-engine\"))"), []byte("implementation(project(\":IdeaVIM:vim-engine\"))"), -1)

	if err = ioutil.WriteFile("test/IdeaVIM/build.gradle.kts", output, 0666); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// ------

	f, err := os.OpenFile("test/settings.gradle.kts", os.O_APPEND|os.O_WRONLY, 0644)

	_, err = f.WriteString("include(\"IdeaVIM\", \"IdeaVIM:vim-engine\")")

	err = f.Close()
	if err != nil {
		return
	}

	// ------

	log.Printf("Run build")
	cmd = exec.Command("./gradlew", "build", "-x", "test", "-x", "buildSearchableOptions")
	cmd.Dir = filepath.Join(currDir, "test")
	out, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Printf(string(out))
		log.Fatal(err.Error())
	}
	fmt.Printf("%s\n", out)
}
