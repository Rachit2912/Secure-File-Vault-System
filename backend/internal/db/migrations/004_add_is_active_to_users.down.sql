-- removing is_active col from users:
ALTER TABLE users
DROP COLUMN IF EXISTS is_active;

