package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/denisenkom/go-mssqldb"

	"github.com/knq/dburl"

	"github.com/knq/xo/examples/booktest/mssql/models"
)

var flagVerbose = flag.Bool("v", false, "verbose")
var flagURL = flag.String("url", "mssql://booktest:booktest@localhost/booktest", "url")

func main() {
	var err error

	// set logging
	flag.Parse()
	if *flagVerbose {
		models.XOLog = func(s string, p ...interface{}) {
			fmt.Printf("-------------------------------------\nQUERY: %s\n  VAL: %v\n", s, p)
		}
	}

	// open database
	db, err := dburl.Open(*flagURL)
	if err != nil {
		log.Fatal(err)
	}

	// create an author
	a := models.Author{
		Name: "Unknown Master",
	}

	// save author to database
	err = a.Save(db)
	if err != nil {
		log.Fatal(err)
	}

	// create transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// save first book
	now := time.Now()
	b0 := models.Book{
		AuthorID:  a.AuthorID,
		Isbn:      "1",
		Title:     "my book title",
		Year:      2016,
		Available: now,
	}
	err = b0.Save(tx)
	if err != nil {
		log.Fatal(err)
	}

	// save second book
	b1 := models.Book{
		AuthorID:  a.AuthorID,
		Isbn:      "2",
		Title:     "the second book",
		Year:      2016,
		Available: now,
		Tags:      "cool unique",
	}
	err = b1.Save(tx)
	if err != nil {
		log.Fatal(err)
	}

	// update the title and tags
	b1.Title = "changed second title"
	b1.Tags = "cool disastor"
	err = b1.Update(tx)
	if err != nil {
		log.Fatal(err)
	}

	// save third book
	b2 := models.Book{
		AuthorID:  a.AuthorID,
		Isbn:      "3",
		Title:     "the third book",
		Year:      2001,
		Available: now,
		Tags:      "cool",
	}
	err = b2.Save(tx)
	if err != nil {
		log.Fatal(err)
	}

	// save fourth book
	b3 := models.Book{
		AuthorID:  a.AuthorID,
		Isbn:      "4",
		Title:     "4th place finisher",
		Year:      2011,
		Available: now,
		Tags:      "other",
	}
	err = b3.Save(tx)
	if err != nil {
		log.Fatal(err)
	}

	// tx commit
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	/*
		// upsert, changing ISBN and title
		b4 := models.Book{
			BookID:    b3.BookID,
			AuthorID:  a.AuthorID,
			Isbn:      "NEW ISBN",
			Title:     "never ever gonna finish, a quatrain",
			Year:      b3.Year,
			Available: b3.Available,
			Tags:      "someother",
		}
		err = b4.Upsert(db)
		if err != nil {
			log.Fatal(err)
		}
	*/

	// retrieve first book
	books0, err := models.BooksByTitleYear(db, "my book title", 2016)
	if err != nil {
		log.Fatal(err)
	}
	for _, book := range books0 {
		fmt.Printf("Book %d: %s available: %s\n", book.BookID, book.Title, book.Available.Format(time.RFC822Z))
		author, err := book.Author(db)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Book %d author: %s\n", book.BookID, author.Name)
	}

	// find a book with either "cool" or "other" tag
	fmt.Printf("---------\nTag search results:\n")
	res, err := models.AuthorBookResultsByTags(db, "cool")
	if err != nil {
		log.Fatal(err)
	}
	for _, ab := range res {
		fmt.Printf("Book %d: '%s', Author: '%s', ISBN: '%s' Tags: '%v'\n", ab.BookID, ab.BookTitle, ab.AuthorName, ab.BookIsbn, ab.BookTags)
	}

	// get book 4 and delete
	b5, err := models.BookByBookID(db, 4)
	if err != nil {
		log.Fatal(err)
	}
	err = b5.Delete(db)
	if err != nil {
		log.Fatal(err)
	}
}
