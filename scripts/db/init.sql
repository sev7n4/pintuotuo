-- Initial database schema will be created in Week 2
-- This file is a placeholder for the full schema

-- Database: pintuotuo_db (created automatically by docker-compose)

-- TODO: Add schema creation in Week 2 after architecture review
-- Reference: 03_Data_Model_Design.md for complete schema

CREATE SCHEMA IF NOT EXISTS public;

-- Placeholder tables - to be replaced with real schema
CREATE TABLE IF NOT EXISTS schema_version (
    version_id SERIAL PRIMARY KEY,
    description VARCHAR(255) NOT NULL,
    installed_on TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO schema_version (description) VALUES
('Initial schema - Week 2 planning');
