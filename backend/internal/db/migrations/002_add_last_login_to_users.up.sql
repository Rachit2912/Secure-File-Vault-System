-- aadding last_login time to users :
ALTER TABLE users
ADD COLUMN IF NOT EXISTS last_login TIMESTAMP;
