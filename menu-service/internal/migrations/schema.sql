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
);
