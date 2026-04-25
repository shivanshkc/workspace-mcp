package workspace

import (
	"context"
	"fmt"
)

// DriveFile holds the basic metadata for a file returned by search or folder listing.
type DriveFile struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	MIMEType string `json:"mimeType"`
	Type     string `json:"type"` // human-friendly type derived from mimeType
}

// DriveFileMetadata holds full metadata for a single file.
type DriveFileMetadata struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	MIMEType     string `json:"mimeType"`
	Type         string `json:"type"`
	Size         int64  `json:"size"`
	ModifiedTime string `json:"modifiedTime"`
	OwnerEmail   string `json:"ownerEmail"`
}

// SearchFiles searches Google Drive using the provided query string and returns
// matching files. Results are limited to 50 and exclude trashed files.
func (c *Client) SearchFiles(ctx context.Context, query string) ([]DriveFile, error) {
	q := fmt.Sprintf("(%s) and trashed = false", query)

	resp, err := c.driveService.Files.List().
		Q(q).
		Fields("files(id,name,mimeType)").
		PageSize(50).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}

	files := make([]DriveFile, len(resp.Files))
	for i, f := range resp.Files {
		files[i] = DriveFile{
			ID:       f.Id,
			Name:     f.Name,
			MIMEType: f.MimeType,
			Type:     mimeTypeToFriendly(f.MimeType),
		}
	}
	return files, nil
}

// ListFolderContents returns the direct children of the given folder.
// Use "root" as folderID to list the top level of My Drive.
func (c *Client) ListFolderContents(ctx context.Context, folderID string) ([]DriveFile, error) {
	q := fmt.Sprintf("'%s' in parents and trashed = false", folderID)

	resp, err := c.driveService.Files.List().
		Q(q).
		Fields("files(id,name,mimeType)").
		PageSize(50).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list folder %q: %w", folderID, err)
	}

	files := make([]DriveFile, len(resp.Files))
	for i, f := range resp.Files {
		files[i] = DriveFile{
			ID:       f.Id,
			Name:     f.Name,
			MIMEType: f.MimeType,
			Type:     mimeTypeToFriendly(f.MimeType),
		}
	}
	return files, nil
}

// GetFileMetadata returns full metadata for a single file by ID.
func (c *Client) GetFileMetadata(ctx context.Context, fileID string) (DriveFileMetadata, error) {
	f, err := c.driveService.Files.Get(fileID).
		Fields("id,name,mimeType,size,modifiedTime,owners").
		Context(ctx).
		Do()
	if err != nil {
		return DriveFileMetadata{}, fmt.Errorf("failed to get metadata for file %s: %w", fileID, err)
	}

	ownerEmail := ""
	if len(f.Owners) > 0 {
		ownerEmail = f.Owners[0].EmailAddress
	}

	return DriveFileMetadata{
		ID:           f.Id,
		Name:         f.Name,
		MIMEType:     f.MimeType,
		Type:         mimeTypeToFriendly(f.MimeType),
		Size:         f.Size,
		ModifiedTime: f.ModifiedTime,
		OwnerEmail:   ownerEmail,
	}, nil
}

// mimeTypeToFriendly converts a MIME type to a short, readable type label.
func mimeTypeToFriendly(mimeType string) string {
	switch mimeType {
	case "application/vnd.google-apps.document":
		return "google_doc"
	case "application/vnd.google-apps.spreadsheet":
		return "google_sheet"
	case "application/vnd.google-apps.presentation":
		return "google_slide"
	case "application/vnd.google-apps.folder":
		return "folder"
	case "application/pdf":
		return "pdf"
	default:
		return mimeType
	}
}
