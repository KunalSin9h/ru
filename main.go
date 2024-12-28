package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Test struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type Problem struct {
	Name  string `json:"name"`
	Tests []Test `json:"tests"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
	})

	fmt.Println("Listening on :6174")
	if err := http.ListenAndServe(":6174", nil); err != nil {
		fmt.Println(err)
	}
}

func createProblem(problem Problem) error {
	fmt.Printf("Creating problem: %s\n", problem.Name)
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

	return nil
}
