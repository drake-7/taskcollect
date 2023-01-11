package daymap

import (
	//"errors"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"main/errors"
	"main/logger"
)

type User struct {
	Timezone *time.Location
	Token    string
}

// Get a webpage from DayMap using a username and password.
// Used for authenticating with DayMap.
func get(webUrl, username, password string) (string, string, error) {
	// Stage 1 - Get a DayMap redirect to SAML.

	// A persistent cookie jar is required for the entire process.

	jar, err := cookiejar.New(nil)
	if err != nil {
		newErr := errors.NewError("daymap.get", "failed to create cookiejar", err)
		return "", "", newErr
	}

	client := &http.Client{Jar: jar}

	s1, err := client.Get(webUrl)
	if err != nil {
		newErr := errors.NewError("daymap.get", "GET request failed", err)
		return "", "", newErr
	}

	s1body, err := io.ReadAll(s1.Body)
	if err != nil {
		newErr := errors.NewError("daymap.get", "failed to read s1.Body", err)
		return "", "", newErr
	}

	s1page := string(s1body)

	// Stage 2 - POST credentials to SAML.

	// Generate POST form data with provided credentials.

	s2form := url.Values{}
	s2form.Set("UserName", username)
	s2form.Set("Password", password)
	s2form.Set("AuthMethod", "FormsAuthentication")
	s2data := strings.NewReader(s2form.Encode())

	// Get SAML request ID. This must be extracted to make a valid login.

	idIndex := strings.Index(s1page, "&client-request-id=")
	if idIndex == -1 {
		err = errNoClientRequestID
		return "", "", err
	}

	idEnd := strings.Index(s1page[idIndex:], `"`)
	idEnd += idIndex

	if idEnd == -1 {
		err = errUnterminatedClientRequestID
		return "", "", err
	}

	s2id := s1page[idIndex:idEnd]
	s2url := s1.Request.URL.String() + s2id

	// Send the POST request with the generated form data.

	s2req, err := http.NewRequest("POST", s2url, s2data)
	if err != nil {
		newErr := errors.NewError("daymap.get", "(s2) POST request failed", err)
		return "", "", newErr
	}

	s2req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s2, err := client.Do(s2req)
	if err != nil {
		newErr := errors.NewError("daymap.get", "s2req", err)
		return "", "", newErr
	}

	s2body, err := io.ReadAll(s2.Body)
	if err != nil {
		newErr := errors.NewError("daymap.get", "failed to read s2.Body", err)
		return "", "", newErr
	}

	s2page := string(s2body)

	// Stage 3 - Parse gigantic SAML form and redirect to OpenIdConnect.

	// Parse the SAML HTML form.

	s3form := url.Values{}
	s3index := strings.Index(s2page, "action=")

	if s3index == -1 {
		err := errNoActionAttrib
		return "", "", err
	}

	s3search := s2page[s3index:]

	for {
		s3index = strings.Index(s3search, `name="`)

		if s3index == -1 {
			break
		}

		s3search = s3search[s3index+6:]
		s3index = strings.Index(s3search, `"`)

		if s3index == -1 {
			err := errInvalidHtmlForm
			return "", "", err
		}

		key := s3search[:s3index]
		s3index = strings.Index(s3search, `value="`)

		if s3index == -1 {
			err := errAuthFailed
			return "", "", err
		}

		s3search = s3search[s3index+7:]
		s3index = strings.Index(s3search, `"`)

		if s3index == -1 {
			err := errInvalidHtmlForm
			return "", "", err
		}

		value := s3search[:s3index]
		s3form.Add(key, value)
	}

	s3wr := s3form.Get("wresult")
	s3wr = html.UnescapeString(s3wr)
	s3form.Set("wresult", s3wr)
	s3data := strings.NewReader(s3form.Encode())

	// Send the POST request with the payload.

	s3url := "https://portal.daymap.net/daymapidentity/adfs/gihs/"
	s3req, err := http.NewRequest("POST", s3url, s3data)
	if err != nil {
		newErr := errors.NewError("daymap.get", "(s3) POST request failed", err)
		return "", "", newErr
	}

	s3req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s3, err := client.Do(s3req)
	if err != nil {
		newErr := errors.NewError("daymap.get", "s3req", err)
		return "", "", newErr
	}

	s3body, err := io.ReadAll(s3.Body)
	if err != nil {
		newErr := errors.NewError("daymap.get", "failed to read s3.Body", err)
		return "", "", newErr
	}

	s3page := string(s3body)

	// Stage 4 - Parse OpenIdConnect form and POST to "/DayMap".

	// Parse the OpenIdConnect HTML form.

	s4form := url.Values{}
	s4index := strings.Index(s3page, "action=")

	if s4index == -1 {
		err := errNoActionAttrib
		return "", "", err
	}

	s4search := s3page[s4index:]

	for {
		s4index = strings.Index(s4search, `name='`)

		if s4index == -1 {
			break
		}

		s4search = s4search[s4index+6:]
		s4index = strings.Index(s4search, `'`)

		if s4index == -1 {
			err := errInvalidHtmlForm
			return "", "", err
		}

		key := s4search[:s4index]
		s4index = strings.Index(s4search, `value='`)

		if s4index == -1 {
			err := errInvalidHtmlForm
			return "", "", err
		}

		s4search = s4search[s4index+7:]
		s4index = strings.Index(s4search, `'`)

		if s4index == -1 {
			err := errInvalidHtmlForm
			return "", "", err
		}

		value := s4search[:s4index]
		s4form.Set(key, value)
	}

	s4data := strings.NewReader(s4form.Encode())

	// Send the POST request with the payload.

	s4url := "https://gihs.daymap.net/Daymap/"
	s4req, err := http.NewRequest("POST", s4url, s4data)
	if err != nil {
		newErr := errors.NewError("daymap.get", "(s4) POST request failed", err)
		return "", "", newErr
	}

	s4req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	s4, err := client.Do(s4req)
	if err != nil {
		newErr := errors.NewError("daymap.get", "s4req", err)
		return "", "", newErr
	}

	s4body, err := io.ReadAll(s4.Body)
	if err != nil {
		newErr := errors.NewError("daymap.get", "failed to read s4.Body", err)
		return "", "", newErr
	}

	s4page := string(s4body)

	daymapUrl := url.URL{
		Scheme: "https",
		Host:   "gihs.daymap.net",
	}

	cookies := jar.Cookies(&daymapUrl)
	authToken := ""

	for i := 0; i < len(cookies); i++ {
		authToken += cookies[i].String()

		if i < len(cookies)-1 {
			authToken += "; "
		}
	}

	return s4page, authToken, nil
}

// Authenticate to DayMap and retrieve a session token (an HTTP cookie).
func Auth(school, usr, pwd string) (User, error) {
	timezone, err := time.LoadLocation("Australia/Adelaide")
	if err != nil {
		newErr := errors.NewError("daymap.Auth", "failed to load timezone", err)
		logger.Error(newErr)
	}

	// TODO: Implement DayMap auth for other schools as well
	page := "https://gihs.daymap.net/daymap/student/dayplan.aspx"
	_, authToken, err := get(page, usr, pwd)
	if err != nil {
		newErr := errors.NewError("daymap.Auth", "could not authenticate user with DayMap", err)
		return User{}, newErr
	}

	creds := User{
		Timezone: timezone,
		Token:    authToken,
	}

	return creds, nil
}
