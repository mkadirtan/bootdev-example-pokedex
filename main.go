package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

func getCommands() map[string]cliCommand {
	mapFunc, mapbFunc := getLocationFunctions()
	return map[string]cliCommand{
		"map": {
			name:        "map",
			description: "display the next 20 locations",
			callback:    mapFunc,
		},
		"mapb": {
			name:        "mapb",
			description: "display the previous 20 locations, errors when you are on the first page",
			callback:    mapbFunc,
		},
	}
}

func main() {
	commands := getCommands()
	for {
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("error: ", err.Error())
			continue
		}
		text = strings.TrimRight(text, "\n")
		command, ok := commands[text]
		if !ok {
			fmt.Printf("command: %v not found\n", text)
			continue
		}

		err = command.callback()
		if err != nil {
			fmt.Printf("encountered errors while executing %v: %v\n", text, err.Error())
		}
	}
}

func getLocationFunctions() (func() error, func() error) {
	type LocationResponse struct {
		Count    int     `json:"count"`
		Next     *string `json:"next"`
		Previous *string `json:"previous"`
		Results  []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"results"`
	}

	var curr *LocationResponse = nil

	return func() error {
			reqUrl := "https://pokeapi.co/api/v2/location/"

			if curr != nil && curr.Next != nil {
				reqUrl = *curr.Next
			}
			resp, err := http.Get(reqUrl)
			if err != nil {
				return err
			}

			bodyB, err := io.ReadAll(resp.Body)
			var locResp LocationResponse
			err = json.Unmarshal(bodyB, &locResp)
			if err != nil {
				return err
			}

			curr = &locResp
			for _, r := range locResp.Results {
				fmt.Println(r.Name)
			}

			return nil
		}, func() error {

			if curr == nil || curr.Previous == nil {
				return errors.New("no more prev pages")
			}
			reqUrl := *curr.Previous
			resp, err := http.Get(reqUrl)
			if err != nil {
				return err
			}

			bodyB, err := io.ReadAll(resp.Body)
			var locResp LocationResponse
			err = json.Unmarshal(bodyB, &locResp)
			if err != nil {
				return err
			}

			curr = &locResp
			for _, r := range locResp.Results {
				fmt.Println(r.Name)
			}

			return nil
		}
}
