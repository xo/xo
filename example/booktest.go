package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/knq/xo/example/models"
)

func main() {
	var err error

	db, err := sql.Open("postgres", "postgres://booktest:booktest@localhost/booktest")
	if err != nil {
		log.Fatal(err)
	}

	// create an author
	a := models.Author{
		Isbn:    "some isbn",
		Name:    "Author Name",
		Subject: "a subject",
	}

	// save author to database
	err = a.Save(db)
	if err != nil {
		log.Fatal(err)
	}

	// create book in a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()
	b := models.Book{
		AuthorID:  a.AuthorID,
		Title:     "my book title",
		Booktype:  models.FictionBookType,
		Year:      2016,
		Available: &now,
	}
	err = b.Save(db)
	if err != nil {
		log.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	// retrieve created book
	books, err := models.BooksByTitle(db, "my book title", 2016)
	if err != nil {
		log.Fatal(err)
	}
	for _, book := range books {
		fmt.Printf("Book %d (%s): %s available: %s\n", book.BookID, book.Booktype, book.Title, book.Available.Format(time.RFC822Z))
		author, err := book.Author(db)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Book %d author: %s\n", book.BookID, author.Name)
	}

	// call say_hello(varchar)
	str, err := models.SayHello(db, "john")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("SayHello response: %s\n", str)
}
