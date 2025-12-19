CREATE TABLE IF NOT EXISTS comics (
    id INTEGER PRIMARY KEY,
    url TEXT NOT NULL,
    title TEXT NOT NULL,
    transcript TEXT,
    alt TEXT,
    image_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS comic_words (
    id SERIAL PRIMARY KEY,
    comic_id INTEGER REFERENCES comics(id) ON DELETE CASCADE,
    word TEXT NOT NULL,
    UNIQUE(comic_id, word)
);

CREATE TABLE IF NOT EXISTS search_index (
    word TEXT NOT NULL,
    comic_id INTEGER NOT NULL,
    PRIMARY KEY (word, comic_id),
    FOREIGN KEY (comic_id) REFERENCES comics(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_comic_words_word ON comic_words(word);
CREATE INDEX IF NOT EXISTS idx_comic_words_comic_id ON comic_words(comic_id);
CREATE INDEX IF NOT EXISTS idx_search_index_word ON search_index(word);
CREATE INDEX IF NOT EXISTS idx_search_index_comic_id ON search_index(comic_id);
