package steps

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	datasetModels "github.com/ONSdigital/dp-dataset-api/models"

	datasetError "github.com/ONSdigital/dp-dataset-api/apierrors"
	"github.com/maxcnunes/httpfake"
)

const (
	testCollectionID = "migration-job-test-collection"
)

// FakeAPI contains all the information for a fake component API
type FakeAPI struct {
	fakeHTTP                         *httpfake.HTTPFake
	datasetCreateHandler             *httpfake.Request
	collectionCreateHandler          *httpfake.Request
	collectionUpdateHandler          *httpfake.Request
	collectionContentCompleteHandler *httpfake.Request
	collectionContentApproveHandler  *httpfake.Request
	collectionGetHandler             *httpfake.Request
	collectionApproveHandler         *httpfake.Request
	collectionDetailsHandler         *httpfake.Request
	collectionPublishHandler         *httpfake.Request
}

// NewFakeAPI creates a new fake component API
func NewFakeAPI() *FakeAPI {
	fakeAPI := httpfake.New()

	// These are setting success criteria for collection interactions with zebedee.
	// To control this from component tests you will need to implement steps to update
	// these responses.
	collectionContentCompleteHandler := fakeAPI.NewHandler().Post(fmt.Sprintf("/complete/%s", testCollectionID))
	collectionContentCompleteHandler.Reply(200)

	collectionContentApproveHandler := fakeAPI.NewHandler().Post(fmt.Sprintf("/review/%s", testCollectionID))
	collectionContentApproveHandler.Reply(200)

	// This is setting success criteria for collection interactions with zebedee.
	// To control this from component tests you will need to implement steps to update
	// these responses.
	collectionCreateHandler := fakeAPI.NewHandler().Post("/collection")
	collectionCreateHandler.Reply(200).BodyStruct(zebedee.Collection{
		ID: testCollectionID,
	})

	collectionUpdateHandler := fakeAPI.NewHandler().Post(fmt.Sprintf("/content/%s", testCollectionID))
	collectionUpdateHandler.Reply(200)

	collectionGetHandler := fakeAPI.NewHandler().Get(fmt.Sprintf("/collection/%s", testCollectionID))
	collectionGetHandler.Reply(200).BodyStruct(zebedee.Collection{
		ID:             testCollectionID,
		ApprovalStatus: "COMPLETE",
	})

	collectionApproveHandler := fakeAPI.NewHandler().Post(fmt.Sprintf("/approve/%s", testCollectionID))
	collectionApproveHandler.Reply(200)

	collectionDetailsHandler := fakeAPI.NewHandler().Get(fmt.Sprintf("/collectionDetails/%s", testCollectionID))
	collectionDetailsHandler.Reply(200).BodyStruct(zebedee.Collection{
		ID:             testCollectionID,
		ApprovalStatus: "COMPLETE",
	})

	collectionPublishHandler := fakeAPI.NewHandler().Post(fmt.Sprintf("/publish/%s", testCollectionID))
	collectionPublishHandler.Reply(200)

	return &FakeAPI{
		fakeHTTP:                         fakeAPI,
		datasetCreateHandler:             fakeAPI.NewHandler().Post("/datasets"),
		collectionCreateHandler:          collectionCreateHandler,
		collectionUpdateHandler:          collectionUpdateHandler,
		collectionContentCompleteHandler: collectionContentCompleteHandler,
		collectionContentApproveHandler:  collectionContentApproveHandler,
		collectionGetHandler:             collectionGetHandler,
		collectionApproveHandler:         collectionApproveHandler,
		collectionDetailsHandler:         collectionDetailsHandler,
		collectionPublishHandler:         collectionPublishHandler,
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
