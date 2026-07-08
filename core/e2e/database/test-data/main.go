// Command generate_cafe_data emits cafe_data.sql: the Postgres seed for the
// sample "Copper Kettle Coffee & Bakery" database used by the e2e database tests
// (startPostgres in containers_test.go) and the dev stack (docker-compose.dev.yml
// runs it as an initdb script; databases.dev.yaml points the engine at it).
//
// The schema DDL and the trailing PK/UNIQUE constraints are carried verbatim
// (embedded from schema.sql / constraints.sql); the generator only produces the
// data section. Rows are emitted as INSERT ... VALUES (not COPY) so time columns
// can be load-relative expressions — every timestamp is rendered as
// `now() - interval '<offset>'`, so the data always ends ~today no matter when
// the dump is loaded. A fixed RNG seed makes the output deterministic
// (byte-identical between runs) for clean diffs.
//
// The generator is split across several files in this package (main, rng, types,
// dimensions, menu, customers, orders, emit, sql); run the whole package:
//
//	go run .
//
//go:generate go run .

package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// seed fixes the RNG so regeneration is deterministic.
const seed = 20260117

// window is how far back the oldest orders reach from load time.
const window = 180 * 24 * time.Hour
func main() {
	g := &gen{rng: rand.New(rand.NewSource(seed))}
	g.buildDimensions()
	g.buildOrders()

	var b strings.Builder
	b.WriteString(schemaSQL)
	b.WriteString("\n")
	g.writeData(&b)
	b.WriteString(constraintsSQL)

	if err := os.WriteFile("cafe_data.sql", []byte(b.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("wrote cafe_data.sql: %d customers, %d menu items, %d orders\n",
		len(g.customers), len(g.menuItems), len(g.orders))
}

// gen holds the RNG and the generated rows.
type gen struct {
	rng *rand.Rand

	locations  []location
	allergens  []allergen
	categories []category
	menuItems  []menuItem
	variants   []variant
	modGroups  []modGroup
	modifiers  []modifier
	itemMods   []itemModGroup // menu_item_modifier_groups
	itemAllerg []itemAllergen // menu_item_allergens
	customers  []customer
	addresses  []address
	drivers    []driver
	promos     []promo
	loyalty    []loyaltyAccount

	orders     []order
	orderItems []orderItem
	oiMods     []oiModifier
	payments   []payment
	deliveries []delivery
	reviews    []review
	orderPromo []orderPromotion
	loyaltyTxn []loyaltyTxn
	notifs     []notification
	carts      []cart
	cartItems  []cartItem

	// lookups built during buildDimensions and read during order generation.
	variantsByItem map[string][]variant
	groupsByItem   map[string][]string
	modsByGroup    map[string][]modifier
	addrByCust     map[string][]address
	loyaltyByCust  map[string]*loyaltyAccount
	orderSeq       int
}
