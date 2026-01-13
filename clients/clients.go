package clients

import (
	datasetAPI "github.com/ONSdigital/dp-dataset-api/sdk"
	filesAPI "github.com/ONSdigital/dp-files-api/sdk"
	topicAPI "github.com/ONSdigital/dp-topic-api/sdk"
	uploadService "github.com/ONSdigital/dp-upload-service/sdk"
)

// ClientList holds all the API clients used by the service.
type ClientList struct {
	DatasetAPI    datasetAPI.Clienter
	FilesAPI      filesAPI.Clienter
	RedirectAPI   RedirectAPIClient
	TopicAPI      topicAPI.Clienter
	UploadService uploadService.Clienter
	Zebedee       ZebedeeClient
}
