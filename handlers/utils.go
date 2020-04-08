package handlers

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/DataManager-Go/DataManagerServer/models"

	"github.com/JojiiOfficial/gaw"
	log "github.com/sirupsen/logrus"
)

func readRequestLimited(w http.ResponseWriter, r *http.Request, p interface{}, limit int64) bool {
	return readRequestBody(w, io.LimitReader(r.Body, limit), p)
}

//parseUserInput tries to read the body and parse it into p. Returns true on success
func readRequestBody(w http.ResponseWriter, r io.Reader, p interface{}) bool {
	body, err := ioutil.ReadAll(r)

	if LogError(err) {
		return false
	}

	return !handleAndSendError(json.Unmarshal(body, p), w, models.WrongInputFormatError, http.StatusUnprocessableEntity)
}

//LogError returns true on error
func LogError(err error, context ...map[string]interface{}) bool {
	if err == nil {
		return false
	}

	if len(context) > 0 {
		log.WithFields(context[0]).Error(err.Error())
	} else {
		log.Error(err.Error())
	}
	return true
}

//AllowedSchemes schemes that are allowed in urls
var AllowedSchemes = []string{"http", "https"}

func isValidHTTPURL(inp string) bool {
	//check for valid URL
	u, err := url.Parse(inp)
	if err != nil {
		return false
	}

	return gaw.IsInStringArray(u.Scheme, AllowedSchemes)
}

func isStructInvalid(x interface{}) bool {
	s := reflect.TypeOf(x)
	for i := s.NumField() - 1; i >= 0; i-- {
		e := reflect.ValueOf(x).Field(i)

		if hasEmptyValue(e) {
			return true
		}
	}
	return false
}

func hasEmptyValue(e reflect.Value) bool {
	switch e.Type().Kind() {
	case reflect.String:
		if e.String() == "" || strings.Trim(e.String(), " ") == "" {
			return true
		}
	case reflect.Array:
		for j := e.Len() - 1; j >= 0; j-- {
			isEmpty := hasEmptyValue(e.Index(j))
			if isEmpty {
				return true
			}
		}
	case reflect.Slice:
		return isStructInvalid(e)

	case
		reflect.Uintptr, reflect.Ptr, reflect.UnsafePointer,
		reflect.Uint64, reflect.Uint, reflect.Uint8, reflect.Bool,
		reflect.Struct, reflect.Int64, reflect.Int:
		{
			return false
		}
	default:
		log.Error(e.Type().Kind(), e)
		return true
	}
	return false
}

func downloadHTTP(user *models.User, url string, f *os.File, file *models.File) (int, error) {
	if !isValidHTTPURL(url) {
		return 0, errors.New("invalid url")
	}

	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}

	res, err := httpClient.Get(url)
	if LogError(err) {
		return 0, err
	}

	//Don't read content on http error
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return res.StatusCode, nil
	}

	//Check if file is too large
	if user.HasUploadLimit() && res.ContentLength > user.Role.MaxURLcontentSize {
		return res.StatusCode, errors.New("File too large")
	}

	//read response
	var reader io.Reader
	if user.HasUploadLimit() {
		//Use limited reader if user has limited download content size
		reader = io.LimitReader(res.Body, user.Role.MaxURLcontentSize)
	} else {
		//use body as reader to read everything
		reader = res.Body
	}

	//Save body in file
	size, err := io.Copy(f, reader)
	if LogError(err) {
		return 0, err
	}

	if err = res.Body.Close(); LogError(err) {
		return 0, err
	}

	//Set file size
	file.FileSize = size
	return res.StatusCode, nil
}

const bufferSize = 10 * 1024

// Just a little magic, nothing to see here
func readMultipartToFile(f *os.File, reader io.Reader, w http.ResponseWriter) (int64, bool, bool) {
	// Create multipart reader
	partReader := multipart.NewReader(reader, boundary)
	part, err := partReader.NextPart()

	// EOF is in this case 'no file found'
	if err == io.EOF {
		sendResponse(w, models.ResponseError, "No file provided", nil, http.StatusUnprocessableEntity)
		return 0, false, true
	} else if LogError(err) {
		return 0, false, true
	}

	// Buf the buffer
	buffer := make([]byte, bufferSize)
	sum, currTemp := make([]byte, 16), make([]byte, 16)
	var n, currTempCount int
	hash := md5.New()
	var exit, success bool
	var size int64
	// Multiwriter, to write hash and file at once
	hw := io.MultiWriter(hash, f)

	for {
		br := false
		n, err = part.Read(buffer)

		if n > 0 {
			if err == io.EOF {
				br = true
			} else if err != nil {
				exit = true
				break
			}

			if n >= 16 {
				sum = buffer[n-16 : n]

				if currTempCount > 0 {
					hw.Write(currTemp[:currTempCount])
					currTempCount = 0
				}

				hw.Write(buffer[:n-16])

				if n-16 >= 16 {
					currTempCount = n - (n - 16)
					copy(currTemp, buffer[n-16:n])
				}
			} else {
				if currTempCount+n > 16 {
					rb := ((n + currTempCount) - 16)
					hw.Write(currTemp[:rb])

					// like -> currTemp = append(currTemp[rb:], buffer[:n]...)
					// but better
					copy(currTemp, currTemp[rb:])
					app(16-rb, currTemp, buffer[:n])

					currTempCount = 16
				} else {
					add := 16 - currTempCount
					if add > n {
						add = n
					}

					copy(currTemp[currTempCount:currTempCount+add], buffer[:add])
					currTempCount += add
				}

				// like sum = append(sum[n:16], buffer[:n]...)
				// but better
				c := copy(sum[:16-n], sum[n:16])
				copy(sum[c:], buffer[:n])
			}

			size += int64(n)

			if LogError(err) {
				exit = true
				break
			}
		}

		if br || n == 0 {
			break
		}
	}

	// If only 16 bytes were read, return
	if size < 16 {
		return 0, false, true
	}

	// File do match if the generated and the passed
	// hashes match
	success = bytes.Equal(hash.Sum(nil), sum)

	// Substract 16 bytes for the hashsum
	size = size - 16
	return size, success, exit
}

// app appends src to dest using offset 'start'
func app(start int, dest, src []byte) {
	for i := start; i-start < len(src); i++ {
		dest[i] = src[i-start]
	}
}
