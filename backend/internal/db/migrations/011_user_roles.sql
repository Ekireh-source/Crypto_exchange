-- 011_user_roles.sql
-- Add role-based access control to the users table

ALTER TABLE users 
ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'user' 
CHECK (role IN ('user', 'admin', 'superadmin'));

-- Index on role for faster querying of admins
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
