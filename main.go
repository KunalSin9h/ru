package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/atotto/clipboard"
)

var gray = color.RGB(152, 152, 152) // gray

type Test struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type Problem struct {
	Name  string `json:"name"`
	Tests []Test `json:"tests"`
}

// Flag for copying code to clipboard
var copy bool

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
		Use:     "parse",
		Aliases: []string{"p"},
		Short:   "Parse a problem",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startServerAndParse()
		},
	}

	var testCmd = &cobra.Command{
		Use:     "test",
		Aliases: []string{"t"},
		Short:   "Run tests",
		RunE: func(cmd *cobra.Command, args []string) error {
			return testProblem()
		},
	}

	testCmd.PersistentFlags().BoolVarP(&copy, "copy", "c", false, "Copy solution to clipboard")

	rootCmd.AddCommand(parseCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(configCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
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

	fmt.Println()

	return nil
}

var home string = os.Getenv("HOME")
var configDir string = fmt.Sprintf("%s/.config/ru.conf", home)

func testProblem() error {
	gray.Println("Running tests...")
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	dirs := strings.Split(dir, "/")

	problemDir := dirs[len(dirs)-1]

	if len(problemDir) != 1 {
		return fmt.Errorf("you are not in a problem directory\n")
	}

	data, err := os.ReadFile(configDir)
	if err != nil {
		return err
	}

	solutionFile := fmt.Sprintf("%s.cpp", problemDir)

	cppCmd := strings.TrimSuffix(string(data), "\n")
	cppCmd = strings.TrimSpace(cppCmd)

	cmds := strings.Split(cppCmd, " ")
	cmds = append(cmds, solutionFile)

	// compile program,
	// C++ compile command
	cmd := exec.Command(cmds[0], cmds[1:]...)
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
		defer inData.Close()

		run.Stdin = inData
		var stdout, stderr bytes.Buffer
		run.Stdout = &stdout
		run.Stderr = &stderr

		if err := run.Run(); err != nil {
			gray.Println(stdout.String())
			color.Red(stderr.String())
			return err
		}

		output := stdout.Bytes()
		if bytes.Equal(output, outData) {
			color.HiGreen("PASSED")
			if copy {
				gray.Print("Copying solution... ")
				solutionSource, err := os.ReadFile(solutionFile)
				if err != nil {
					return err
				}
				if err := clipboard.WriteAll(string(solutionSource)); err != nil {
					return err
				}
				gray.Println("Done")
			}
		} else {
			color.HiRed("FAILED")
			gray.Println("Correct:")
			fmt.Println(string(outData))
			gray.Println("Your Output:")
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

	color.Green("C++ compilation command saved to: %s\n", configDir)
	return nil
}
