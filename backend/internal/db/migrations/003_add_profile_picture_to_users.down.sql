-- removing profile_picture col from users :
ALTER TABLE users
DROP COLUMN IF EXISTS profile_picture;
