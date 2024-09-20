DROP TABLE IF EXISTS publication;

CREATE TABLE publication
(
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    annotation TEXT,
    publication_time INTEGER NOT NULL,
    publication_url TEXT NOT NULL,
    link TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS publication_title_uq
    ON public.publication USING btree
    (title COLLATE pg_catalog."default" ASC NULLS LAST)
    WITH (deduplicate_items=True)
    TABLESPACE pg_default;