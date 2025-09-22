-- adding file_description to files :
ALTER TABLE files
ADD COLUMN IF NOT EXISTS description TEXT;
