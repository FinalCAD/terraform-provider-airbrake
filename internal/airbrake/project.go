package airbrake

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Notifier struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Version string `json:"version"`
}

type Project struct {
	Id                    int    `json:"id"`
	Name                  string `json:"name"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
	AccountID             int    `json:"account_id"`
	APIKey                string `json:"api_key"`
	ResolveErrorsOnDeploy bool   `json:"resolve_errors_on_deploy"`
	MinAppVersion         string `json:"min_app_version"`
	StrictErrorTypes      string `json:"strict_error_types"`
	GlobalErrorTypes      string `json:"global_error_types"`
	ExceptionalAppID      string `json:"exceptional_app_id"`
	SeverityThreshold     struct {
		Level string `json:"level"`
	} `json:"severity_threshold"`
	Language                  string   `json:"language"`
	RetentionPeriodDays       int      `json:"retention_period_days"`
	EnableOldGrouping         bool     `json:"enable_old_grouping"`
	FirstNoticeReceivedAt     string   `json:"first_notice_received_at"`
	ApdexThreshold            string   `json:"apdex_threshold"`
	NotifierName              string   `json:"notifier_name"`
	NotifierVersion           string   `json:"notifier_version"`
	AnomalyNotificationEnv    []string `json:"anomaly_notification_environments"`
	LastDeployAt              string   `json:"last_deploy_at"`
	ServerErrorAlertThreshold int      `json:"server_error_alert_threshold"`
	LastCommentAt             string   `json:"last_comment_at"`
	LastUserGroupResolvedAt   string   `json:"last_user_group_resolved_at"`
	IsFirstProject            bool     `json:"is_first_project"`
	DemoModeUntilDate         string   `json:"demo_mode_until_date"`
}

type Response struct {
	Projects []Project `json:"projects"`
}

func (c *Client) GetProjects(ctx context.Context) ([]Project, error) {
	var responseData Response
	status, response, err := c.Get("projects", nil)

	if err != nil || status >= 300 {
		return nil, fmt.Errorf("error during aibrake projects listing")
	}

	err = json.Unmarshal([]byte(response), &responseData)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %s", err.Error())
	}
	return responseData.Projects, nil
}

func (c *Client) GetProject(ctx context.Context, name string) (Project, error) {
	projects, err := c.GetProjects(ctx)
	if err != nil {
		return Project{}, err
	}
	for _, p := range projects {
		if p.Name == name {
			return p, nil
		}
	}
	return Project{}, fmt.Errorf("can't find project : %s", name)
}

func (c *Client) GetProjectById(ctx context.Context, id int) (Project, error) {
	projects, err := c.GetProjects(ctx)
	if err != nil {
		return Project{}, err
	}
	for _, p := range projects {
		if p.Id == id {
			return p, nil
		}
	}
	return Project{}, fmt.Errorf("can't find project : %d", id)
}

func (c *Client) CreateProject(ctx context.Context, p Project) (Project, error) {
	var responseData Project

	tflog.Debug(ctx, "create project airbrake", map[string]interface{}{
		"project": p,
	})

	params := map[string]string{
		"name": p.Name,
	}
	status, response, err := c.Post("projects", params)

	if err != nil || status >= 300 {
		return Project{}, fmt.Errorf("error during new aibrake project")
	}

	err = json.Unmarshal([]byte(response), &responseData)
	if err != nil {
		return Project{}, fmt.Errorf("error parsing JSON: %s", err.Error())
	}

	responseData.Language = p.Language
	tflog.Debug(ctx, "create project airbrake", map[string]interface{}{
		"responseData": responseData,
	})
	err = c.UpdateProject(ctx, responseData)
	if err != nil {
		return Project{}, err
	}

	return responseData, nil
}

func (c *Client) UpdateProject(ctx context.Context, p Project) error {
	values := map[string]string{
		"language": p.Language,
	}
	status, err := c.Put("projects", strconv.Itoa(p.Id), values)
	if err != nil || status >= 300 {
		return fmt.Errorf("error during aibrake project update")
	}
	return nil
}

func (c *Client) DeleteProject(ctx context.Context, id string) error {
	status, err := c.Delete("projects", id)
	if err != nil || status >= 300 {
		return fmt.Errorf("error during aibrake project delete")
	}
	return nil
}
