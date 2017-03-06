SELECT
  a.author_id::integer AS author_id,
  a.name::text AS author_name,
  b.book_id::integer AS book_id,
  b.isbn::text AS book_isbn,
  b.title::text AS book_title,
  b.tags::text[] AS book_tags
FROM books b
JOIN authors a ON a.author_id = b.author_id
WHERE b.tags && %%tags StringSlice%%::varchar[]
