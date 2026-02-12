package executor

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ONSdigital/dis-migration-service/application"
	"github.com/ONSdigital/dis-migration-service/clients"
	"github.com/ONSdigital/dis-migration-service/domain"
	appErrors "github.com/ONSdigital/dis-migration-service/errors"
	"github.com/ONSdigital/dis-migration-service/mapper"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetSDK "github.com/ONSdigital/dp-dataset-api/sdk"
	uploadSDK "github.com/ONSdigital/dp-upload-service/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

const (
	maxRetries = 5
	baseDelay  = 50 * time.Millisecond
)

// DatasetDownloadTaskExecutor executes migration tasks for dataset downloads.
type DatasetDownloadTaskExecutor struct {
	jobService       application.JobService
	clientList       *clients.ClientList
	serviceAuthToken string
}

// NewDatasetDownloadTaskExecutor creates a new DatasetDownloadTaskExecutor
func NewDatasetDownloadTaskExecutor(jobService application.JobService, clientList *clients.ClientList, serviceAuthToken string) *DatasetDownloadTaskExecutor {
	return &DatasetDownloadTaskExecutor{
		jobService:       jobService,
		clientList:       clientList,
		serviceAuthToken: serviceAuthToken,
	}
}

// Migrate handles the migration operations for a dataset download task.
func (e *DatasetDownloadTaskExecutor) Migrate(ctx context.Context, task *domain.Task) error {
	if task == nil || task.Source == nil || task.Target == nil || task.Source.ID == "" {
		err := appErrors.ErrInvalidTask
		log.Error(ctx, "invalid task or missing source/target information", err)
		return err
	}

	logData := log.Data{"task_id": task.ID, "job_number": task.JobNumber, "source_id": task.Source.ID}

	log.Info(ctx, "starting migration for dataset download task", logData)

	fileSize, err := e.clientList.Zebedee.GetFileSize(ctx, e.serviceAuthToken, zebedee.EmptyCollectionId, zebedee.EnglishLangCode, task.Source.ID)
	if err != nil {
		log.Error(ctx, "failed to get file size from zebedee", err, logData)
		return err
	}

	// Get resource stream from Zebedee
	resourceStream, err := e.clientList.Zebedee.GetResourceStream(ctx, e.serviceAuthToken, zebedee.EmptyCollectionId, zebedee.EnglishLangCode, task.Source.ID)
	if err != nil {
		log.Error(ctx, "failed to get file stream from zebedee", err, logData)
		return err
	}

	// Map to upload service metadata
	uploadMetadata, err := mapper.MapResourceToUploadServiceMetadata(task.Source.ID, fileSize)
	if err != nil {
		resourceStream.Close()
		log.Error(ctx, "failed to map upload service metadata", err, logData)
		return err
	}

	// Upload to upload service
	headers := uploadSDK.Headers{
		ServiceAuthToken: e.serviceAuthToken,
	}

	err = e.clientList.UploadService.Upload(ctx, resourceStream, uploadMetadata, headers)
	// Close stream immediately after upload attempt
	if closeErr := resourceStream.Close(); closeErr != nil {
		log.Error(ctx, "failed to close resource stream", closeErr, logData)
	}

	if err != nil {
		log.Error(ctx, "failed to upload file to upload service", err, logData)
		return appErrors.ErrFailedToUploadFileToUploadService
	}

	// Map to dataset distribution
	distribution, err := mapper.MapUploadServiceMetadataToDistribution(uploadMetadata)
	if err != nil {
		log.Error(ctx, "failed to map upload service metadata to dataset distribution", err, logData)
		return err
	}

	// Update dataset version with new distribution
	log.Info(ctx, "updating dataset version metadata with new distribution", logData)
	err = e.updateDownloadMetadata(ctx, task, distribution)
	if err != nil {
		log.Error(ctx, "failed to update dataset version metadata with new distribution", err, logData)
		return err
	}

	// Update task state to in review
	err = e.jobService.UpdateTaskState(ctx, task.ID, domain.StateInReview)
	if err != nil {
		log.Error(ctx, "failed to update migration task", err, logData)
		return err
	}

	log.Info(ctx, "completed migration for dataset download task", logData)
	return nil
}

// Publish handles the publish operations for a dataset download task.
func (e *DatasetDownloadTaskExecutor) Publish(ctx context.Context, task *domain.Task) error {
	// Implementation of publish for dataset download
	return nil
}

// PostPublish handles the post-publish operations for a dataset download task.
func (e *DatasetDownloadTaskExecutor) PostPublish(ctx context.Context, task *domain.Task) error {
	// Implementation of post-publish for dataset download
	return nil
}

// Revert handles the revert operations for a dataset download task.
func (e *DatasetDownloadTaskExecutor) Revert(ctx context.Context, task *domain.Task) error {
	// Implementation of revert for dataset download
	return nil
}

// updateDownloadMetadata updates the dataset version with a new distribution.
func (e *DatasetDownloadTaskExecutor) updateDownloadMetadata(ctx context.Context, task *domain.Task, distribution datasetModels.Distribution) error {
	// Validate target fields
	if task.Target.DatasetID == "" || task.Target.EditionID == "" || task.Target.VersionID == "" {
		return appErrors.ErrInvalidTask
	}

	logData := log.Data{
		"task_id":    task.ID,
		"dataset_id": task.Target.DatasetID,
		"edition_id": task.Target.EditionID,
		"version_id": task.Target.VersionID,
		"title":      distribution.Title,
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := getRandomDelay(baseDelay)

			log.Info(ctx, "ETag conflict, retrying after delay", log.Data{
				"task_id": task.ID,
				"attempt": attempt,
				"delay":   delay.String(),
			})

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := e.tryUpdateDownloadMetadata(ctx, task, distribution, logData)
		if err == nil {
			return nil
		}

		if !isConflictError(err) {
			return err
		}
	}

	return fmt.Errorf("failed to update download metadata after %d attempts", maxRetries)
}

// getRandomDelay returns a random duration between 0 and baseDelay
func getRandomDelay(baseDelay time.Duration) time.Duration {
	maxDelay := big.NewInt(int64(baseDelay))
	n, err := rand.Int(rand.Reader, maxDelay)
	if err != nil {
		return baseDelay
	}
	return time.Duration(n.Int64())
}

// tryUpdateDownloadMetadata attempts a single update
// of the dataset version with the distribution.
func (e *DatasetDownloadTaskExecutor) tryUpdateDownloadMetadata(ctx context.Context, task *domain.Task, distribution datasetModels.Distribution, logData log.Data) error {
	headers := datasetSDK.Headers{
		AccessToken: e.serviceAuthToken,
	}

	currentVersion, respHeaders, err := e.clientList.DatasetAPI.GetVersionWithHeaders(ctx, headers, task.Target.DatasetID, task.Target.EditionID, task.Target.VersionID)
	if err != nil {
		return err
	}

	e.applyDistributionUpdate(ctx, &currentVersion, distribution, logData)

	// Update version with eTag
	eTag := respHeaders.ETag
	log.Info(ctx, "updating dataset version with new distribution", log.Data{"task_id": task.ID, "e_tag": eTag})

	headers.IfMatch = eTag

	_, err = e.clientList.DatasetAPI.PutVersion(ctx, headers, task.Target.DatasetID, task.Target.EditionID, task.Target.VersionID, currentVersion)
	return err
}

// applyDistributionUpdate modifies the version's distributions
// with the new distribution.
func (e *DatasetDownloadTaskExecutor) applyDistributionUpdate(ctx context.Context, currentVersion *datasetModels.Version, distribution datasetModels.Distribution, logData log.Data) {
	// Initialize distributions slice if nil
	if currentVersion.Distributions == nil {
		distributions := []datasetModels.Distribution{distribution}
		currentVersion.Distributions = &distributions
		log.Info(ctx, "created new distributions array", logData)
	} else {
		// Find and update existing distribution, or append if new
		index := findDistributionIndexByTitle(
			*currentVersion.Distributions,
			distribution.Title,
		)

		if index >= 0 {
			// Distribution exists - enrich it with full metadata
			// Version task created this with Title + Format only
			// Now enriching with DownloadURL, ByteSize, and MediaType
			logData["title"] = distribution.Title
			logData["download_url"] = distribution.DownloadURL
			logData["byte_size"] = distribution.ByteSize
			logData["index"] = index
			log.Info(ctx, "enriching existing distribution with download metadata", logData)
			(*currentVersion.Distributions)[index] = distribution
		} else {
			// Distribution not found - append as new
			// This shouldn't normally happen if version task ran correctly
			log.Info(
				ctx,
				"distribution not found in version, appending as new",
				logData,
			)
			*currentVersion.Distributions = append(
				*currentVersion.Distributions,
				distribution,
			)
		}
	}
}

// isConflictError checks if the error is an HTTP 409 Conflict.
func isConflictError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), strconv.Itoa(http.StatusConflict))
}

// findDistributionIndexByTitle finds the index of a distribution in the slice
// based on Title. The Title is the stable identifier that links the partial
// distribution created by the version task with the full metadata added by
// the download task. Returns -1 if not found.
func findDistributionIndexByTitle(
	distributions []datasetModels.Distribution,
	title string,
) int {
	for i, d := range distributions {
		if d.Title == title {
			return i
		}
	}
	return -1
}
