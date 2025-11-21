-- Add archived column to tasks table
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS archived BOOLEAN NOT NULL DEFAULT FALSE;

-- Create index for archived column
CREATE INDEX IF NOT EXISTS idx_tasks_archived ON tasks(archived);
