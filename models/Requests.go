package models

//PingRequest ping request
type PingRequest struct {
	Payload string
}

// FileRequest contains file info (and a file)
type FileRequest struct {
	FileID     int            `json:"fid"`
	Name       string         `json:"name"`
	Attributes FileAttributes `json:"attributes"`
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
