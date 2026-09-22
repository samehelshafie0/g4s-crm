-- PostgreSQL cannot drop a single enum value; rental services are retyped as maintenance instead.
UPDATE recurring_services SET service_type = 'maintenance' WHERE service_type = 'rental';
