package gcloud

import (
	"io"
	"net/http"
)

func getMetadata(url string) ([]byte, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "http://metadata.google.internal/"+url, nil)
	if err != nil {
		return nil, nil
	}

	req.Header.Add("Metadata-Flavor", `Google`)
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil
	}

	return res, nil
}

func getInstanceId() (string, error) {
	res, err := getMetadata("/computeMetadata/v1/instance/id")
	if err != nil {
		return "", err
	}

	return string(res), nil
}

func getZone() (string, error) {
	res, err := getMetadata("/computeMetadata/v1/instance/zone")
	if err != nil {
		return "", err
	}

	return string(res), nil
}
