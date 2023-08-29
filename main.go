package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sort"
	"time"
)

type Attributes struct {
	AssetIdentifier string `json:"asset_identifier"`
	CreatedAt       string `json:"created_at"`
}

type Data struct {
	Attributes Attributes `json:"attributes"`
}

type Links struct {
	Self string `json:"self"`
	Next string `json:"next"`
	Last string `json:"last"`
}

type Response struct {
	Data  []Data `json:"data"`
	Links Links  `json:"links"`
}

var (
	program  string
	username string
	apiKey   string
	showHelp bool
)

func fetchData(url string) (Response, error) {
	var response Response

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return response, err
	}

	req.SetBasicAuth(username, apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func fetchAllPages(baseURL string) ([]Data, error) {
	var allData []Data

	for baseURL != "" {
		response, err := fetchData(baseURL)
		if err != nil {
			return nil, err
		}
		allData = append(allData, response.Data...)

		baseURL = response.Links.Next
	}

	return allData, nil
}

func printCSVDescending(data []Data) {
	sort.SliceStable(data, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, data[i].Attributes.CreatedAt)
		timeJ, _ := time.Parse(time.RFC3339, data[j].Attributes.CreatedAt)
		return timeI.After(timeJ)
	})

	for _, d := range data {
		fmt.Printf("%s\n", d.Attributes.AssetIdentifier)
	}
}
func help() {}

func main() {

	flag.StringVar(&program, "p", "", "HackerOne program name")
	flag.StringVar(&username, "u", "", "HackerOne API username")
	flag.StringVar(&apiKey, "k", "", "HackerOne API Key")
	flag.BoolVar(&showHelp, "h", false, "Show help message")
	flag.Parse()

	if showHelp || program == "" || username == "" || apiKey == "" {
		asciiArt := `
		 _     _                               __      _       _               
		| |__ / |  ___  ___ ___  _ __   ___   / _| ___| |_ ___| |__   ___ _ __ 
		| '_ \| | / __|/ __/ _ \| '_ \ / _ \ | |_ / _ \ __/ __| '_ \ / _ \ '__|
		| | | | | \__ \ (_| (_) | |_) |  __/ |  _|  __/ || (__| | | |  __/ |   
		|_| |_|_| |___/\___\___/| .__/ \___| |_|  \___|\__\___|_| |_|\___|_|   
								|_|                                            
	`

		fmt.Println(asciiArt)
		fmt.Println("\"h1 scope fetcher\" is a tool to fetch all inscope assets of HackerOne programs for integration in your automation and hacking workflow.")
		fmt.Println("\nUseage:\n \t h1scopefetcher [Flags]")
		fmt.Println("Flags:\n \t -p\t\"Your Program Name\"\n \t -u\t\"HackerOne API username\"\n \t -k\t\"HackerOne API Key\"")

		return
	}

	baseURL := fmt.Sprintf("https://api.hackerone.com/v1/hackers/programs/%s/structured_scopes?page%%5Bsize%%5D=100", program)

	allData, err := fetchAllPages(baseURL)
	if err != nil {
		fmt.Println("Error fetching data:", err)
		return
	}

	printCSVDescending(allData)
}
