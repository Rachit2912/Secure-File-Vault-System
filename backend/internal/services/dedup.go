package services

import (
	"backend/internal/models"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"
)

// Compute SHA-256 hash of uploaded file
func ComputeHash(file multipart.File) (string, error) {
    hash := sha256.New()
    if _, err := io.Copy(hash, file); err != nil {
        return "", err
    }
    return hex.EncodeToString(hash.Sum(nil)), nil
}

// Find duplicate by comparing hash with candidates
func FindDuplicate(file multipart.File, candidates []*models.File) (*models.File, string, error) {
    //computing hash of file : 
    hash, err := ComputeHash(file)
    if err != nil {
        return nil, "", err
    }

    // matching hash with candidates :
    for _, c := range candidates {
        if c.Hash == hash {
            return c, hash, nil
        }
    }
    return nil, hash, nil
}
