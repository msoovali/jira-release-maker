package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/msoovali/jira-release-maker/pkg/config"
	"github.com/msoovali/jira-release-maker/pkg/entity"
)

const (
	JIRA_V3_REST_URL       = "https://%s.atlassian.net/rest/api/3"
	JIRA_CREATE_ISSUE_JSON = "/jira-create-issue.json"
	SUMMARY                = "{{SUMMARY}}"
	COMPONENT              = "{{COMPONENT}}"
	TEST_RELEASE_DATE      = "{{TEST_RELEASE_DATE}}"
	PROJECT_NAME           = "{{PROJECT_NAME}}"
)

type jiraClient struct {
	httpClient    *http.Client
	authorization string
	jiraV3RestUrl string
}

type jiraSearchResponse struct {
	Total  int            `json:"total"`
	Issues []jiraResource `json:"issues"`
}

type jiraResource struct {
	Id  string `json:"id"`
	Key string `json:"key"`
}

func NewClient(authorizationToken string, atlassianDomain string) *jiraClient {
	client := &http.Client{Timeout: time.Duration(10 * time.Second)}
	return &jiraClient{httpClient: client, authorization: authorizationToken, jiraV3RestUrl: fmt.Sprintf(JIRA_V3_REST_URL, atlassianDomain)}
}

func (client jiraClient) ReleaseExists(summary string) (existingReleaseKey string, err error) {
	searchResponse, err := client.getJiraIssuesBySummary(summary)
	if err != nil {
		return "", err
	}
	if searchResponse.Total > 0 {
		return searchResponse.Issues[0].Key, nil
	}

	return "", nil
}

func (client jiraClient) getJiraIssuesBySummary(summary string) (*jiraSearchResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/search?jql=summary~%s", client.jiraV3RestUrl, summary), nil)
	if err != nil {
		return nil, err
	}
	client.addBasicHeader(req)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("unexpected status code: %d. Response body: %s", resp.StatusCode, string(body))
	}
	var response jiraSearchResponse
	err = json.NewDecoder(resp.Body).Decode(&response)

	return &response, err
}

func (client jiraClient) addBasicHeader(request *http.Request) {
	request.Header.Add("Authorization", fmt.Sprintf("Basic %s", client.authorization))
}

func (client jiraClient) addApplicationJsonContentTypeHeader(request *http.Request) {
	request.Header.Add("Content-Type", "application/json")
}

func (client jiraClient) CreateInitialTestRelease(project entity.Project, summary string) (releaseKey string, err error) {
	createIssueContentPath := config.CONF_DIR + JIRA_CREATE_ISSUE_JSON
	createIssueBytes := config.LoadFileAndReadContent(createIssueContentPath)
	if createIssueBytes == nil {
		return "", fmt.Errorf("cannot read jira create issue body from %s", createIssueContentPath)
	}
	now := time.Now()
	requestBody := string(createIssueBytes)
	requestBody = strings.ReplaceAll(requestBody, SUMMARY, summary)
	requestBody = strings.ReplaceAll(requestBody, COMPONENT, project.JiraComponentName)
	requestBody = strings.ReplaceAll(requestBody, PROJECT_NAME, project.Name)
	requestBody = strings.ReplaceAll(requestBody, TEST_RELEASE_DATE, fmt.Sprintf("%d-%02d-%02d", now.Year(), now.Month(), now.Day()))
	fmt.Printf("%s\n", requestBody)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/issue", client.jiraV3RestUrl), bytes.NewBufferString(requestBody))
	if err != nil {
		return "", err
	}
	client.addBasicHeader(req)
	client.addApplicationJsonContentTypeHeader(req)
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("unexpected status code: %d. Response body: %s", resp.StatusCode, string(body))
	}
	var response jiraResource
	err = json.NewDecoder(resp.Body).Decode(&response)

	return response.Key, nil
}
