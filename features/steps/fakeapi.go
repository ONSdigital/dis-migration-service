package steps

import (
	"context"
	"fmt"
	"strings"

	datasetError "github.com/ONSdigital/dp-dataset-api/apierrors"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/maxcnunes/httpfake"
)

// FakeAPI contains all the information for a fake component API
type FakeAPI struct {
	fakeHTTP        *httpfake.HTTPFake
	pageDataRequest *httpfake.Request
}

// NewFakeAPI creates a new fake component API
func NewFakeAPI() *FakeAPI {
	return &FakeAPI{
		fakeHTTP: httpfake.New(),
	}
}

// Close closes the fake API
func (f *FakeAPI) Close() {
	f.fakeHTTP.Close()
}

func (f *FakeAPI) setJSONResponseForGetPageData(url, pageType string, statusCode int) {
	specialCharURL := strings.Replace(url, "/", "%2F", -1)
	path := "/data?uri=" + specialCharURL + "&lang=en"
	bodyStr := `{}`
	if pageType != "" {
		bodyStr = `{"type": "` + pageType + `", "description": {"title": "Labour Market statistics", "edition": "March 2024"}`
		bodyStr += "}"
	}
	f.fakeHTTP.NewHandler().Get(path).Reply(statusCode).BodyString(bodyStr)
}

func (f *FakeAPI) setJSONResponseForGetDataset(id string, statusCode int) {
	path := fmt.Sprintf("/datasets/%s", id)
	err := datasetError.ErrDatasetNotFound.Error()
	log.Info(context.TODO(), "err issued", log.Data{
		"err": err,
	})
	f.fakeHTTP.NewHandler().Get(path).Reply(statusCode).BodyString(err)
}
