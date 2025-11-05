package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type JobDetails struct {
	Description string `json:"description"`
	S3Path      string `json:"s3Path"`
	Compute     string `json:"compute"`
}

func main() {
	jobsToRun := &[]JobDetails{}
	data, err := os.ReadFile("../gdea-cert.api")
	if err != nil {
		fmt.Printf("Error reading gdea-cert.api: %v\n", err)
		return
	}
	apiKey := string(data)
	data, err = os.ReadFile("./week-3.json")
	if err != nil {
		fmt.Printf("Error reading week-3.json: %v\n", err)
		return
	}
	err = json.Unmarshal(data, jobsToRun)
	if err != nil {
		fmt.Printf("Error unmarshalling jobsToRun: %v\n", err)
		return
	}

	for _, job := range *jobsToRun {
		err = RunJob(job, apiKey)
		if err != nil {
			fmt.Printf("Ending job! Failed processing %v.\n", job)
			return
		}
	}

	fmt.Println("Multi-job process complete!")
}

type PostRunBody struct {
	Runtime   string          `json:"runtime"`
	Name      string          `json:"name"`
	RunPython RunPythonObject `json:"runPython"`
	Timeout   int             `json:"timeoutSeconds"`
}

type RunPythonObject struct {
	Uri string `json:"uri"`
}

type PartialRunResponse struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func RunJob(job JobDetails, apiKey string) error {
	body := PostRunBody{
		Runtime: job.Compute,
		Name: fmt.Sprintf(
			"silver_%v",
			strings.ReplaceAll(job.Description, "-", "_"),
		),
		RunPython: RunPythonObject{
			Uri: job.S3Path,
		},
		Timeout: 3600,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("Error marshalling PostRunBody for %s: %v\n", job.Description, err)
		return err
	}

	fmt.Println(bytes.NewBuffer(jsonBody))

	postUrl := "https://api.cloud.wherobots.com/runs?region=aws-us-east-1"
	req, err := http.NewRequest("POST", postUrl, bytes.NewReader(jsonBody))
	if err != nil {
		fmt.Printf("Error creating POST request for %s: %v\n", job.Description, err)
		return err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: time.Duration(body.Timeout) * time.Second,
	}

	fmt.Printf("Starting %s job.\n", body.Name)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error executing POST request for %s: %v\n", job.Description, err)
		return err
	}
	defer resp.Body.Close()

	rBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading POST response body for %s: %v\n", job.Description, err)
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("Received status code of %d!\n", resp.StatusCode)
		return errors.New("RunJob: Unexpected status error!")
	}

	response := &PartialRunResponse{}
	err = json.Unmarshal(rBody, response)
	if err != nil {
		fmt.Printf("Unable to parse response body! %v\n", err)
		return err
	}
	id := response.Id
	fmt.Printf("%s (%s) - %s\n", job.Description, id, response.Status)

	getUrl := fmt.Sprintf("https://api.cloud.wherobots.com/runs/%s", id)

	for strings.ToLower(response.Status) != "completed" {
		req, err = http.NewRequest("GET", getUrl, nil)
		if err != nil {
			fmt.Printf("Error creating GET request for %s: %v\n", job.Description, err)
			return err
		}

		req.Header.Set("accept", "application/json")
		req.Header.Set("X-API-Key", apiKey)

		client = &http.Client{}

		resp, err = client.Do(req)
		if err != nil {
			fmt.Printf("Error executing GET request for %s: %v\n", job.Description, err)
			return err
		}
		defer resp.Body.Close()

		rBody, err = io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading GET response body for %s: %v\n", job.Description, err)
			return err
		}

		response = &PartialRunResponse{}
		err = json.Unmarshal(rBody, response)
		if err != nil {
			fmt.Printf("Unable to parse response body! %v\n", err)
			return err
		}

		fmt.Printf("%s - %s\n", job.Description, response.Status)

		time.Sleep(30 * time.Second)
	}

	return nil
}
