package clients

import (
	datasetAPI "github.com/ONSdigital/dp-dataset-api/sdk"
	filesAPI "github.com/ONSdigital/dp-files-api/sdk"
	uploadService "github.com/ONSdigital/dp-upload-service/sdk"
)

// ClientList holds all the API clients used by the service.
type ClientList struct {
	DatasetAPI    datasetAPI.Clienter
	FilesAPI      filesAPI.Clienter
	RedirectAPI   RedirectAPIClient
	UploadService uploadService.Clienter
	Zebedee       ZebedeeClient
}
