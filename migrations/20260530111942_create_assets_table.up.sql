CREATE TABLE IF NOT EXISTS assets (
    asset_id VARCHAR(50) PRIMARY KEY,
    asset_name VARCHAR(255) NOT NULL,
    description TEXT,
    serial_number VARCHAR(255) UNIQUE NULL,
    quantity_stock INT NOT NULL DEFAULT 1,
    purchased_date DATE NULL,
    processor VARCHAR(20) NULL,
    ram VARCHAR(20) NULL,
    storage VARCHAR(20)  NULL,

    status ENUM(
        'available',
        'assigned',
        'maintenance',
        'retired'
    ) NOT NULL DEFAULT 'available',
 
    brand_id VARCHAR(50) NULL,
    category_id VARCHAR(50) NULL,

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_assets_brand
        FOREIGN KEY (brand_id)
        REFERENCES brands(brand_id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE,

    CONSTRAINT fk_assets_category
        FOREIGN KEY (category_id)
        REFERENCES categories(category_id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);