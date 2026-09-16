-- Empty optional registration numbers are not business identifiers.
UPDATE customers SET cr_number = NULL WHERE btrim(cr_number) = '';
