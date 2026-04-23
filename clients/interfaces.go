package clients

import (
	"context"
	"io"

	redirectModels "github.com/ONSdigital/dis-redirect-api/models"
	redirectSDK "github.com/ONSdigital/dis-redirect-api/sdk/go"
	redirectErrors "github.com/ONSdigital/dis-redirect-api/sdk/go/errors"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
)

//go:generate moq -out mock/redirect.go -pkg mock . RedirectAPIClient
//go:generate moq -out mock/zebedee.go -pkg mock . ZebedeeClient

// RedirectAPIClient is an interface defining the methods for the
// Redirect API (github.com/ONSdigital/dis-redirect-api) client.
// TODO: this interface should live in the dis-redirect-api repo
type RedirectAPIClient interface {
	PutRedirect(ctx context.Context, options redirectSDK.Options, id string, payload redirectModels.Redirect) redirectErrors.Error
}

// ZebedeeClient is an interface defining the methods for the Zebedee
// (github.com/ONSdigital/zebedee) client.
type ZebedeeClient interface {
	CreateCollection(ctx context.Context, authToken string, collection zebedee.Collection) (zebedee.Collection, error)
	GetDataset(ctx context.Context, authToken, collectionID, lang, path string) (d zebedee.Dataset, err error)
	GetDatasetLandingPage(ctx context.Context, authToken, collectionID, lang, path string) (d zebedee.DatasetLandingPage, err error)
	GetFileSize(ctx context.Context, authToken, collectionID, lang, uri string) (f zebedee.FileSize, err error)
	GetPageData(ctx context.Context, authToken, collectionID, lang, path string) (m zebedee.PageData, err error)
	GetResourceStream(ctx context.Context, authToken, collectionID, lang, path string) (s io.ReadCloser, err error)
	SaveContentToCollection(ctx context.Context, authToken, collectionID, path string, content interface{}) error
}
