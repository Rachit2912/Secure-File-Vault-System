-- removing last_login time from users :
ALTER TABLE users
DROP COLUMN IF EXISTS last_login;
