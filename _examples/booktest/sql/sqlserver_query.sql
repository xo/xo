SELECT
  a.author_id AS author_id,
  a.name AS author_name,
  b.book_id AS book_id,
  b.isbn AS book_isbn,
  b.title AS book_title,
  b.tags AS book_tags
FROM books b
JOIN authors a ON a.author_id = b.author_id
WHERE b.tags LIKE '%' + %%tags string%% + '%'
