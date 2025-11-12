package clients

// ClientList holds all the API clients used by the service.
type ClientList struct {
	DatasetAPI    DatasetAPIClient
	FilesAPI      FilesAPIClient
	RedirectAPI   RedirectAPIClient
	UploadService UploadServiceClient
	Zebedee       ZebedeeClient
}
