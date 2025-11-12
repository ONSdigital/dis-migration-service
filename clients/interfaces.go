package clients

import (
	"context"
	"io"

	redirectModels "github.com/ONSdigital/dis-redirect-api/models"
	redirectSDK "github.com/ONSdigital/dis-redirect-api/sdk/go"
	redirectErrors "github.com/ONSdigital/dis-redirect-api/sdk/go/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-api-clients-go/v2/upload"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"
	datasetSDK "github.com/ONSdigital/dp-dataset-api/sdk"
)

//go:generate moq -out mock/dataset.go -pkg mock . DatasetAPIClient
//go:generate moq -out mock/files.go -pkg mock . FilesAPIClient
//go:generate moq -out mock/redirect.go -pkg mock . RedirectAPIClient
//go:generate moq -out mock/upload.go -pkg mock . UploadServiceClient
//go:generate moq -out mock/zebedee.go -pkg mock . ZebedeeClient

// DatasetAPIClient is an interface defining the methods for the
// Dataset API (github.com/ONSdigital/dp-dataset-api) client.
type DatasetAPIClient interface {
	GetDataset(ctx context.Context, headers datasetSDK.Headers, collectionID, datasetID string) (dataset datasetModels.Dataset, err error)
}

// FilesAPIClient is an interface defining the methods for the
// Files API (github.com/ONSdigital/dp-files-api) client.
type FilesAPIClient interface {
	GetFile(ctx context.Context, path string, authToken string) (files.FileMetaData, error)
}

// RedirectAPIClient is an interface defining the methods for the
// Redirect API (github.com/ONSdigital/dis-redirect-api) client.
type RedirectAPIClient interface {
	PutRedirect(ctx context.Context, options redirectSDK.Options, id string, payload redirectModels.Redirect) redirectErrors.Error
}

// UploadServiceClient is an interface defining the methods for the
// Upload Service (github.com/ONSdigital/dp-upload-service) client.
type UploadServiceClient interface {
	Upload(ctx context.Context, fileContent io.ReadCloser, metadata upload.Metadata) error
}

// ZebedeeClient is an interface defining the methods for the Zebedee
// (github.com/ONSdigital/zebedee) client.
type ZebedeeClient interface {
	GetPageData(ctx context.Context, userAuthToken, collectionID, lang, path string) (m zebedee.PageData, err error)
}
