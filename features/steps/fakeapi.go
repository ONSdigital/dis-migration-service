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
	fakeHTTP *httpfake.HTTPFake
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

func (f *FakeAPI) setFullJSONResponseForGetPageData(url string, statusCode int, payload string) {
	specialCharURL := strings.ReplaceAll(url, "/", "%2F")
	path := "/data?uri=" + specialCharURL + "&lang=en"
	f.fakeHTTP.NewHandler().Get(path).Reply(statusCode).BodyString(payload)
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
