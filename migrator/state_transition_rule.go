package migrator

import (
	"context"
	"errors"
	"fmt"

	"github.com/ONSdigital/dis-migration-service/domain"
	"github.com/ONSdigital/dis-migration-service/slack"
	"github.com/ONSdigital/log.go/v2/log"

	appErrors "github.com/ONSdigital/dis-migration-service/errors"
)

// StateTransitionRule defines when a job should transition based on task states
type StateTransitionRule struct {
	// TargetState is the state all tasks must reach
	TargetState domain.State
	// Description explains the rule
	Description string
	// FailureState is the state tasks should transition to if they fail
	FailureState domain.State
}

// GetStateTransitionRules returns all rules for job state transitions
// based on task completion
func (mig *migrator) GetStateTransitionRules() map[domain.State]StateTransitionRule {
	return map[domain.State]StateTransitionRule{
		domain.StateMigrating: {
			TargetState:  domain.StateInReview,
			FailureState: domain.StateFailedMigration,
			Description:  "migration successful, move to in review",
		},
		domain.StatePublishing: {
			TargetState:  domain.StatePublished,
			FailureState: domain.StateFailedPublish,
			Description:  "publish successful, move to published",
		},
		domain.StateReverting: {
			TargetState:  domain.StateRejected,
			FailureState: domain.StateFailedMigration,
			Description:  "revert successful, move to rejected",
		},
	}
}

// CheckAndUpdateJobStateBasedOnTasks checks if all tasks have reached
// the target state and updates the job accordingly
func (mig *migrator) CheckAndUpdateJobStateBasedOnTasks(ctx context.Context, jobNumber int, rule StateTransitionRule) error {
	logData := log.Data{
		"job_number":        jobNumber,
		"task_target_state": rule.TargetState,
		"job_target_state":  rule.TargetState,
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
		rule.TargetState,
	)
	if err != nil {
		log.Error(ctx, "failed to count tasks in target state", err, logData)
		return err
	}

	// Count tasks in the failure state, if defined
	tasksInFailureState, err := mig.countTasksInState(
		ctx,
		jobNumber,
		rule.FailureState,
	)
	if err != nil {
		log.Error(ctx, "failed to count tasks in failure state", err, logData)
		return err
	}

	// Count total tasks (with no state filter to get all tasks)
	totalTasks, err := mig.jobService.CountTasksByJobNumber(ctx, jobNumber)
	if err != nil {
		log.Error(ctx, "failed to count total tasks", err, logData)
		return err
	}

	tasksCompleted := tasksInTargetState + tasksInFailureState

	logData["tasks_in_target_state"] = tasksInTargetState
	logData["tasks_in_failure_state"] = tasksInFailureState
	logData["total_tasks"] = totalTasks

	// Check if all tasks have reached the target state
	if tasksCompleted < totalTasks {
		return nil
	}

	if tasksInFailureState > 0 {
		return mig.transitionJobFailure(ctx, job, rule, fmt.Sprintf("%d tasks failed out of %d", tasksInFailureState, totalTasks))
	}

	return mig.transitionJobSuccess(ctx, job, rule)
}

func (mig *migrator) transitionJobFailure(ctx context.Context, job *domain.Job, rule StateTransitionRule, failureReason string) error {
	log.Info(ctx, "transitioning job to failure state", log.Data{
		"job_number":     job.JobNumber,
		"job_state":      job.State,
		"failure_reason": failureReason,
	})

	transitioned, err := mig.transitionJob(ctx, job, rule.FailureState)
	if err != nil {
		log.Error(ctx, "failed to update job state", err)
		return err
	}

	if transitioned {
		log.Info(ctx, "job was transitioned - updating slack", log.Data{
			"job_number":     job.JobNumber,
			"job_state":      job.State,
			"failure_reason": failureReason,
		})
		slackDetails := slack.SlackDetails{
			"Job Number":     job.JobNumber,
			"Job Label":      job.Label,
			"Job State":      job.State,
			"Failure Reason": failureReason,
		}

		err = mig.slackClient.SendInfo(ctx, mig.getJobCompletionSummary(job.State, rule.FailureState), slackDetails, false)
		if err != nil {
			log.Error(ctx, "failed to send slack notification", err)
			// Not a critical failure - log and continue
		}
	}

	return nil
}

func (mig *migrator) transitionJobSuccess(ctx context.Context, job *domain.Job, rule StateTransitionRule) error {
	transitioned, err := mig.transitionJob(ctx, job, rule.TargetState)
	if err != nil {
		log.Error(ctx, "failed to update job state", err)
		return err
	}

	if transitioned {
		slackDetails := slack.SlackDetails{
			"Job Number": job.JobNumber,
			"Job Label":  job.Label,
			"Old State":  job.State,
			"New State":  rule.TargetState,
		}

		err = mig.slackClient.SendInfo(ctx, mig.getJobCompletionSummary(job.State, rule.TargetState), slackDetails, true)
		if err != nil {
			log.Error(ctx, "failed to send slack notification", err)
			// Not a critical failure - log and continue
		}
	}
	return nil
}

func (mig *migrator) transitionJob(ctx context.Context, job *domain.Job, targetState domain.State) (bool, error) {
	err := mig.jobService.UpdateJobState(ctx, job.JobNumber, targetState, "")
	if errors.Is(err, appErrors.ErrStateAlreadyAtTarget) {
		log.Info(ctx, "transitionJob: job is already in the target state, no transition needed", log.Data{
			"job_number": job.JobNumber,
			"state":      targetState,
		})
		return false, nil // Job is already in the target state, no transition needed
	} else if err != nil {
		log.Error(ctx, "failed to update job state", err)
		slackDetails := slack.SlackDetails{
			"Job Number":     job.JobNumber,
			"Job Label":      job.Label,
			"Old State":      job.State,
			"New State":      targetState,
			"Failure reason": fmt.Sprintf("failed to transition job to %s", targetState),
		}
		slackErr := mig.slackClient.SendAlarm(ctx, EventUpdateJobStateFailed, nil, slackDetails)
		if slackErr != nil {
			log.Error(ctx, "failed to send slack notification", slackErr)
			// Not a critical failure - log and continue
		}
		return false, err
	}
	return true, nil
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

	if job.State == domain.StateReverting {
		return nil
	}

	rule, ok := mig.GetStateTransitionRules()[job.State]
	if !ok {
		return nil // No transitions available from current state
	}

	err = mig.CheckAndUpdateJobStateBasedOnTasks(ctx, jobNumber, rule)
	if err != nil {
		return err
	}

	return nil
}

// getJobCompletionSummary returns a human-readable summary for state completion
func (mig *migrator) getJobCompletionSummary(fromState, toState domain.State) string {
	switch toState {
	case domain.StateInReview:
		return "Job migration completed successfully"
	case domain.StatePublished:
		return "Job publishing completed successfully"
	case domain.StateCompleted:
		return "Job post-publishing completed successfully"
	case domain.StateFailedMigration:
		return "Job migration failed"
	case domain.StateFailedPublish:
		return "Job publishing failed"
	case domain.StateFailedPostPublish:
		return "Job post-publishing failed"
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
