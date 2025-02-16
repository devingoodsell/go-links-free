DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT FROM information_schema.columns 
        WHERE table_name = 'links' AND column_name = 'is_active'
    ) THEN
        ALTER TABLE links ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;
    END IF;
END $$; 