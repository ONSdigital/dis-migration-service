package migrator

import (
	"context"
	"strings"

	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/slack"
	"github.com/ONSdigital/log.go/v2/log"
)

// StateTransitionRule defines when a job should transition based on task states
type StateTransitionRule struct {
	// TaskTargetState is the state all tasks must reach
	TaskTargetState domain.State
	// JobTargetState is the state the job should transition to
	JobTargetState domain.State
	// Description explains the rule
	Description string
}

// GetStateTransitionRules returns all rules for job state transitions
// based on task completion
func (mig *migrator) GetStateTransitionRules() map[domain.State][]StateTransitionRule {
	return map[domain.State][]StateTransitionRule{
		domain.StateMigrating: {
			{
				TaskTargetState: domain.StateInReview,
				JobTargetState:  domain.StateInReview,
				Description:     "all tasks migrated, job moves to in_review",
			},
		},
		domain.StatePublishing: {
			{
				TaskTargetState: domain.StatePublished,
				JobTargetState:  domain.StatePublished,
				Description:     "all tasks published, job moves to published",
			},
		},
	}
}

// CheckAndUpdateJobStateBasedOnTasks checks if all tasks have reached
// the target state and updates the job accordingly
func (mig *migrator) CheckAndUpdateJobStateBasedOnTasks(ctx context.Context, jobNumber int, rule StateTransitionRule) error {
	logData := log.Data{
		"job_number":        jobNumber,
		"task_target_state": rule.TaskTargetState,
		"job_target_state":  rule.JobTargetState,
	}

	// Get the job
	job, err := mig.jobService.GetJob(ctx, jobNumber)
	if err != nil {
		log.Error(ctx, "failed to get job for state check", err, logData)
		return err
	}

	// Count tasks in the target state
	tasksInTargetState, err := mig.countTasksInState(
		ctx,
		jobNumber,
		rule.TaskTargetState,
	)
	if err != nil {
		log.Error(ctx, "failed to count tasks in target state", err, logData)
		return err
	}

	// Count total tasks (with no state filter to get all tasks)
	totalTasks, err := mig.jobService.CountTasksByJobNumber(ctx, jobNumber)
	if err != nil {
		log.Error(ctx, "failed to count total tasks", err, logData)
		return err
	}

	logData["tasks_in_target_state"] = tasksInTargetState
	logData["total_tasks"] = totalTasks

	// Check if all tasks have reached the target state
	if tasksInTargetState < totalTasks {
		log.Info(
			ctx,
			"not all tasks in target state, skipping job state update",
			logData,
		)
		return nil
	}

	log.Info(ctx, strings.ToLower(rule.Description), logData)

	// Update job state
	err = mig.jobService.UpdateJobState(ctx, job.JobNumber, rule.JobTargetState, "")
	if err != nil {
		log.Error(ctx, "failed to update job state", err, logData)
		return err
	}

	logData["new_state"] = rule.JobTargetState
	log.Info(ctx, "job state updated successfully", logData)

	// Send notification for completed active states
	if isActiveStateCompletion(job.State, rule.JobTargetState) {
		slackDetails := slack.SlackDetails{
			"Job Number": job.JobNumber,
			"Job Label":  job.Label,
			"Old State":  job.State,
			"New State":  rule.JobTargetState,
		}
		err = mig.slackClient.SendInfo(ctx, mig.getJobCompletionSummary(job.State, rule.JobTargetState), slackDetails)
		if err != nil {
			log.Error(ctx, "failed to send slack notification", err, logData)
		}
	}

	return nil
}

// countTasksInState counts how many tasks are in a specific state
func (mig *migrator) countTasksInState(ctx context.Context, jobNumber int, targetState domain.State) (int, error) {
	// Get count of tasks in the target state
	_, totalCount, err := mig.jobService.GetJobTasks(
		ctx,
		[]domain.State{targetState},
		jobNumber,
		1, // Minimal limit - we only need the count
		0,
	)
	if err != nil {
		return 0, err
	}

	return totalCount, nil
}

// TriggerJobStateTransitions checks all transition rules and
// updates job state if conditions are met
func (mig *migrator) TriggerJobStateTransitions(ctx context.Context, jobNumber int) error {
	job, err := mig.jobService.GetJob(ctx, jobNumber)
	if err != nil {
		return err
	}

	rules := mig.GetStateTransitionRules()[job.State]
	if len(rules) == 0 {
		return nil // No transitions available from current state
	}

	for _, rule := range rules {
		err := mig.CheckAndUpdateJobStateBasedOnTasks(ctx, jobNumber, rule)
		if err != nil {
			return err
		}
	}
	return nil
}

// getJobCompletionSummary returns a human-readable summary for state completion
func (mig *migrator) getJobCompletionSummary(fromState, toState domain.State) string {
	switch fromState {
	case domain.StateMigrating:
		return "Job migration completed successfully"
	case domain.StatePublishing:
		return "Job publishing completed successfully"
	case domain.StatePostPublishing:
		return "Job post-publishing completed successfully"
	default:
		return "Job state updated"
	}
}

// isActiveStateCompletion checks if the transition represents completion
// of an active processing state
func isActiveStateCompletion(fromState, toState domain.State) bool {
	// Check if we're completing one of the active processing states
	switch fromState {
	case domain.StateMigrating:
		return toState == domain.StateInReview
	case domain.StatePublishing:
		return toState == domain.StatePublished
	case domain.StatePostPublishing:
		return toState == domain.StateCompleted
	default:
		return false
	}
}
