package services

import (
	"backend/internal/db"
	"database/sql"
	"time"
)

// FileMeta represents file metadata plus uploader details :
type FileMeta struct {
    //file metadata :
    ID            int       `json:"id"`
    Filename      string    `json:"filename"`
    Filepath      string    `json:"filepath"`
    Uploader      string    `json:"uploader"`
    Size          int64     `json:"size"`
    UploadedAt    time.Time `json:"uploaded_at"`
    IsMaster      bool      `json:"is_master"`
    IsPublic      bool      `json:"is_public"`
    DownloadCount int       `json:"download_count"`

    // uploader info :
    UploaderID       int       `json:"uploader_id"`
    UploaderUsername string    `json:"uploader_username"`
    UploaderEmail    string    `json:"uploader_email"`
    UploaderRole     string    `json:"uploader_role"`
    UploaderCreated  time.Time `json:"uploader_created_at"`
}



// GetFileByID returns file metadata (and uploader username) or (nil, nil) if not found.
func GetFileByID(fileID int) (*FileMeta, error) {
    var f FileMeta

    // query files joined with users to get uploader info :
    err := db.DB.QueryRow(`
        SELECT f.id, f.filename, f.filepath, f.size, f.uploaded_at, 
               f.is_master, f.is_public, f.download_count,
               u.id, u.username, u.email, u.role, u.created_at
        FROM files f
        JOIN users u ON f.user_id = u.id
        WHERE f.id = $1
    `, fileID).Scan(
        &f.ID,
        &f.Filename,
        &f.Filepath,
        &f.Size,
        &f.UploadedAt,
        &f.IsMaster,
        &f.IsPublic,
        &f.DownloadCount,
        &f.UploaderID,
        &f.UploaderUsername,
        &f.UploaderEmail,
        &f.UploaderRole,
        &f.UploaderCreated,
    )

    // returning nil if no rows found :
    if err == sql.ErrNoRows {
        return nil, nil
    }

    // return error if query failed : 
    if err != nil {
        return nil, err
    }

    // return populated structure :
    return &f, nil
}
