package middleware

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ashrafAli23/nestgo/core"
)

// UploadConfig holds file upload validation configuration.
type UploadConfig struct {
	// FieldName is the form field name for the file. Default: "file".
	FieldName string
	// MaxFileSize is the maximum file size in bytes. Default: 10MB.
	MaxFileSize int64
	// AllowedMIMETypes restricts uploads to specific MIME types.
	// Empty means all types are allowed.
	// Example: []string{"image/png", "image/jpeg", "application/pdf"}
	AllowedMIMETypes []string
	// AllowedExtensions restricts uploads to specific file extensions (with dot).
	// Empty means all extensions are allowed.
	// Example: []string{".png", ".jpg", ".pdf"}
	AllowedExtensions []string
}

// DefaultUploadConfig returns sensible upload defaults.
func DefaultUploadConfig() UploadConfig {
	return UploadConfig{
		FieldName:   "file",
		MaxFileSize: 10 * 1024 * 1024, // 10MB
	}
}

// UploadedFile holds the validated file header after the Upload middleware runs.
// Retrieve it from context via c.Get("uploaded_file").
type UploadedFile struct {
	Header   *multipart.FileHeader
	Filename string
	Size     int64
	MIMEType string
}

// Upload returns a middleware that validates a file upload before the handler runs.
// On success, stores an *UploadedFile in the context under "uploaded_file".
//
// Usage:
//
//	r.POST("/avatar", handler, middleware.Upload(middleware.UploadConfig{
//	    MaxFileSize:      5 * 1024 * 1024,
//	    AllowedMIMETypes: []string{"image/png", "image/jpeg"},
//	}))
//
//	// In handler:
//	func (ctrl *Ctrl) UploadAvatar(c core.Context) error {
//	    uf := c.Get("uploaded_file").(*middleware.UploadedFile)
//	    // uf.Header, uf.Filename, uf.Size, uf.MIMEType are available
//	}
func Upload(config ...UploadConfig) core.MiddlewareFunc {
	cfg := DefaultUploadConfig()
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.FieldName == "" {
		cfg.FieldName = "file"
	}
	if cfg.MaxFileSize <= 0 {
		cfg.MaxFileSize = 10 * 1024 * 1024
	}

	// Pre-build lookup maps for O(1) checks.
	mimeMap := make(map[string]struct{}, len(cfg.AllowedMIMETypes))
	for _, m := range cfg.AllowedMIMETypes {
		mimeMap[strings.ToLower(m)] = struct{}{}
	}
	extMap := make(map[string]struct{}, len(cfg.AllowedExtensions))
	for _, e := range cfg.AllowedExtensions {
		extMap[strings.ToLower(e)] = struct{}{}
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(c core.Context) error {
			fh, err := c.FormFile(cfg.FieldName)
			if err != nil {
				return core.ErrBadRequest("file upload required: " + cfg.FieldName)
			}

			// Check file size
			if fh.Size > cfg.MaxFileSize {
				return core.NewHTTPError(http.StatusRequestEntityTooLarge, "file too large")
			}

			// Check extension
			if len(extMap) > 0 {
				ext := strings.ToLower(filepath.Ext(fh.Filename))
				if _, ok := extMap[ext]; !ok {
					return core.ErrBadRequest("file extension not allowed: " + ext)
				}
			}

			// Detect MIME type from file content (not the header, which can be spoofed).
			mimeType := ""
			if len(mimeMap) > 0 {
				f, err := fh.Open()
				if err != nil {
					return core.ErrInternalServer("failed to open uploaded file")
				}
				// Read first 512 bytes for MIME detection
				buf := make([]byte, 512)
				n, readErr := f.Read(buf)
				f.Close()
				if readErr != nil && readErr != io.EOF {
					return core.ErrInternalServer("failed to read uploaded file for MIME detection")
				}
				mimeType = http.DetectContentType(buf[:n])
				if _, ok := mimeMap[mimeType]; !ok {
					return core.ErrBadRequest("file type not allowed: " + mimeType)
				}
			}

			c.Set("uploaded_file", &UploadedFile{
				Header:   fh,
				Filename: fh.Filename,
				Size:     fh.Size,
				MIMEType: mimeType,
			})

			return next(c)
		}
	}
}

// SaveFile is a helper to save an uploaded file to disk.
//
//	uf := c.Get("uploaded_file").(*middleware.UploadedFile)
//	err := middleware.SaveFile(uf.Header, "/uploads/avatar.png")
func SaveFile(fh *multipart.FileHeader, dst string) error {
	src, err := fh.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Ensure destination directory exists
	if dir := filepath.Dir(dst); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
