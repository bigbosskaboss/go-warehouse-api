CREATE TABLE if not exists warehouses(
                                         id SERIAL PRIMARY KEY,
                                         name VARCHAR(255),
                                         is_available BOOLEAN
);