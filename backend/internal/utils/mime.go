package utils

import (
	"errors"
	"mime"
	"net/http"
	"path/filepath"
)

// for checking if extension matches detected MIME type
func ValidateMIME(filename string, header []byte) error {
	// 1. detecting MIME type from content :
	detected := http.DetectContentType(header)

	// 2. extracting extension from filename :
	ext := filepath.Ext(filename)

	// 3. compare detected MIME with extension :
	if exts, _ := mime.ExtensionsByType(detected); len(exts) > 0 {
		for _, e := range exts {
			if e == ext {
				return nil 
			}
		}

		// 4. mismmatch case :
		return errors.New("file extension does not match detected MIME type (" + detected + ")")
	}

	return nil
}
