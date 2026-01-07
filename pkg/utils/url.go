package utils

import (
	"net/url"
	"strings"
)

func isLoopbackHost(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	return host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "0.0.0.0"
}

func normalizePublicBaseURL(baseURL string) string {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return ""
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	if parsed.Hostname() == "" || parsed.Scheme == "" {
		return ""
	}

	if isLoopbackHost(parsed.Hostname()) {
		return ""
	}

	return strings.TrimSuffix(baseURL, "/")
}

func GetSystemFileURL(path string) string {
	if path == "" {
		return path
	}

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	baseUrl := normalizePublicBaseURL(GetBaseUrl())
	if baseUrl == "" {
		return path
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return baseUrl + path
}

func GenerateFullURL(path string, storageType string) string {
	if storageType != "local" {
		return path
	}
	return GetSystemFileURL(path)
}

func GetFileFullURL(fileID string) string {
	baseUrl := normalizePublicBaseURL(GetBaseUrl())
	if baseUrl == "" {
		return "/f/" + fileID
	}
	return baseUrl + "/f/" + fileID
}

func GetFileThumbnailFullURL(fileID string) string {
	baseUrl := normalizePublicBaseURL(GetBaseUrl())
	if baseUrl == "" {
		return "/t/" + fileID
	}
	return baseUrl + "/t/" + fileID
}
