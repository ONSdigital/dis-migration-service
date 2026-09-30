package clients

import (
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetAPI "github.com/ONSdigital/dp-dataset-api/sdk/go"
	filesAPI "github.com/ONSdigital/dp-files-api/sdk"
	topicAPI "github.com/ONSdigital/dp-topic-api/sdk"
	uploadService "github.com/ONSdigital/dp-upload-service/sdk"
)

// ClientList holds all the API clients used by the service.
type ClientList struct {
	DatasetAPI    datasetAPI.Clienter
	FilesAPI      filesAPI.Clienter
	TopicAPI      topicAPI.Clienter
	UploadService uploadService.Clienter
	Zebedee       zebedee.Clienter
}
