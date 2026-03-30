-- Add merchant approval fields

-- Add new columns for merchant approval workflow
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS business_license_url VARCHAR(500);
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS id_card_front_url VARCHAR(500);
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS id_card_back_url VARCHAR(500);
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS attachments JSONB;

-- Add status-related fields
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP;
ALTER TABLE merchants ADD COLUMN IF NOT EXISTS review_note TEXT;

-- Add index for status queries
CREATE INDEX IF NOT EXISTS idx_merchants_status ON merchants(status);
CREATE INDEX IF NOT EXISTS idx_merchants_reviewed_at ON merchants(reviewed_at);

-- Update existing records to set default values
UPDATE merchants SET 
    status = CASE 
        WHEN status = 'pending' AND business_license_url IS NOT NULL THEN 'reviewing'
        ELSE status 
    END
WHERE status = 'pending' AND business_license_url IS NOT NULL;
