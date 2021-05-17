package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/msoovali/jira-release-maker/pkg/entity"
	"github.com/msoovali/jira-release-maker/pkg/prompt"
)

const (
	CONF_DIR      = "conf"
	SETTINGS_JSON = "/settings.json"
)

type settings struct {
	AtlassianDomain string           `json:"atlassian_domain"`
	Projects        []entity.Project `json:"projects"`
	ApiToken        string           `json:"jira_api_token"`
}

func LoadConfig() (*settings, error) {
	if !configsInPlace() {
		return nil, fmt.Errorf("config directory not found. Look documentation and fill necessary config files")
	}
	settings := loadSettings(CONF_DIR + SETTINGS_JSON)
	if settings == nil {
		return nil, fmt.Errorf("couldn't read settings. Check Your config")
	}
	if !projectNamesUnique(settings.Projects) {
		return nil, fmt.Errorf("config error. Projects array contains non-unique project names")
	}
	if settings.ApiToken == "" {
		token, err := askJiraAuthAndCreateToken()
		if err != nil {
			return nil, err
		}
		settings.ApiToken = token
		data, err := json.MarshalIndent(settings, "", "  ")
		if err != nil {
			return nil, err
		}
		saveToFile(CONF_DIR+SETTINGS_JSON, data)
	}

	return settings, nil
}

func projectNamesUnique(projects []entity.Project) bool {
	visited := map[string]bool{}
	for _, project := range projects {
		if visited[project.Name] {
			fmt.Printf("Error! Project name %s is configured multiple times in projects config file", project.Name)
			return false
		}
		visited[project.Name] = true
	}
	return true
}

func loadSettings(filePath string) *settings {
	byteValue := LoadFileAndReadContent(filePath)
	if byteValue == nil {
		return nil
	}
	var settings settings
	err := json.Unmarshal(byteValue, &settings)
	if err != nil {
		fmt.Printf("Couldn't unmarshal file (%s) contents: %v\n", filePath, err)
		return nil
	}

	return &settings
}

func createBasicToken(username string, password string) string {
	token := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(token))
}

func askJiraAuthAndCreateToken() (string, error) {
	email, err := prompt.AskForInput("Enter Your Jira email")
	if err != nil {
		return "", err
	}
	apiToken, err := prompt.AskForInput("Enter Jira API token (https://id.atlassian.com/manage-profile/security/api-tokens)")
	if err != nil {
		return "", err
	}

	return createBasicToken(email, apiToken), nil
}

func saveToFile(filePath string, data []byte) {
	_ = ioutil.WriteFile(filePath, data, os.ModePerm)
}

func LoadFileAndReadContent(filePath string) []byte {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return nil
	}
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Couldn't open file(%s)! %v\n", filePath, err)
		return nil
	}
	defer file.Close()
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file(%s) content! %v\n", filePath, err)
		return nil
	}

	return byteValue
}

func configsInPlace() bool {
	_, err := os.Stat(CONF_DIR)
	if os.IsNotExist(err) {
		fmt.Printf("Config is empty, creating new %s directory\n", CONF_DIR)
		os.Mkdir(CONF_DIR, os.ModePerm)
		ioutil.WriteFile(CONF_DIR+"/"+SETTINGS_JSON, nil, os.ModePerm)
		return false
	}
	return true
}

func (settings settings) GetProjectsNames() []string {
	var projectsNames []string
	for _, project := range settings.Projects {
		projectsNames = append(projectsNames, project.Name)
	}

	return projectsNames
}

func (settings settings) GetProjectByName(name string) *entity.Project {
	for _, project := range settings.Projects {
		if project.Name == name {
			return &project
		}
	}
	return nil
}
