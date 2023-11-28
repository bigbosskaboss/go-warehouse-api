-- warehouses
INSERT INTO warehouses (name, is_available)
VALUES
    ('Warehouse 1', true),
    ('Warehouse 2', true),
    ('Warehouse 3', false),
    ('Warehouse 4', true),
    ('Warehouse 5', false);

-- items
INSERT INTO items (id, name, size)
VALUES
    (1, 'Item 1', 'Small'),
    (2, 'Item 2', 'Medium'),
    (3, 'Item 3', 'Large'),
    (4, 'Item 4', 'Small'),
    (5, 'Item 5', 'Large');

-- warehouse_items
INSERT INTO warehouse_items (warehouse_id, item_id, amount, reserved_amount)
VALUES
    (1, 1, 100, 0),
    (1, 2, 50, 0),
    (1, 3, 200, 0),
    (2, 1, 150, 0),
    (2, 2, 70, 0),
    (2, 3, 80, 0),
    (3, 1, 60, 0),
    (3, 2, 30, 0),
    (3, 5, 70, 0),
    (4, 1, 50, 0),
    (4, 4, 100, 0),
    (4, 5, 120, 0),
    (5, 3, 90, 0),
    (5, 4, 130, 0),
    (5, 5, 45, 0);