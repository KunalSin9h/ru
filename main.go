package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Test struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type Problem struct {
	Name  string `json:"name"`
	Tests []Test `json:"tests"`
}

func startServerAndParse() error {
	done := make(chan bool)

	server := &http.Server{
		Addr: ":6174",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			data, err := io.ReadAll(r.Body)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			var problem Problem
			err = json.Unmarshal(data, &problem)

			if err != nil {
				fmt.Println(err.Error())
				return
			}

			if err := createProblem(problem); err != nil {
				fmt.Println(err.Error())
				return
			}

			defer r.Body.Close()
			done <- true
		}),
	}

	go func() {
		<-done
		if err := server.Shutdown(context.Background()); err != nil {
			fmt.Println(err.Error())
		}
	}()

	fmt.Println("Waiting for you...")
	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}

	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "ru",
		Short: "Parse problems, contests and run test.",
	}

	var configCmd = &cobra.Command{
		Use:   "config",
		Short: "Setup compilation options",
		RunE: func(cmd *cobra.Command, args []string) error {
			return configSetup()
		},
	}

	var parseCmd = &cobra.Command{
		Use:   "parse",
		Short: "Parse a problem",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startServerAndParse()
		},
	}

	var testCmd = &cobra.Command{
		Use:   "test",
		Short: "Run tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			return testProblem()
		},
	}

	rootCmd.AddCommand(parseCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(configCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

func createProblem(problem Problem) error {
	fmt.Printf("Creating problem: %s ", problem.Name)
	// like A, B, C in Codeforces
	problemNameInitial := problem.Name[0]

	if err := os.Mkdir(fmt.Sprintf("%c", problemNameInitial), os.ModePerm); err != nil {
		return err
	}

	files := make([]*os.File, 0)

	for index, t := range problem.Tests {
		fileInput, err := os.Create(fmt.Sprintf("%c/in%d.txt", problemNameInitial, index))
		if err != nil {
			return err
		}
		_, err = fileInput.WriteString(t.Input)
		if err != nil {
			return err
		}

		fileOut, err := os.Create(fmt.Sprintf("%c/out%d.txt", problemNameInitial, index))
		if err != nil {
			return err
		}

		_, err = fileOut.WriteString(t.Output)
		if err != nil {
			return err
		}

		files = append(files, fileInput)
		files = append(files, fileOut)
	}

	for _, file := range files {
		err := file.Close()
		if err != nil {
			return err
		}
	}

	fmt.Println("✔️")
	return nil
}

var home string = os.Getenv("HOME")
var configDir string = fmt.Sprintf("%s/.config/ru.conf", home)

func testProblem() error {
	fmt.Println("Running tests...")
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	dirs := strings.Split(dir, "/")

	problemDir := dirs[len(dirs)-1]

	if len(problemDir) != 1 {
		return fmt.Errorf("error: you are not in a problem directory")
	}

	data, err := os.ReadFile(configDir)
	if err != nil {
		return err
	}

	cppCmd := strings.TrimSuffix(string(data), "\n")
	cppCmd = strings.TrimSpace(cppCmd)

	// compile program,
	// C++ compile command
	cmd := exec.Command(cppCmd, fmt.Sprintf("%s.cpp", problemDir))
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}

	// run tests cases
	// with a.out

	for i := 0; true; i++ {
		inFile := fmt.Sprintf("in%d.txt", i)

		outData, err := os.ReadFile(fmt.Sprintf("out%d.txt", i))
		if err != nil {
			// no such input file
			// we are done
			return nil
		}

		run := exec.Command("./a.out")

		inData, err := os.Open(inFile)
		if err != nil {
			// no such input file
			// we are done
			return nil
		}

		run.Stdin = inData

		output, err := run.Output()
		if err != nil {
			return err
		}

		if bytes.Equal(output, outData) {
			color.Green("PASSED")
		} else {
			c := color.RGB(152, 152, 152) // gray
			color.Red("FAILED\n")
			c.Println("Correct:")
			fmt.Println(string(outData))
			c.Println("Your Output:")
			fmt.Println(string(output))
		}
	}

	return nil
}

func configSetup() error {
	fmt.Print("Paste your c++ compile command: ")

	reader := bufio.NewReader(os.Stdin)
	cmd, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	//os.Stat(configDir)
	// create file
	f, err := os.Create(configDir)
	if err != nil {
		return err
	}

	_, err = f.WriteString(cmd)
	if err != nil {
		return err
	}

	fmt.Printf("C++ compilation command saved to: %s\n", configDir)
	return nil
}
