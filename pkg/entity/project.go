package entity

type Project struct {
	Name                 string `json:"name"`
	JiraComponentName    string `json:"jira_component_name"`
	GitUrl               string `json:"git_url"`
	ReleaseBranchExample string `json:"release_branch_example"`
	MainBranch           string `json:"main_branch"`
}
