-- removing file_description from files :
ALTER TABLE files
DROP COLUMN IF EXISTS description;
