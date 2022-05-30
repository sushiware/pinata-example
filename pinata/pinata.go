package pinata

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const pinataURL = "https://api.pinata.cloud/"

var (
	PinataAPIKey    = os.Getenv("PINATA_API_KEY")
	PinataAPISecret = os.Getenv("PINATA_API_SECRET")
)

type PinataClient struct {
	httpClient   *http.Client
	apiKey       string
	secretAPIKey string
}

func New(apiKey string, secretAPIKey string) *PinataClient {
	return &PinataClient{
		httpClient:   http.DefaultClient,
		apiKey:       apiKey,
		secretAPIKey: secretAPIKey,
	}
}

type HealthcheckResult struct {
	Message string `json:"message"`
}

func (s *PinataClient) Healthcheck() error {
	const path = "data/testAuthentication"

	req, err := http.NewRequest(
		http.MethodGet,
		pinataURL+path,
		nil,
	)

	if err != nil {
		return err
	}

	req.Header.Set("PINATA_API_KEY", s.apiKey)
	req.Header.Set("PINATA_SECRET_API_KEY", s.secretAPIKey)

	res, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}

	result := &HealthcheckResult{}

	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		return err
	}

	return nil
}

type PinataMetadata struct {
	Name      string            `json:"name"`
	KeyValues map[string]string `json:"keyvalues"`
}

type PinResult struct {
	IPFSHash    string    `json:"IpfsHash"`
	PinSize     int       `json:"PinSize"`
	Timestamp   time.Time `json:"Timestamp"`
	IsDuplicate bool      `json:"isDuplicate"`
}

func (c *PinataClient) PinDir(contents [][]byte, names []string, dir string, metadata *PinataMetadata) (*PinResult, error) {
	const path = "pinning/pinFileToIPFS"

	body := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(body)

	for i, content := range contents {
		if err := c.writeFile(writer, filepath.Join(dir, names[i]), content); err != nil {
			return nil, err
		}
	}

	if err := c.writeMetadataPart(writer, metadata); err != nil {
		return nil, err
	}

	if err := c.writeOptionsPart(writer); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		pinataURL+path,
		body,
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("PINATA_API_KEY", c.apiKey)
	req.Header.Set("PINATA_SECRET_API_KEY", c.secretAPIKey)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	result := &PinResult{}

	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *PinataClient) PinJSON(content []byte) (*PinResult, error) {
	const path = "pinning/pinJSONToIPFS"

	body := bytes.NewBuffer(nil)

	type Body struct {
		PinataContent json.RawMessage `json:"pinataContent"`
	}

	if err := json.NewEncoder(body).Encode(&Body{PinataContent: content}); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		pinataURL+path,
		body,
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("PINATA_API_KEY", c.apiKey)
	req.Header.Set("PINATA_SECRET_API_KEY", c.secretAPIKey)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	result := &PinResult{}

	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *PinataClient) writeFile(writer *multipart.Writer, name string, file []byte) error {
	filePart, err := writer.CreateFormFile("file", name)
	if err != nil {
		return err
	}

	if _, err := filePart.Write(file); err != nil {
		return err
	}

	return nil
}

func (c *PinataClient) writeOptionsPart(writer *multipart.Writer) error {
	optionsPart, err := writer.CreateFormField("pinataOptions")
	if err != nil {
		return err
	}

	options := map[string]interface{}{
		"cidVersion":        1,
		"wrapWithDirectory": true,
	}

	bs, err := json.Marshal(options)
	if err != nil {
		return err
	}

	if _, err := optionsPart.Write(bs); err != nil {
		return err
	}

	return nil
}

func (c *PinataClient) writeMetadataPart(writer *multipart.Writer, metadata *PinataMetadata) error {
	if metadata == nil {
		return nil
	}

	metadataPart, err := writer.CreateFormField("pinataMetadata")
	if err != nil {
		return err
	}

	bs, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	if _, err := metadataPart.Write(bs); err != nil {
		return err
	}

	return nil
}
