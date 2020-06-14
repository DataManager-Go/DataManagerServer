# DataManager
A data storing and sharing solution for you or your business. Upload files to this server using the [CLI Client](https://github.com/Yukaru-san/DataManager_Client) or the [GUI Client](https://github.com/DataManager-Go/DataManagerGUI). This stores it and allows sharing via a preview html page.

# Basic concept
- There are namespaces
  - User namespaces start with the username + "_" + namespace
  - Each user has a default namespace called $username+"_default"
- A file belongs to 1 namespace
- Groups and tags can be assigned to files, this makes it easier to find files
- Registration can be enabled/disabled to allow/prevent users from creating an account
- Roles can give certain access to users
- File encryption is client side only. The server only stores the used cipher and the encrypted file but the en/decryption happens only in client side
- File are 'private' by default. Using the `publish` command or upload with `--public` makes a file available via the webpage
  
# Installation

### Docker
The dockerfiles are hosted on https://hub.docker.com/r/jojii/dmanager
Since the file store path is `/app/files/` you have to map /app/files with a path on the host.<br>For instance `-v $(pwd)/files/:/app/files/`
### Manual
```go
go mod download && go build -o main
```
To download the dependencies and build the application. Go 1.11+ is required.

# Configuration
Create an example config using `./main config create`<br>
By default the config file is stored in `./data/config.yml`<br><br>
### Configurations:
#### Server
`database` A postgres database<br>
`pathconfig.filestore` The store for files. Can be default but if you want to store the files in a different folder<br>
`roles` The default roles. You <b>must</b> change them <b>before</b> the first start of server. Changes later on will be ignored.<br>
`allowregistration` Allows registrations from users<br>

#### Webserver
`useragentsrawfile` Respond with the raw file instead of the preview file. Very nice if you want to download the file instead of the preview if you are using wget or curl<br>
`maxpreviewfilesize` Max filesize for the preivew<br>
`htmlfiles` Path for the webroot. By default `./html`<br>

# Run
Run the server using `./main server start`<br>
You can add `-l debug` to view debug logs
