CREATE TABLE devices (
    id BIGSERIAL PRIMARY KEY,
    device_code VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    location VARCHAR(255),
    status VARCHAR(50) DEFAULT 'inactive',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE sensors (
    id BIGSERIAL PRIMARY KEY,
    device_id BIGINT NOT NULL REFERENCES devices(id),
    sensor_code VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    unit VARCHAR(50),
    status VARCHAR(50) DEFAULT 'inactive',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE monitoring_data (
    id BIGSERIAL PRIMARY KEY,
    device_id BIGINT NOT NULL REFERENCES devices(id),
    sensor_id BIGINT NOT NULL REFERENCES sensors(id),

    value DOUBLE PRECISION NOT NULL,
    unit VARCHAR(50),

    status VARCHAR(50) DEFAULT 'pending',

    scheduled_at TIMESTAMP NULL,
    sent_at TIMESTAMP NULL,

    retry_count INT DEFAULT 0,
    next_retry_at TIMESTAMP NULL,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE delivery_logs (
    id BIGSERIAL PRIMARY KEY,

    monitoring_data_id BIGINT NOT NULL
    REFERENCES monitoring_data(id),

    target_url TEXT NOT NULL,

    status VARCHAR(50) NOT NULL,

    status_code INT,

    error_message TEXT,

    attempt INT DEFAULT 1,

    created_at TIMESTAMP DEFAULT NOW()
);