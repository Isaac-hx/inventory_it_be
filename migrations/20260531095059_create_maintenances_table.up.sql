CREATE TABLE IF NOT EXISTS maintenances (
    maintenance_id VARCHAR(50) PRIMARY KEY,
    assignment_id VARCHAR(50) NOT NULL,
    status ENUM(
        'pending',
        'in_progress',
        'completed',
        'cancelled'
    ) NOT NULL DEFAULT 'pending',
    description TEXT,
    cost BIGINT NOT NULL DEFAULT 0,
    maintenance_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP 
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_maintenances_assignment
        FOREIGN KEY (assignment_id)
        REFERENCES asset_assignments(assignment_id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);