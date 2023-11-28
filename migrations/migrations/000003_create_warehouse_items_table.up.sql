CREATE TABLE if not exists warehouse_items (
                                               id SERIAL PRIMARY KEY,
                                               warehouse_id INT,
                                               item_id INT,
                                               amount INT,
                                               reserved_amount INT,
                                               FOREIGN KEY (warehouse_id) REFERENCES warehouses(id),
                                               FOREIGN KEY (item_id) REFERENCES items(id)
);