package models

import "time"

const (
	//NotFoundError error from server
	NotFoundError string = "Not found"
	//ActionNotAllowed error from server
	ActionNotAllowed string = "Action not allowed"
	//WrongLength error from server
	WrongLength string = "Wrong length"
	//ServerError error from server
	ServerError string = "Server Error"
	//WrongInputFormatError wrong user input
	WrongInputFormatError string = "Wrong inputFormat!"
	//InvalidTokenError token is not valid
	InvalidTokenError string = "Token not valid"
	//InvalidCallbackURL token is not valid
	InvalidCallbackURL string = "Callback url is invalid"
	//BatchSizeTooLarge batch is too large
	BatchSizeTooLarge string = "BatchSize soo large!"
	//WrongIntegerFormat integer is probably no integer
	WrongIntegerFormat string = "Number is string"
	//MultipleSourceNameErr err name already exists
	MultipleSourceNameErr string = "You can't have multiple sources with the same name"
	//UserIsInvalidErr err if user is invalid
	UserIsInvalidErr string = "user is invalid"
)

//ResponseStatus the status of response
type ResponseStatus uint8

const (
	//ResponseError if there was an error
	ResponseError ResponseStatus = 0
	//ResponseSuccess if the response is successful
	ResponseSuccess ResponseStatus = 1
)

const (
	//HeaderStatus headername for status in response
	HeaderStatus string = "X-Response-Status"
	//HeaderStatusMessage headername for status in response
	HeaderStatusMessage string = "X-Response-Message"
	//HeaderContentType contenttype of response
	HeaderContentType string = "Content-Type"
	//HeaderFileName filename header
	HeaderFileName string = "X-Filename"
)

//StringResponse response containing only one string
type StringResponse struct {
	String string `json:"content"`
}

//StringSliceResponse response containing only one string slice
type StringSliceResponse struct {
	Slice []string `json:"slice"`
}

//FileResponseItem file item for file response
type FileResponseItem struct {
	ID           uint           `json:"id"`
	Size         int64          `json:"size"`
	CreationDate time.Time      `json:"creation"`
	Name         string         `json:"name"`
	PublicName   string         `json:"pubname"`
	IsPublic     bool           `json:"isPub"`
	Attributes   FileAttributes `json:"attrib"`
}

//PublishResponse response for publishing a file
type PublishResponse struct {
	PublicFilename string `json:"pubName"`
}

//BulkPublishResponse response for publishing a file
type BulkPublishResponse struct {
	Files []UploadResponse `json:"files"`
}

//ListFileResponse response for list files
type ListFileResponse struct {
	Files []FileResponseItem `json:"files"`
}

//UploadResponse response for uploading file
type UploadResponse struct {
	FileID         uint   `json:"fileID"`
	Filename       string `json:"filename"`
	PublicFilename string `json:"publicFilename,omitempty"`
}

//LoginResponse response for login
type LoginResponse struct {
	Token     string `json:"token"`
	Namespace string `json:"ns"`
}

//CountResponse response containing a count of changed items
type CountResponse struct {
	Count uint32 `json:"count"`
}
