package clients

import (
	"context"

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
	GetPageData(ctx context.Context, userAuthToken, collectionID, lang, path string) (m zebedee.PageData, err error)
}
