ALTER TABLE links ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;
UPDATE links SET is_active = true WHERE is_active IS NULL; 