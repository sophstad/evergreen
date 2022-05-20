package data

import (
	"fmt"
	"net/http"

	"github.com/evergreen-ci/evergreen/model"
	serviceModel "github.com/evergreen-ci/evergreen/model"
	"github.com/evergreen-ci/evergreen/model/task"
	"github.com/evergreen-ci/gimlet"
	"github.com/pkg/errors"
)

// FindTasksByBuildId uses the service layer's task type to query the backing database for a
// list of task that matches buildId. It accepts the startTaskId and a limit
// to allow for pagination of the queries. It returns results sorted by taskId.
func FindTasksByBuildId(buildId, taskId, status string, limit int, sortDir int) ([]task.Task, error) {
	pipeline := task.TasksByBuildIdPipeline(buildId, taskId, status, limit, sortDir)
	res := []task.Task{}

	err := task.Aggregate(pipeline, &res)
	if err != nil {
		return []task.Task{}, err
	}

	if taskId != "" {
		found := false
		for _, t := range res {
			if t.Id == taskId {
				found = true
				break
			}
		}
		if !found {
			return []task.Task{}, gimlet.ErrorResponse{
				StatusCode: http.StatusNotFound,
				Message:    fmt.Sprintf("task '%s' not found", taskId),
			}
		}
	}
	return res, nil
}

// FindTasksByProjectAndCommit is a method to find a set of tasks which ran as part of
// certain version in a project. It takes the projectId, commit hash, and a taskId
// for paginating through the results.
func FindTasksByProjectAndCommit(opts task.GetTasksByProjectAndCommitOptions) ([]task.Task, error) {
	projectId, err := model.GetIdForProject(opts.Project)
	if err != nil {
		return nil, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    err.Error(),
		}
	}

	pipeline := task.TasksByProjectAndCommitPipeline(opts)

	res := []task.Task{}
	err = task.Aggregate(pipeline, &res)
	if err != nil {
		return []task.Task{}, err
	}
	if len(res) == 0 {
		var message string
		if opts.Status != "" {
			message = fmt.Sprintf("task from project '%s' and commit '%s' with status '%s' "+
				"not found", projectId, opts.CommitHash, opts.Status)
		} else {
			message = fmt.Sprintf("task from project '%s' and commit '%s' not found",
				projectId, opts.CommitHash)
		}
		return []task.Task{}, gimlet.ErrorResponse{
			StatusCode: http.StatusNotFound,
			Message:    message,
		}
	}

	if opts.StartingTaskId != "" {
		found := false
		for _, t := range res {
			if t.Id == opts.StartingTaskId {
				found = true
				break
			}
		}
		if !found {
			return []task.Task{}, gimlet.ErrorResponse{
				StatusCode: http.StatusNotFound,
				Message:    fmt.Sprintf("task '%s' not found", opts.StartingTaskId),
			}
		}
	}
	return res, nil
}

func CheckTaskSecret(taskID string, r *http.Request) (int, error) {
	_, code, err := serviceModel.ValidateTask(taskID, true, r)
	if code == http.StatusConflict {
		if err == nil {
			err = errors.Errorf("conflict for task '%s'", taskID)
		}
		return http.StatusUnauthorized, errors.Wrapf(err, "invalid task '%s'", taskID)
	}
	return code, errors.WithStack(err)
}
