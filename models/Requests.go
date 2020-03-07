package models

//CredentialRequest request containing credentials
type CredentialRequest struct {
	Username string `json:"username"`
	Password string `json:"pass"`
}

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
	Name       string         `json:"name"`
	Attributes FileAttributes `json:"attributes"`
}

//RestFile a file for rest responses
type RestFile struct {
	ID         int64 `json:"id"`
	Name       string
	Attributes FileAttributes
}
