CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  email VARCHAR(255) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
  fname VARCHAR(100),
  surname VARCHAR(100),
  gender VARCHAR(10) CHECK (gender IN ('male','female','other')),
  birthdate DATE,
  about TEXT,
  hobbies JSONB,
  interests JSONB,
  country VARCHAR(100),
  city VARCHAR(100),

  -- Preferences: who are they looking for, and age range
  looking_for_gender VARCHAR(10) CHECK (looking_for_gender IN ('male','female','other','any')),
  looking_for_min_age INT,
  looking_for_max_age INT,
  
  profile_picture_url TEXT DEFAULT NULL,
  
  preferred_hobbies JSONB,
  preferred_interests JSONB,

  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);

ALTER TABLE users
ADD COLUMN IF NOT EXISTS preferred_hobbies JSONB,
ADD COLUMN IF NOT EXISTS preferred_interests JSONB;