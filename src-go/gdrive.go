//go:build none

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// GDriveConfig holds the configuration for Google Drive connection
type GDriveConfig struct {
	CredentialsFile string // Path to credentials.json file
	TokenFile       string // Path to token.json file
	Scopes          []string
}

// getClient retrieves a token, saves it, then returns the generated client
func getClient(config *GDriveConfig) (*http.Client, error) {
	b, err := os.ReadFile(config.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config2, err := google.ConfigFromJSON(b, config.Scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	tok, err := tokenFromFile(config.TokenFile)
	if err != nil {
		tok, err = getTokenFromWeb(config2)
		if err != nil {
			return nil, err
		}
		saveToken(config.TokenFile, tok)
	}
	return config2.Client(context.Background(), tok), nil
}

// getTokenFromWeb requests a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %v", err)
	}
	return tok, nil
}

// tokenFromFile retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

// getFolderIdByPath retrieves the folder ID for a given path
// Path format: "folder1/folder2/folder3"
// Returns the ID of the deepest folder in the path
func getFolderIdByPath(config *GDriveConfig, path string) (string, error) {
	ctx := context.Background()
	client, err := getClient(config)
	if err != nil {
		return "", fmt.Errorf("unable to get client: %v", err)
	}

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	// Start with root folder
	parentId := "root"

	// If path is empty or root, return the root folder ID
	if path == "" || path == "/" {
		return parentId, nil
	}

	// Split the path into folder names
	folders := strings.Split(strings.Trim(path, "/"), "/")

	// Navigate through each folder in the path
	for _, folderName := range folders {
		if folderName == "" {
			continue
		}

		// Search for the folder in the current parent
		query := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and '%s' in parents and trashed = false",
			folderName, parentId)

		fileList, err := srv.Files.List().Q(query).Fields("files(id, name)").Do()
		if err != nil {
			return "", fmt.Errorf("unable to search for folder '%s': %v", folderName, err)
		}

		if len(fileList.Files) == 0 {
			return "", fmt.Errorf("folder not found: %s", folderName)
		}

		// Update parent ID to the found folder
		parentId = fileList.Files[0].Id
	}

	return parentId, nil
}

// createFolderPath creates a folder path if it doesn't exist and returns the ID of the deepest folder
func createFolderPath(config *GDriveConfig, path string) (string, error) {
	ctx := context.Background()
	client, err := getClient(config)
	if err != nil {
		return "", fmt.Errorf("unable to get client: %v", err)
	}

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	// Start with root folder
	parentId := "root"

	// If path is empty or root, return the root folder ID
	if path == "" || path == "/" {
		return parentId, nil
	}

	// Split the path into folder names
	folders := strings.Split(strings.Trim(path, "/"), "/")

	// Navigate through each folder in the path
	for _, folderName := range folders {
		if folderName == "" {
			continue
		}

		// Search for the folder in the current parent
		query := fmt.Sprintf("name = '%s' and mimeType = 'application/vnd.google-apps.folder' and '%s' in parents and trashed = false",
			folderName, parentId)

		fileList, err := srv.Files.List().Q(query).Fields("files(id, name)").Do()
		if err != nil {
			return "", fmt.Errorf("unable to search for folder '%s': %v", folderName, err)
		}

		if len(fileList.Files) == 0 {
			// Folder doesn't exist, create it
			folderMetadata := &drive.File{
				Name:     folderName,
				MimeType: "application/vnd.google-apps.folder",
				Parents:  []string{parentId},
			}

			folder, err := srv.Files.Create(folderMetadata).Fields("id").Do()
			if err != nil {
				return "", fmt.Errorf("unable to create folder '%s': %v", folderName, err)
			}

			parentId = folder.Id
			fmt.Printf("Created folder '%s' with ID: %s\n", folderName, folder.Id)
		} else {
			// Folder exists, use its ID
			parentId = fileList.Files[0].Id
			fmt.Printf("Found folder '%s' with ID: %s\n", folderName, parentId)
		}
	}

	return parentId, nil
}

// uploadFile uploads a file to Google Drive
func uploadFile(config *GDriveConfig, filePath string, parentFolderId string) (*drive.File, error) {
	ctx := context.Background()
	client, err := getClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to get client: %v", err)
	}

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %v", err)
	}
	defer file.Close()

	// Get file metadata
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("unable to get file info: %v", err)
	}

	// Create the file metadata
	driveFile := &drive.File{
		Name: filepath.Base(filePath),
	}

	// Set parent folder if provided
	if parentFolderId != "" {
		driveFile.Parents = []string{parentFolderId}
	}

	// Upload the file
	res, err := srv.Files.Create(driveFile).Media(file).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to create file: %v", err)
	}

	fmt.Printf("File '%s' uploaded successfully with ID: %s\n", fileInfo.Name(), res.Id)
	return res, nil
}

// uploadFileToPath uploads a file to a specific path in Google Drive
// Creates the path if it doesn't exist
func uploadFileToPath(config *GDriveConfig, filePath string, drivePath string) (*drive.File, error) {
	// Get or create the folder path
	folderId, err := createFolderPath(config, drivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve drive path '%s': %v", drivePath, err)
	}

	// Upload the file to the resolved folder
	return uploadFile(config, filePath, folderId)
}

/* Example main function
func main() {
	// Initialize the GDriveConfig
	config := &GDriveConfig{
		CredentialsFile: "credentials.json", // Path to your credentials.json file
		TokenFile:       "token.json",       // Path where token will be stored
		Scopes:          []string{drive.DriveFileScope},
	}

	// Example 1: Upload a file to a specific folder ID
	filePath := "example.txt" // Path to the file you want to upload

	// Example 2: Get folder ID by path
	drivePath := "MyDocs/2023/Reports"
	folderId, err := getFolderIdByPath(config, drivePath)
	if err != nil {
		log.Printf("Failed to get folder ID: %v", err)
		// If folder doesn't exist, you might want to create it
		folderId, err = createFolderPath(config, drivePath)
		if err != nil {
			log.Fatalf("Failed to create folder path: %v", err)
		}
	}

	// Upload file to the found/created folder
	_, err = uploadFile(config, filePath, folderId)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	// Example 3: Upload directly to a path (creates path if needed)
	_, err = uploadFileToPath(config, "another-file.txt", "MyDocs/Projects/2023")
	if err != nil {
		log.Fatalf("Failed to upload file to path: %v", err)
	}
}*/
