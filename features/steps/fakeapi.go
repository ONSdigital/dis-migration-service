package steps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	datasetModels "github.com/ONSdigital/dp-dataset-api/models"

	datasetError "github.com/ONSdigital/dp-dataset-api/apierrors"
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
	var body []byte
	path := fmt.Sprintf("/datasets/%s", id)

	switch statusCode {
	case http.StatusNotFound:
		body = []byte(datasetError.ErrDatasetNotFound.Error())
	case http.StatusOK:
		body, _ = json.Marshal(datasetModels.Dataset{
			ID: id,
		})
	}

	f.fakeHTTP.NewHandler().Get(path).Reply(statusCode).Body(body)
}
