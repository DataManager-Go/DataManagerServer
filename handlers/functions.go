package handlers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/JojiiOfficial/DataManagerServer/models"

	gaw "github.com/JojiiOfficial/GoAw"
	log "github.com/sirupsen/logrus"
)

func sendResponse(w http.ResponseWriter, status models.ResponseStatus, message string, payload interface{}, params ...int) {
	statusCode := http.StatusOK
	s := "0"
	if status == 1 {
		s = "1"
	}

	w.Header().Set(models.HeaderStatus, s)
	w.Header().Set(models.HeaderStatusMessage, message)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	if len(params) > 0 {
		statusCode = params[0]
		w.WriteHeader(statusCode)
	}

	var err error
	if payload != nil {
		err = json.NewEncoder(w).Encode(payload)
	} else if len(message) > 0 {
		_, err = fmt.Fprintln(w, message)
	}

	LogError(err)
}

//parseUserInput tries to read the body and parse it into p. Returns true on success
func parseUserInput(config *models.Config, w http.ResponseWriter, r *http.Request, p interface{}) bool {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, config.Webserver.MaxBodyLength))

	if LogError(err) || LogError(r.Body.Close()) {
		return false
	}

	return !handleAndSendError(json.Unmarshal(body, p), w, models.WrongInputFormatError, http.StatusUnprocessableEntity)
}

func handleAndSendError(err error, w http.ResponseWriter, message string, statusCode int) bool {
	if !LogError(err) {
		return false
	}
	sendResponse(w, models.ResponseError, message, nil, statusCode)
	return true
}

func sendServerError(w http.ResponseWriter) {
	sendResponse(w, models.ResponseError, "internal server error", nil, http.StatusInternalServerError)
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

//GetMD5Hash return hash of input
func GetMD5Hash(text []byte) string {
	hash := md5.Sum(text)
	return hex.EncodeToString(hash[:])
}

func doHTTPGetRequest(config *models.Config, url string) (int, []byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return 0, []byte{}, err
	}

	//Don't read content on http error
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return res.StatusCode, []byte{}, nil
	}

	//Check if file is too large
	if res.ContentLength > config.Server.MaxHTTPDownloadSize {
		return res.StatusCode, []byte{}, errors.New("File too large")
	}

	//read response
	body, err := ioutil.ReadAll(io.LimitReader(res.Body, config.Server.MaxHTTPDownloadSize))
	if LogError(err) || LogError(res.Body.Close()) {
		return 0, []byte{}, err
	}

	return res.StatusCode, body, nil
}

//Returns the size in bytes of the header
func getHeaderSize(headers http.Header) uint32 {
	var size uint32
	for k, v := range headers {
		size += uint32(len(k))
		for _, val := range v {
			size += uint32(len(val))
		}
	}
	return size
}
