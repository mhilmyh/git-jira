package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type JiraTicketInformation struct {
	Title       string
	Description string
}

func main() {
	log.SetFlags(0)

	config, err := ReadConfigFile("./config.yaml")
	if err != nil {
		log.Fatalln(err.Error())
	}

	ticketNumber, err := StringPrompt(fmt.Sprintf("Enter %s Number: ", config.GetProjectKey()))
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.SetPrefix(fmt.Sprintf("[%s-%s] ", config.GetProjectKey(), ticketNumber))

	ticketInfo, err := QueryingJiraTicketData(config, ticketNumber)
	if err != nil {
		log.Fatalln(err.Error())
	}
	err = CreateGitCommit(ticketInfo)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func StringPrompt(text string) (string, error) {
	_, err := fmt.Fprint(os.Stdout, text)
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	sanitizedInput := strings.TrimSpace(input)
	if re := regexp.MustCompile("^\\d+$"); !re.MatchString(sanitizedInput) {
		return "", errors.New("input must be a number")
	}
	return sanitizedInput, nil
}

func QueryingJiraTicketData(conf *Config, ticketNumber string) (*JiraTicketInformation, error) {
	var err error
	var req *http.Request
	var resp *http.Response

	httpClient := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	url := fmt.Sprintf("https://%s.atlassian.net/rest/api/%s/issue/%s-%s", conf.GetOrganization(), conf.GetJiraApiVersion(), conf.GetProjectKey(), ticketNumber)
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(conf.GetUserEmail(), conf.GetPersonalAccessToken())
	resp, err = httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}(resp.Body)

	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type schema struct {
		Fields struct {
			Summary     string `json:"summary"`
			Description string `json:"description"`
		}
		ErrorMessages []string `json:"errorMessages"`
		Error         string   `json:"error"`
	}
	var obj schema
	err = json.Unmarshal(body, &obj)
	if err != nil {
		return nil, err
	}

	if len(obj.ErrorMessages) != 0 {
		return nil, errors.New(strings.ToLower(obj.ErrorMessages[0]))
	}
	if obj.Error != "" {
		return nil, errors.New(strings.ToLower(obj.Error))
	}

	return &JiraTicketInformation{
		Title:       obj.Fields.Summary,
		Description: obj.Fields.Description,
	}, nil
}

func CreateGitCommit(info *JiraTicketInformation) error {
	cmd := exec.Command("git", "commit", "-m", info.Title, "-m", info.Description)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
