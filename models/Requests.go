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
	PublicName string         `json:"pubname,omitempty"`
	Updates    FileUpdateItem `json:"updates,omitempty"`
	All        bool           `json:"all"`
	Attributes FileAttributes `json:"attributes"`
}

//NamespaceRequest namespace action request
type NamespaceRequest struct {
	Namespace string        `json:"ns"`
	NewName   string        `json:"newName,omitempty"`
	Type      NamespaceType `json:"nstype"`
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

// UpdateAttributeRequest contains data to update a tag
type UpdateAttributeRequest struct {
	Name      string `json:"name"`
	NewName   string `json:"newname"`
	Namespace string `json:"namespace"`
}

// FileListRequest contains file info (and a file)
type FileListRequest struct {
	FileID         uint                     `json:"fid"`
	Name           string                   `json:"name"`
	AllNamespaces  bool                     `json:"allns"`
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
	URL        string         `json:"url"`
	Name       string         `json:"name"`
	Public     bool           `json:"public"`
	PublicName string         `json:"pbname"`
	FileType   string         `json:"ftype"`
	Attributes FileAttributes `json:"attributes"`
}

//UploadType type of upload
type UploadType uint8

//Available upload types
const (
	FileUploadType UploadType = iota
	URLUploadType
)
