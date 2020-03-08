package models

//PingRequest ping request
type PingRequest struct {
	Payload string
}

//CredentialsRequest request containing credentials
type CredentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"pass"`
}

// FileRequest contains data to update a file
type FileRequest struct {
	FileID     uint           `json:"fid"`
	Name       string         `json:"name,omitempty"`
	Updates    FileUpdateItem `json:"updates,omitempty"`
	Attributes FileAttributes `json:"attributes"`
}

// FileUpdateItem lists changes to a file
type FileUpdateItem struct {
	IsPublic     string   `json:"ispublic,omitempty"`
	NewName      string   `json:"name,omitempty"`
	NewNamespace string   `json:"namespace,omitempty"`
	RemoveTags   []string `json:"rem_tags,omitempty"`
	RemoveGroups []string `json:"rem_groups,omitempty"`
	AddTags      []string `json:"add_tags,omitempty"`
	AddGroups    []string `json:"add_groups,omitempty"`
}

// FileListRequest contains file info (and a file)
type FileListRequest struct {
	FileID         uint                     `json:"fid"`
	Name           string                   `json:"name"`
	OptionalParams OptionalRequetsParameter `json:"opt"`
	Attributes     FileAttributes           `json:"attributes"`
}

//OptionalRequetsParameter optional request parameter
type OptionalRequetsParameter struct {
	Verbose uint8 `json:"verb"`
}

// UploadRequest contains file info (and a file)
type UploadRequest struct {
	UploadType UploadType     `json:"type"`
	Data       []byte         `json:"data"`
	URL        string         `json:"url"`
	Sum        string         `json:"sum"`
	Name       string         `json:"name"`
	Attributes FileAttributes `json:"attributes"`
}

//UploadType type of upload
type UploadType uint8

//Available upload types
const (
	FileUploadType UploadType = iota
	URLUploadType
)
