package generated

// ContentType represents the mIME type of the file.
type ContentType string

const (
	ContentTypeApplicationPdf                                                     ContentType = "application/pdf"
	ContentTypeImageJpeg                                                          ContentType = "image/jpeg"
	ContentTypeImagePng                                                           ContentType = "image/png"
	ContentTypeApplicationMsword                                                  ContentType = "application/msword"
	ContentTypeApplicationVndOpenxmlformatsOfficedocumentWordprocessingmlDocument ContentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	ContentTypeApplicationVndMsExcel                                              ContentType = "application/vnd.ms-excel"
	ContentTypeApplicationVndOpenxmlformatsOfficedocumentSpreadsheetmlSheet       ContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	ContentTypeApplicationZip                                                     ContentType = "application/zip"
	ContentTypeImageTiff                                                          ContentType = "image/tiff"
	ContentTypeImageWebp                                                          ContentType = "image/webp"
	ContentTypeImageBmp                                                           ContentType = "image/bmp"
)

type DocumentControllerRequestUploadRequest struct {
	ContentType ContentType `json:"contentType,omitempty"`
	Filename    string      `json:"filename,omitempty"`
	Title       string      `json:"title,omitempty"`
}
