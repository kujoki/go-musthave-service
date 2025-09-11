ALTER TABLE IF EXISTS url_data
ADD CONSTRAINT short_urls_original_url_key UNIQUE (long_url);