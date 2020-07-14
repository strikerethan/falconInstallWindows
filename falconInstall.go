package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

type TokenAPIResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type SensorInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Platform    string `json:"platform"`
	Os          string `json:"os"`
	OsVersion   string `json:"os_version"`
	Sha256      string `json:"sha256"`
	ReleaseDate string `json:"release_date"`
	Version     string `json:"version"`
	FileSize    int    `json:"file_size"`
	FileType    string `json:"file_type"`
}

type RequestMetaInfo struct {
	QueryTime float64 `json:"query_time"`
	PoweredBy string  `json:"powered_by"`
	TraceID   string  `json:"trade_id"`
}

type RequestErrors struct {
	Errors map[string]string
}

type SensorAPIResponse struct {
	Meta      RequestMetaInfo `json:"meta"`
	Errors    []RequestErrors `json:"errors"`
	Resources []SensorInfo    `json:"resources"`
}

type CCIDAPIResponse struct {
	Meta      RequestMetaInfo `json:"meta"`
	Resources []string        `json:"resources"`
	Errors    []RequestErrors `json:"errors"`
}

func main() {

	// Authenticate and get token
	token := getToken()
	fmt.Println(token) // Just to show that we have our token to use

	// Get list of sensor versions, get latest -1
	sensor := getSensor(token)
	fmt.Println(sensor)

	// Create download link for specific sensor
	link := downloadLink(sensor)
	fmt.Println(link)

	//Download the Sensor
	var filename string = "falcon.exe"
	downloadSensor(link, filename, token)

	// Get CCID for Configuration
	ccid := getCCID(token)
	fmt.Println(ccid)

	// Install Sensor
	installSensor(filename, ccid)

	// Verify installation
	// TODO

}

// This uses command flags for clientId and clientSecret to query the API to get and return a token
func getToken() string {

	// This gets the clientId and secret from the command line flags
	clientIDPtr := flag.String("clientId", "", "This is the client Id")
	clientSecretPtr := flag.String("clientSecret", "", "This is the client secret")
	flag.Parse()

	body := strings.NewReader("client_id=" + *clientIDPtr + "&client_secret=" + *clientSecretPtr)
	req, err := http.NewRequest("POST", "https://api.crowdstrike.com/oauth2/token", body)
	if err != nil {
		// handle err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	// The defer keyword will run what is after it when the function completes
	defer resp.Body.Close()

	rbody, err := ioutil.ReadAll(resp.Body)

	// Here we send the rbody var to our function to unmarshall the data based on the struct we created above for TokenAPIResponse
	s, err := getTokenResponseData([]byte(rbody))

	return s.AccessToken
}

func getTokenResponseData(body []byte) (*TokenAPIResponse, error) {
	var s = new(TokenAPIResponse)
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("whoops:", err)
	}
	return s, err
}

func getSensor(token string) string {

	req, err := http.NewRequest("GET", "https://api.crowdstrike.com/sensors/combined/installers/v1?filter=platform:%27windows%27", nil)
	if err != nil {
		// handle err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()

	rbody, err := ioutil.ReadAll(resp.Body)

	s, err := getSensorVersionData([]byte(rbody))

	return s.Resources[1].Sha256

}

func getSensorVersionData(body []byte) (*SensorAPIResponse, error) {

	var s = new(SensorAPIResponse)
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("whoops:", err)
	}
	return s, err
}

func downloadLink(sensor string) string {

	str := "https://api.crowdstrike.com/sensors/entities/download-installer/v1?"

	u, _ := url.Parse(str)

	q, _ := url.ParseQuery(u.RawQuery)

	fmt.Println()

	q.Add("id", sensor)

	u.RawQuery = q.Encode()
	return u.String()
}

func downloadSensor(link string, filepath string, token string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		// handle err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
	

func getCCID(token string) string {

	req, err := http.NewRequest("GET", "https://api.crowdstrike.com/sensors/queries/installers/ccid/v1", nil)
	if err != nil {
		// handle err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle err
	}
	defer resp.Body.Close()

	rbody, err := ioutil.ReadAll(resp.Body)

	s, err := getCCIDdata([]byte(rbody))

	return s.Resources[0]

}

func getCCIDdata(body []byte) (*CCIDAPIResponse, error) {

	var s = new(CCIDAPIResponse)
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("whoops:", err)
	}
	return s, err
}

func installSensor(filename string, ccid string) string {

        out, err := exec.Command(filename, "/install", "/quiet", "/norestart", "CID="+ccid).Output()
        if err != nil {
                fmt.Printf("%s", err)
        }

        fmt.Println("Command Successfully Executed")
        output := string(out[:])
        fmt.Println(output)
        return output

}
