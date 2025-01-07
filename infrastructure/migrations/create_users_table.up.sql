CREATE TABLE users (
    id CHAR(36) PRIMARY KEY,           -- UUID type, use CHAR(36) for storing UUID as a string
    name VARCHAR(100) NOT NULL,        -- Name field
    email VARCHAR(100) UNIQUE NOT NULL, -- Email field
    password VARCHAR(255) NULL,         -- Password field, nullable
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Created timestamp with default value
);
