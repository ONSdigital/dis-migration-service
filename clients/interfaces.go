package clients

import (
	"context"
	"io"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
)

//go:generate moq -out mock/zebedee.go -pkg mock . ZebedeeClient

// ZebedeeClient is an interface defining the methods for the Zebedee
// (github.com/ONSdigital/zebedee) client.
type ZebedeeClient interface {
	ApproveCollection(ctx context.Context, authToken, collectionID string) error
	ApproveCollectionContent(ctx context.Context, authToken, collectionID, lang, pagePath string) error
	CheckCollectionsForURI(ctx context.Context, authToken, uri string) (string, bool, error)
	CompleteCollectionContent(ctx context.Context, authToken, collectionID, lang, pagePath string) error
	CreateCollection(ctx context.Context, authToken string, collection zebedee.Collection) (zebedee.Collection, error)
	DeleteCollection(ctx context.Context, userAuthToken, collectionID string) error
	DeleteCollectionContent(ctx context.Context, userAuthToken, collectionID, path string) error
	GetCollection(ctx context.Context, authToken, collectionID string) (zebedee.Collection, error)
	GetDataset(ctx context.Context, authToken, collectionID, lang, path string) (d zebedee.Dataset, err error)
	GetDatasetLandingPage(ctx context.Context, authToken, collectionID, lang, path string) (d zebedee.DatasetLandingPage, err error)
	GetFileSize(ctx context.Context, authToken, collectionID, lang, uri string) (f zebedee.FileSize, err error)
	GetPageData(ctx context.Context, authToken, collectionID, lang, path string) (m zebedee.PageData, err error)
	GetResourceStream(ctx context.Context, authToken, collectionID, lang, path string) (s io.ReadCloser, err error)
	PublishCollection(ctx context.Context, authToken, collectionID string) error
	SaveContentToCollection(ctx context.Context, authToken, collectionID, path string, content interface{}) error
}
