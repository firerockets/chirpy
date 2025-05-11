package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	apiKey := headers.Get("Authorization")

	if apiKey == "" {
		return "", fmt.Errorf("no value found for key Authorization")
	}

	splitted := strings.Split(apiKey, " ")

	if len(splitted) != 2 || splitted[0] != "ApiKey" {
		return "", fmt.Errorf("ApiKey token format is wrong")
	}

	return splitted[1], nil
}
