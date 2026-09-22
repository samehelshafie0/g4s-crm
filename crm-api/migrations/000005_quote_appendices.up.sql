CREATE TABLE quote_appendices (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(), deleted_at timestamptz,
 quote_id uuid NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
 document_id uuid NOT NULL REFERENCES documents(id),
 document_version_id uuid NOT NULL REFERENCES document_versions(id),
 label varchar(255) NOT NULL, document_name text NOT NULL, version text NOT NULL,
 file_name text NOT NULL, file_type text NOT NULL, file_size bigint NOT NULL,
 sort_order integer NOT NULL CHECK(sort_order >= 0),
 UNIQUE(quote_id, document_version_id), UNIQUE(quote_id, sort_order)
);
CREATE INDEX quote_appendices_quote_id ON quote_appendices(quote_id);
