package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	url2 "net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/storage"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/sirupsen/logrus"
)

const maxFileSize = 1024 * 1024 * 100

func crawlColly(ctx context.Context, input *MetadataInput, output *MetadataOutput, logOut *logrus.Logger, gptscript *gptscript.GPTScript) error {
	visited := make(map[string]struct{})
	for url := range output.State.WebsiteCrawlingState.VisitedURLs {
		filePath, err := convertUrlToFilePath(url)
		if err != nil {
			logOut.Errorf("Failed to convert URL to file path: %v", err)
			continue
		}
		visited[filePath] = struct{}{}
	}
	for _, url := range input.WebsiteCrawlingConfig.URLs {
		if err := scrape(ctx, logOut, output, gptscript, visited, url, output.State.WebsiteCrawlingState.CurrentURL, input.Limit); err != nil {
			return fmt.Errorf("failed to scrape %s: %w", url, err)
		}
	}

	for p, file := range output.Files {
		if _, ok := visited[p]; !ok {
			logOut.Infof("removing file %s", file.FilePath)
			if err := gptscript.DeleteFileInWorkspace(ctx, file.FilePath); err != nil {
				return err
			}
			delete(output.Files, p)
		}
	}

	output.Status = ""
	output.State.WebsiteCrawlingState = WebsiteCrawlingState{}
	return writeMetadata(ctx, output, gptscript)
}

func scrape(ctx context.Context, logOut *logrus.Logger, output *MetadataOutput, gptscriptClient *gptscript.GPTScript, visited map[string]struct{}, url, urlToResume string, limit int) error {
	collector := colly.NewCollector()

	inMemoryStore := &storage.InMemoryStorage{}
	inMemoryStore.Init()

	for url := range visited {
		if url == urlToResume {
			continue
		}
		h := fnv.New64a()
		h.Write([]byte(url))
		urlHash := h.Sum64()
		inMemoryStore.Visited(urlHash)
	}

	collector.SetStorage(inMemoryStore)

	collector.OnHTML("body", func(e *colly.HTMLElement) {
		html, err := e.DOM.Html()
		if err != nil {
			logOut.Errorf("Failed to grab HTML: %v", err)
			return
		}
		filePath, err := convertUrlToFilePath(e.Request.URL.String())
		if err != nil {
			logOut.Errorf("Failed to convert URL to file path: %v", err)
			return
		}
		if _, ok := visited[filePath]; ok {
			return
		}

		logOut.Infof("scraping %s", e.Request.URL.String())
		fileNotExists := false
		var notFoundError *gptscript.NotFoundInWorkspaceError
		if _, err := gptscriptClient.ReadFileInWorkspace(ctx, filePath); errors.As(err, &notFoundError) {
			fileNotExists = true
		}

		etag := e.Response.Headers.Get("ETag")
		lastModified := e.Response.Headers.Get("Last-Modified")
		var updatedAt string
		if etag != "" {
			updatedAt = etag
		} else if lastModified != "" {
			updatedAt = lastModified
		} else {
			updatedAt = time.Now().Format(time.RFC3339)
		}

		defer func() {
			if err := writeMetadata(ctx, output, gptscriptClient); err != nil {
				logOut.Infof("Failed to write metadata: %v", err)
			}
		}()

		if updatedAt == output.Files[e.Request.URL.String()].UpdatedAt && !fileNotExists {
			output.Status = fmt.Sprintf("Skipping %s because it has not changed", e.Request.URL.String())
			logOut.Infof("skipping %s because it has not changed for etag/last-modified: %s/%s", e.Request.URL.String(), etag, lastModified)
			return
		}
		data := []byte(html)

		if len(data) > maxFileSize {
			logOut.Infof("skipping %s because it is larger than %d MB", e.Request.URL.String(), maxFileSize/(1024*1024))
			return
		}

		checksum, err := getChecksum(data)
		if err != nil {
			logOut.Errorf("Failed to get checksum for %s: %v", e.Request.URL.String(), err)
			return
		}
		if checksum == output.Files[e.Request.URL.String()].Checksum && !fileNotExists {
			output.Status = fmt.Sprintf("Skipping %s because it has not changed", e.Request.URL.String())
			logOut.Infof("skipping %s because it has not changed", e.Request.URL.String())
			return
		}

		if err := gptscriptClient.WriteFileInWorkspace(ctx, filePath, data); err != nil {
			logOut.Errorf("Failed to write file %s: %v", filePath, err)
			return
		}

		visited[filePath] = struct{}{}

		output.Files[filePath] = FileDetails{
			FilePath:    filePath,
			URL:         e.Request.URL.String(),
			UpdatedAt:   updatedAt,
			Checksum:    checksum,
			SizeInBytes: int64(len(data)),
		}

		output.State.WebsiteCrawlingState.CurrentURL = e.Request.URL.String()
		output.State.WebsiteCrawlingState.VisitedURLs[output.State.WebsiteCrawlingState.CurrentURL] = struct{}{}

		output.Status = fmt.Sprintf("Scraped %v", e.Request.URL.String())
	})

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if len(visited) >= limit {
			return
		}

		baseURL, err := url2.Parse(url)
		if err != nil {
			logOut.Infof("Invalid base URL: %v", err)
			return
		}
		linkURL, err := url2.Parse(link)
		if err != nil {
			logOut.Infof("Invalid link URL %s: %v", link, err)
			return
		}
		if _, ok := visited[linkURL.String()]; ok {
			return
		}
		if strings.ToLower(path.Ext(linkURL.Path)) == ".pdf" {
			if err := scrapePDF(ctx, logOut, output, visited, linkURL, baseURL, gptscriptClient); err != nil {
				logOut.Infof("Failed to scrape PDF %s: %v", linkURL.String(), err)
			}
		} else {
			// don't scrape if linkURL link to external host
			if linkURL.Host != "" && !isSameDomainOrSubdomain(linkURL.Host, baseURL.Host) {
				return
			}

			// if linkURL has absolute path, and it doesn't match baseURL, skip
			if strings.HasPrefix(linkURL.Path, "/") && !strings.HasPrefix(linkURL.Path, baseURL.Path) {
				return
			}

			// if it is relative path, join with current path and check again
			finalPath := filepath.Clean(filepath.Join(e.Request.URL.Path, linkURL.Path))

			if !strings.HasPrefix(finalPath, baseURL.Path) {
				return
			}

			if linkURL.Host == "" && !strings.HasPrefix(link, "#") {
				if !strings.HasSuffix(baseURL.Path, "/") {
					baseURL.Path += "/"
				}
				fullLink := baseURL.ResolveReference(linkURL).String()
				parsedLink, err := url2.Parse(fullLink)
				if err != nil {
					logOut.Infof("Invalid link URL %s: %v", link, err)
					return
				}
				// don't scrape duplicate pages for homepage, for example, https://www.acorn.io and https://www.acorn.io/
				if parsedLink.Path == "/" {
					parsedLink.Path = ""
				}
				linkURL = parsedLink
			}
			e.Request.Visit(linkURL.String())
		}
	})

	if urlToResume != "" {
		return collector.Visit(urlToResume)
	}
	return collector.Visit(url)
}

func convertUrlToFilePath(url string) (string, error) {
	parsedUrl, err := url2.Parse(url)
	if err != nil {
		return "", fmt.Errorf("invalid URL %s: %v", url, err)
	}

	hostname := parsedUrl.Hostname()
	urlPathWithQuery := parsedUrl.Path
	if parsedUrl.RawQuery != "" {
		urlPathWithQuery += "?" + url2.QueryEscape(parsedUrl.RawQuery)
	}

	var filePath string
	if urlPathWithQuery == "" {
		filePath = path.Join(hostname, "index.html")
	} else {
		trimmedPath := strings.Trim(urlPathWithQuery, "/")
		if trimmedPath == "" {
			filePath = path.Join(hostname, "index.html")
		} else {
			segments := strings.Split(trimmedPath, "/")
			fileName := segments[len(segments)-1] + ".html"
			filePath = path.Join(hostname, strings.Join(segments[:len(segments)-1], "/"), fileName)
		}
	}

	return filePath, nil
}

func isSameDomainOrSubdomain(linkHostname, baseHostname string) bool {
	if linkHostname == baseHostname {
		return true
	}

	parts := strings.Split(baseHostname, ".")

	// if baseHostname is x.y, linkHostname can be www*.x.y
	if len(parts) == 2 {
		linkParts := strings.Split(linkHostname, ".")
		if len(linkParts) == 3 && (linkParts[0] == "www" || (len(linkParts[0]) == 4 && strings.HasPrefix(linkParts[0], "www"))) {
			return strings.Join(linkParts[1:], ".") == baseHostname
		}
	}

	// if baseHostname is www*.x.y, linkHostname can be x.y
	if len(parts) == 3 {
		if parts[0] == "www" || (len(parts[0]) == 4 && strings.HasPrefix(parts[0], "www")) {
			if linkHostname == parts[1]+"."+parts[2] {
				return true
			}
		}
	}

	return false
}

func scrapePDF(ctx context.Context, logOut *logrus.Logger, output *MetadataOutput, visited map[string]struct{}, linkURL *url2.URL, baseURL *url2.URL, gptscript *gptscript.GPTScript) error {
	if linkURL.Host == "" {
		var err error
		fullLink := baseURL.ResolveReference(linkURL).String()
		linkURL, err = url2.Parse(fullLink)
		if err != nil {
			return fmt.Errorf("invalid link URL %s: %v", fullLink, err)
		}
	}
	filePath := path.Join(linkURL.Host, strings.TrimPrefix(linkURL.Path, "/"))
	if !isSameDomainOrSubdomain(linkURL.Host, baseURL.Host) {
		filePath = path.Join(baseURL.Host, filePath)
	}
	if _, ok := visited[filePath]; ok {
		return nil
	}

	logOut.Infof("downloading PDF %s", linkURL.String())
	resp, err := http.Get(linkURL.String())
	if err != nil {
		return fmt.Errorf("failed to download PDF %s: %v", linkURL.String(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download PDF %s: status code %d", linkURL.String(), resp.StatusCode)
	}

	if resp.Header.Get("Content-Type") != "application/pdf" {
		logOut.Infof("skipping %s because it is not a PDF (likely redirect on old link)", linkURL.String())
		return nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read PDF %s: %v", linkURL.String(), err)
	}

	if len(data) > maxFileSize {
		logOut.Infof("skipping %s because it is larger than 100 MB", linkURL.String())
		return nil
	}

	newChecksum, err := getChecksum(data)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %v", err)
	}

	if fileDetails, exists := output.Files[linkURL.String()]; exists {
		if fileDetails.Checksum == newChecksum {
			logOut.Infof("PDF %s has not been modified", linkURL.String())
			return nil
		}
	}

	if err := gptscript.WriteFileInWorkspace(ctx, filePath, data); err != nil {
		return fmt.Errorf("failed to write PDF %s: %v", linkURL.String(), err)
	}

	visited[filePath] = struct{}{}

	output.Status = fmt.Sprintf("Scraped %v", linkURL.String())
	output.Files[filePath] = FileDetails{
		FilePath:    filePath,
		URL:         linkURL.String(),
		UpdatedAt:   time.Now().String(),
		Checksum:    newChecksum,
		SizeInBytes: int64(len(data)),
	}

	if err := writeMetadata(ctx, output, gptscript); err != nil {
		return fmt.Errorf("failed to write metadata: %v", err)
	}
	return nil
}

func getChecksum(content []byte) (string, error) {
	hash := sha256.New()
	_, err := hash.Write(content)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
