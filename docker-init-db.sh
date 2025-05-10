#!/bin/bash
set -e

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
until PGPASSWORD=password psql -h postgres -U postgres -c '\l' > /dev/null 2>&1; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done

echo "PostgreSQL is up - executing initialization scripts"

# Initialize auth database
echo "Initializing auth database..."
PGPASSWORD=password psql -h postgres -U postgres -d auth -c "CREATE TABLE IF NOT EXISTS refresh_tokens (
  token TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  expires_at TIMESTAMP NOT NULL
);"

# Initialize users database  
echo "Initializing users database..."
PGPASSWORD=password psql -h postgres -U postgres -d users -c "CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    bio TEXT,
    role TEXT NOT NULL DEFAULT 'user'
);"

# Initialize menu database
echo "Initializing menu database..."
PGPASSWORD=password psql -h postgres -U postgres -d menu -c "
CREATE TABLE IF NOT EXISTS category (
    category_id UUID NOT NULL,
    category_pic TEXT,
    name TEXT,
    CONSTRAINT category_pkey PRIMARY KEY (category_id)
);

CREATE TABLE IF NOT EXISTS items (
    item_id UUID PRIMARY KEY,
    category_id UUID,
    name TEXT,
    price REAL,
    is_available BOOLEAN,
    quantity INTEGER,
    has_discount BOOLEAN,
    discount_value REAL,
    CONSTRAINT items_category_id_fkey FOREIGN KEY (category_id)
        REFERENCES category (category_id)
);"

# Insert sample data into menu database
echo "Inserting sample data into menu database..."
PGPASSWORD=password psql -h postgres -U postgres -d menu -c "
INSERT INTO category (category_id, name, category_pic)
VALUES
    ('f54c7e1d-34b4-4b89-8703-90090edf5fb1', 'ام علي', 'https://ipficywkrkybhrhnvjxe.supabase.co/storage/v1/object/public/song-pic//om-3li.PNG'),
    ('b76bca2e-b6aa-47e0-ab09-0c0644797dac', 'قشطوطه', 'https://ipficywkrkybhrhnvjxe.supabase.co/storage/v1/object/public/song-pic//2a4tota.PNG'),
    ('d7da3730-9bf2-4c8c-bad1-dc97616fd2d9', 'قشطوظه', 'https://ipficywkrkybhrhnvjxe.supabase.co/storage/v1/object/public/song-pic//2a4toza.PNG'),
    ('62ad42c0-0b5f-4969-bf51-803b424b5c63', 'كشري', 'https://ipficywkrkybhrhnvjxe.supabase.co/storage/v1/object/public/song-pic//koshary.PNG')
ON CONFLICT (category_id) DO NOTHING;"

# Insert sample data into items table
PGPASSWORD=password psql -h postgres -U postgres -d menu -c "
INSERT INTO items (item_id, category_id, name, price, is_available, quantity, has_discount, discount_value)
VALUES
    ('d704184b-7858-4450-a163-b17335b2c4f7', 'f54c7e1d-34b4-4b89-8703-90090edf5fb1', 'طاجن ام علي البركة', 70.0, true, 50, false, 0),
    ('a0a25e8c-c750-4e29-8741-de4a201beae0', 'f54c7e1d-34b4-4b89-8703-90090edf5fb1', 'طاجن ام علي بالسمنة البلدي', 35.0, true, 50, false, 0),
    ('695a7fd2-5fea-4bb9-8d40-e40ac8623fe1', 'f54c7e1d-34b4-4b89-8703-90090edf5fb1', 'طاجن ام علي بالسمنة البلدي + قشطة', 50.0, true, 50, false, 0),
    ('3dc0ad8f-99bb-48e8-a64f-b1bdb0a795ae', 'f54c7e1d-34b4-4b89-8703-90090edf5fb1', 'طاجن ام علي بالسمنة البلدي و مكسرات', 60.0, true, 50, false, 0),
    ('3de1f5b8-9da5-432d-a482-0e7279d4a457', 'f54c7e1d-34b4-4b89-8703-90090edf5fb1', 'طاجن ام علي قشطة و مكسرات', 70.0, true, 50, false, 0),
    ('37907130-b513-42a5-9577-e9c59ee2e0a3', 'b76bca2e-b6aa-47e0-ab09-0c0644797dac', 'قشططة رز ب لبن كريمة', 55.0, true, 10, false, 0),
    ('39b8c4d6-20a6-4b3a-a2e9-1994d489ca82', 'd7da3730-9bf2-4c8c-bad1-dc97616fd2d9', 'قشطوظة تحفة', 90.0, true, 100, true, 25.0),
    ('afded924-87f8-49a7-b759-4bff7147dba1', '62ad42c0-0b5f-4969-bf51-803b424b5c63', 'كشري أوريو نوتيلا', 65.0, true, 50, false, 0)
ON CONFLICT (item_id) DO NOTHING;"

# Initialize MongoDB database (for order-service)
echo "Initializing MongoDB database for orders..."
mongosh --host mongo << EOF
use orders
db.createCollection("orders")
EOF

echo "Database initialization complete!" 