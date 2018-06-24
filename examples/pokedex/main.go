package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/xo/dburl"
)

var (
	flagVerbose = flag.Bool("v", false, "toggle verbose")
	flagURL     = flag.String("url", "", "database url")
	flagSeed    = flag.Int64("seed", time.Now().UnixNano(), "rand seed")
)

func main() {
	var err error

	flag.Parse()

	// set verbose
	if *flagVerbose {
		XOLog = func(s string, p ...interface{}) {
			fmt.Printf("-------------------------------------\nQUERY: %s\n  VAL: %v\n", s, p)
		}
	}

	// open database
	db, err := dburl.Open(*flagURL)
	if err != nil {
		log.Fatal(err)
	}

	// get random id
	r := rand.New(rand.NewSource(*flagSeed))
	id := r.Intn(720) + 1

	// lookup pokemon
	log.Printf("looking up pokemon with id %d", id)
	p, err := PokemonByID(db, id)
	if err != nil {
		log.Fatal(err)
	}

	// get the pokemon's species
	species, err := p.PokemonSpecies(db)
	if err != nil {
		log.Fatal(err)
	}

	// display
	log.Printf(
		"pokemon #%d: `%s` (species: `%s`, height: %d, weight: %d)\n",
		p.ID, p.Identifier, species.Identifier, p.Height, p.Weight,
	)
}
