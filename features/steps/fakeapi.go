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
	fakeHTTP             *httpfake.HTTPFake
	datasetCreateHandler *httpfake.Request
}

// NewFakeAPI creates a new fake component API
func NewFakeAPI() *FakeAPI {
	fakeAPI := httpfake.New()

	return &FakeAPI{
		fakeHTTP:             fakeAPI,
		datasetCreateHandler: fakeAPI.NewHandler().Post("/datasets"),
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

func (f *FakeAPI) setJSONResponseForCreateDataset(statusCode int) {
	f.datasetCreateHandler.Lock()
	defer f.datasetCreateHandler.Unlock()
	var body []byte

	switch statusCode {
	case http.StatusInternalServerError:
		body = []byte(datasetError.ErrInternalServer.Error())
	case http.StatusCreated:
		body = []byte(`{"_id": "new-dataset-id"}`)
	}

	createDatasetResponse := httpfake.NewResponse()
	createDatasetResponse.Status(statusCode)
	createDatasetResponse.Body(body)

	f.datasetCreateHandler.Response = createDatasetResponse
}
