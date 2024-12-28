package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
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
	}

	rootCmd.AddCommand(parseCmd)
	rootCmd.AddCommand(testCmd)

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