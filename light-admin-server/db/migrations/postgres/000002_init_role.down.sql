-- 000002_init_role.down.sql
-- Reverses 000002_init_role.up.sql. Drop the association first, then the role
-- table. Indexes are dropped implicitly with their tables.

DROP TABLE IF EXISTS t_role_menu;
DROP TABLE IF EXISTS t_role;
