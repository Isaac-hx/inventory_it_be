CREATE TABLE IF NOT EXISTS asset_assignments (
    assignment_id VARCHAR(50) PRIMARY KEY,
    status ENUM('assigned', 'returned', 'damaged', 'lost') NOT NULL DEFAULT 'returned',
    corporation VARCHAR(50) NULL,
    notes TEXT NOT NULL,
    asset_id VARCHAR(50) NOT NULL,
    assigned_by VARCHAR(50) NOT NULL,
    user_id VARCHAR(50) NOT NULL,
    assigned_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    return_date DATETIME NULL,

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP 
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_asset_assignments_asset
        FOREIGN KEY (asset_id)
        REFERENCES assets(asset_id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE, -- <-- Ditambahkan koma di sini

    CONSTRAINT fk_asset_assignments_user
        FOREIGN KEY (user_id)
        REFERENCES users(user_id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE, -- <-- Ditambahkan koma di sini

    CONSTRAINT fk_asset_assignments_assigned_by
        FOREIGN KEY (assigned_by)
        REFERENCES users(user_id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE -- <-- Baris terakhir sebelum tutup kurung TIDAK pakai koma
);