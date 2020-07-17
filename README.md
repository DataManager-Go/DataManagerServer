# DataManager
A data storing, managing and sharing solution for you or your business. Upload files to a server (this repository) using the [CLI Client](https://github.com/Yukaru-san/DataManager_Client) or the [GUI Client](https://github.com/DataManager-Go/DataManagerGUI). This stores it and allows sharing via a preview html page.

# Basic concept
- There are namespaces
  - Namespaces use following format: username + "_" + namespaceName
  - Each user has a default namespace called username+"_default"
- One file belongs to one namespace
- Groups and tags can be assigned to files, this makes it easier to find them
- Registration can be enabled/disabled to allow/prevent users from creating an account
- Roles can give certain access to users
- File encryption is client side only. The server only stores the used cipher and the encrypted file but the en/decryption happens only client side
- Files are 'private' by default. Using the `publish` command or upload with `--public` or `--public-name <name>` makes a file available via a URL
  
# Installation

### Docker
The dockerfiles are hosted on https://hub.docker.com/r/jojii/dmanager
Since the file store path is `/app/files/` you have to map /app/files with a path on the host.<br>For instance `-v $(pwd)/files/:/app/files/`

### Manual

Run: `make`
(GO 1.11+ is required)

# Configuration
Create an example config using `./main config create`<br>
By default the config file is stored in `./data/config.yml`<br><br>

### Configurations:
#### Server
`database` A postgres database<br>
`pathconfig.filestore` The local folder for files. If you want to store the files in a different folder, change this value`<br>
`roles` The default roles. You <b>must</b> change them <b>before</b> the first start of server. Changes later on will be ignored.<br>
`allowregistration` Allows registrations from users<br>

#### Webserver
`useragentsrawfile` Respond with the raw file instead of the preview file. Very nice if you want to download the file instead of the preview if you are using wget or curl<br>
`maxpreviewfilesize` Max filesize for the preivew<br>
`htmlfiles` Path for the webroot. By default `./html`<br>

# Run
Run the server using `./main server start`<br>
You can add `-l debug` to view debug logs
