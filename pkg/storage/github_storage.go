package storage

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GithubStorage struct {
	Token string
	Owner string
	Repo  string
}

func (ss GithubStorage) request(method, u string, body io.Reader, headers map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+ss.Token)
	for headerKey, headerValue := range headers {
		req.Header.Add(headerKey, headerValue)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("request failed: %s", resp.Status)
	}

	return resp, nil
}

func (ss GithubStorage) requestJSON(method, u string, body io.Reader, v interface{}) error {
	resp, err := ss.request(method, u, body, map[string]string{
		"Accept": "application/vnd.github.v3+json",
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
		if err != nil {
			return fmt.Errorf("could not decode response: %w", err)
		}
	}

	return nil
}

func (ss GithubStorage) url(path string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", ss.Owner, ss.Repo, path)
}

type fileContentsResponse struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int64  `json:"size"`
	URL         string `json:"url"`
	HTMLURL     string `json:"html_url"`
	GitURL      string `json:"git_url"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
	Links       struct {
		Self string `json:"self"`
		Git  string `json:"git"`
		HTML string `json:"html"`
	} `json:"_links"`
}

func (ss GithubStorage) getFileContents(path string) (fileContentsResponse, error) {
	var response fileContentsResponse

	u := ss.url(path)
	err := ss.requestJSON(http.MethodGet, u, nil, &response)
	if err != nil {
		return response, err
	}

	return response, nil
}

func (ss GithubStorage) getSHA(path string) string {
	fileContents, err := ss.getFileContents(path)
	if err != nil {
		return ""
	}
	return fileContents.SHA
}

func (ss GithubStorage) Read(path string) ([]byte, error) {
	fileContents, err := ss.getFileContents(path)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(fileContents.DownloadURL)
	if err != nil {
		return nil, fmt.Errorf("could not download file: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (ss GithubStorage) Write(path string, b []byte, message string) error {
	u := ss.url(path)
	sha := ss.getSHA(path)
	content := base64.StdEncoding.EncodeToString(b)

	var request = struct {
		Message string `json:"message"`
		Content string `json:"content"`
		SHA     string `json:"sha"`
	}{
		Message: message,
		SHA:     sha,
		Content: content,
	}

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(request)
	if err != nil {
		return err
	}

	err = ss.requestJSON(http.MethodPut, u, buf, nil)
	if err != nil {
		return err
	}

	return nil
}

func (ss GithubStorage) Delete(path string, message string) error {
	u := ss.url(path)
	sha := ss.getSHA(path)

	var request = struct {
		Message string `json:"message"`
		SHA     string `json:"sha"`
	}{
		Message: message,
		SHA:     sha,
	}

	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(request)
	if err != nil {
		return err
	}

	return ss.requestJSON(http.MethodDelete, u, buf, nil)
}
