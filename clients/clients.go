package clients

type ClientList struct {
	DatasetAPI    DatasetAPIClient
	FilesAPI      FilesAPIClient
	RedirectAPI   RedirectAPIClient
	UploadService UploadServiceClient
	Zebedee       ZebedeeClient
}
