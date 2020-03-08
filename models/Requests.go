package models

//PingRequest ping request
type PingRequest struct {
	Payload string
}

// UploadRequest contains file info (and a file)
type UploadRequest struct {
	Data       []byte         `json:"data"`
	Sum        string         `json:"sum"`
	Name       string         `json:"name"`
	Attributes FileAttributes `json:"attributes"`
}

//CredentialsRequest request containing credentials
type CredentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"pass"`
}

// FileUpdateRequest contains data to update a file
type FileUpdateRequest struct {
	FileID     int            `json:"fid"`
	Name       string         `json:"name,omitempty"`
	Attributes FileAttributes `json:"attributes"`
}

// FileRequest contains file info (and a file)
type FileRequest struct {
	FileID         uint                     `json:"fid"`
	Name           string                   `json:"name"`
	OptionalParams OptionalRequetsParameter `json:"opt"`
	Attributes     FileAttributes           `json:"attributes"`
}

//OptionalRequetsParameter optional request parameter
type OptionalRequetsParameter struct {
	Verbose uint8 `json:"verb"`
}
