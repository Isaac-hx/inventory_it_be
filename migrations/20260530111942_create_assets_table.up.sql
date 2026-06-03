CREATE TABLE IF NOT EXISTS assets (
    asset_id VARCHAR(50) PRIMARY KEY,
    asset_name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    serial_number VARCHAR(255) UNIQUE,
    purchased_date DATE,

    status ENUM(
        'available',
        'assigned',
        'maintenance',
        'retired'
    ) DEFAULT 'available',

    brand_id VARCHAR(50),
    category_id VARCHAR(50),

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_assets_brand
        FOREIGN KEY (brand_id)
        REFERENCES brands(brand_id)
        ON DELETE SET NULL
        ON UPDATE CASCADE,

    CONSTRAINT fk_assets_category
        FOREIGN KEY (category_id)
        REFERENCES categories(category_id)
        ON DELETE SET NULL
        ON UPDATE CASCADE
);