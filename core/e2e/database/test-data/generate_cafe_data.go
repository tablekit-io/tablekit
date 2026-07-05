//go:build ignore

// Command generate_cafe_data emits cafe_data.sql: the Postgres seed for the
// sample "Copper Kettle Coffee & Bakery" database used by the e2e database tests
// (startPostgres in containers_test.go) and the dev stack (docker-compose.dev.yml
// runs it as an initdb script; databases.dev.yaml points the engine at it).
//
// The schema DDL and the trailing PK/UNIQUE constraints are carried verbatim; the
// generator only produces the data section. Rows are emitted as INSERT ... VALUES
// (not COPY) so time columns can be load-relative expressions — every timestamp is
// rendered as `now() - interval '<offset>'`, so the data always ends ~today no
// matter when the dump is loaded. A fixed RNG seed makes the output deterministic
// (byte-identical between runs) for clean diffs.
//
// Run:
//
//	go run generate_cafe_data.go
//
//go:generate go run generate_cafe_data.go
package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
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

// ---- RNG + rendering helpers ------------------------------------------------

func (g *gen) uuid() string {
	var u [16]byte
	g.rng.Read(u[:])
	u[6] = (u[6] & 0x0f) | 0x40 // version 4
	u[8] = (u[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}

func (g *gen) intn(n int) int      { return g.rng.Intn(n) }
func (g *gen) chance(p float64) bool { return g.rng.Float64() < p }

// pick returns a uniformly random element.
func pick[T any](g *gen, xs []T) T { return xs[g.rng.Intn(len(xs))] }

// weightedIndex returns an index chosen by the given weights.
func (g *gen) weightedIndex(weights []float64) int {
	total := 0.0
	for _, w := range weights {
		total += w
	}
	r := g.rng.Float64() * total
	for i, w := range weights {
		r -= w
		if r < 0 {
			return i
		}
	}
	return len(weights) - 1
}

// s renders a SQL string literal (single quotes doubled).
func s(v string) string { return "'" + strings.ReplaceAll(v, "'", "''") + "'" }

// ns renders a nullable string: empty -> NULL.
func ns(v string) string {
	if v == "" {
		return "NULL"
	}
	return s(v)
}

// money renders a numeric with two decimals.
func money(v float64) string { return strconv.FormatFloat(v, 'f', 2, 64) }

func b2(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func jsonbArray(items []string) string {
	if len(items) == 0 {
		return "'[]'::jsonb"
	}
	quoted := make([]string, len(items))
	for i, it := range items {
		quoted[i] = `"` + strings.ReplaceAll(it, `"`, `\"`) + `"`
	}
	return s("["+strings.Join(quoted, ", ")+"]") + "::jsonb"
}

// rel renders a load-relative timestamp: now() - interval '<age>'. A non-positive
// age clamps to now().
func rel(age time.Duration) string {
	if age < 0 {
		age = 0
	}
	total := int64(age / time.Second)
	days := total / 86400
	rem := total % 86400
	h := rem / 3600
	m := (rem % 3600) / 60
	sec := rem % 60
	return fmt.Sprintf("now() - interval '%d days %d hours %d mins %d secs'", days, h, m, sec)
}

// slug turns a name into a URL slug.
func slug(name string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// ---- dimension types --------------------------------------------------------

type location struct {
	id, name, slug, address, phone, opensAt, closesAt string
	lat, lng                                          float64
	metroIdx                                          int
}

type allergen struct{ id, name, slug, emoji string }

type category struct {
	id, name, slug string
	sort           int
	kind           string // drink | pastry | bagel | food
}

type menuItem struct {
	id, categoryID, name, slug, description, imageURL string
	basePrice                                         float64
	veg, vegan, signature                             bool
	calories, prepMinutes                             int
	tags                                              []string
	catKind                                           string
}

type variant struct {
	id, itemID, name string
	priceDelta       float64
	isDefault        bool
	sort             int
}

type modGroup struct {
	id, name           string
	selMin, selMax, srt int
}

type modifier struct {
	id, groupID, name string
	priceDelta        float64
	isDefault         bool
}

type itemModGroup struct {
	id, itemID, groupID string
	required            bool
	sort                int
}

type itemAllergen struct{ id, itemID, allergenID string }

type customer struct {
	id, fullName, email, phone, defaultAddressID, status string
	marketingOptIn                                       bool
	lastLoginAge, createdAge                             time.Duration
}

type address struct {
	id, customerID, label, line1, line2, area, city, postal, instructions string
	lat, lng                                                              float64
	isDefault                                                             bool
	metroIdx                                                              int
}

type driver struct {
	id, fullName, phone, email, vehicle, plate, status string
	lat, lng, rating                                   float64
	totalDeliveries                                    int
	hiredDaysAgo                                       int
	homeLocationID                                     string
}

type promo struct {
	id, code, description, discountType string
	discountValue, minOrder, maxDiscount float64
	hasMaxDiscount                       bool
}

type loyaltyAccount struct {
	id, customerID, tier string
	balance, lifetime    int
}

// ---- fact types -------------------------------------------------------------

type order struct {
	id, customerID, addressID, number, status, paymentMethod, locationID string
	placedAge                                                            time.Duration
	subtotal, deliveryFee, discount, tax, grandTotal                     float64
	promoID                                                              string // "" if none
	delivered                                                            bool
}

type orderItem struct {
	id, orderID, menuItemID, variantID, snapshotName string
	quantity                                         int
	unitPrice, lineTotal                             float64
}

type oiModifier struct {
	id, orderItemID, modifierID, snapshotName string
	snapshotPriceDelta                        float64
}

type payment struct {
	id, orderID, customerID, method, providerRef, status string
	amount                                               float64
	capturedAge                                          time.Duration
	captured                                             bool
}

type delivery struct {
	id, orderID, driverID, status string
	assignedAge, deliveredAge     time.Duration
	distanceKM                    float64
}

type review struct {
	id, orderID, customerID, menuItemID, driverID, title, body string
	rating                                                     int
	createdAge                                                 time.Duration
}

type orderPromotion struct {
	id, orderID, promoID, customerID string
	discount                         float64
	appliedAge                       time.Duration
}

type loyaltyTxn struct {
	id, accountID, customerID, orderID, kind, description string
	points, balanceAfter                                 int
	createdAge                                           time.Duration
}

type notification struct {
	id, customerID, channel, kind, title, body string
	createdAge                                 time.Duration
	read                                       bool
}

type cart struct {
	id, customerID, locationID string
	subtotal                   float64
	createdAge                 time.Duration
}

type cartItem struct {
	id, cartID, menuItemID, variantID, notes string
	quantity                                 int
	lineTotal                                float64
}

// metro is a US city the sample brand operates in; its sales-tax rate drives the
// per-order tax so amounts vary realistically by location.
type metro struct {
	city, state, zip, area string
	streets                []string
	lat, lng, taxRate      float64
}

var metros = []metro{
	{"Portland", "OR", "97209", "503",
		[]string{"NW Lovejoy St", "NW 13th Ave", "NW Marshall St", "NW Johnson St", "N Mississippi Ave", "SE Division St"},
		45.5300, -122.6850, 0.0}, // OR has no sales tax
	{"Austin", "TX", "78704", "512",
		[]string{"S Congress Ave", "E 6th St", "S 1st St", "Rainey St", "Barton Springs Rd", "E Cesar Chavez St"},
		30.2480, -97.7500, 0.0825},
	{"Denver", "CO", "80205", "720",
		[]string{"Larimer St", "Blake St", "Walnut St", "Wynkoop St", "Brighton Blvd", "Market St"},
		39.7600, -104.9830, 0.0881},
}

// ---- dimension building -----------------------------------------------------

func (g *gen) buildDimensions() {
	g.variantsByItem = map[string][]variant{}
	g.groupsByItem = map[string][]string{}
	g.modsByGroup = map[string][]modifier{}
	g.addrByCust = map[string][]address{}
	g.loyaltyByCust = map[string]*loyaltyAccount{}

	g.buildLocations()
	g.buildAllergens()
	g.buildMenu()
	g.buildModifiers()
	g.buildCustomers()
	g.buildDrivers()
	g.buildPromos()
}

func (g *gen) buildLocations() {
	specs := []struct {
		name, area, street, zip string
		lat, lng                float64
		metro                   int
	}{
		{"Copper Kettle Coffee & Bakery — Pearl District", "Pearl District", "1420 NW Lovejoy St", "97209", 45.5300, -122.6850, 0},
		{"Copper Kettle Coffee & Bakery — South Congress", "South Congress", "1601 S Congress Ave", "78704", 30.2480, -97.7500, 1},
		{"Copper Kettle Coffee & Bakery — RiNo", "RiNo", "2601 Larimer St", "80205", 39.7600, -104.9830, 2},
	}
	for i, spec := range specs {
		m := metros[spec.metro]
		g.locations = append(g.locations, location{
			id:       fmt.Sprintf("00000001-0000-7000-8000-%012d", i+1),
			name:     spec.name,
			slug:     slug(spec.area),
			address:  fmt.Sprintf("%s, %s, %s %s", spec.street, m.city, m.state, spec.zip),
			phone:    fmt.Sprintf("+1 (%s) 555-01%02d", m.area, i+1),
			opensAt:  "07:00:00",
			closesAt: "21:00:00",
			lat:      spec.lat,
			lng:      spec.lng,
			metroIdx: spec.metro,
		})
	}
}

func (g *gen) buildAllergens() {
	specs := []struct{ name, emoji string }{
		{"Gluten", "🌾"}, {"Dairy", "🥛"}, {"Eggs", "🥚"}, {"Tree Nuts", "🌰"},
		{"Peanuts", "🥜"}, {"Soy", "🫘"}, {"Sesame", "🫓"}, {"Wheat", "🌾"},
	}
	for _, spec := range specs {
		g.allergens = append(g.allergens, allergen{
			id: g.uuid(), name: spec.name, slug: slug(spec.name), emoji: spec.emoji,
		})
	}
}

func (g *gen) allergenID(name string) string {
	for _, a := range g.allergens {
		if a.name == name {
			return a.id
		}
	}
	return ""
}

func (g *gen) buildMenu() {
	catSpecs := []struct{ name, kind string }{
		{"Breakfast", "food"}, {"Coffee", "drink"}, {"Cold Drinks", "drink"},
		{"Smoothies", "drink"}, {"Pastries", "pastry"}, {"Cakes", "pastry"},
		{"Bagels", "bagel"}, {"Sandwiches", "food"}, {"Salads", "food"},
	}
	catID := map[string]string{}
	for i, spec := range catSpecs {
		id := g.uuid()
		catID[spec.name] = id
		g.categories = append(g.categories, category{
			id: id, name: spec.name, slug: slug(spec.name), sort: i + 1, kind: spec.kind,
		})
	}

	type is struct {
		cat        string
		name       string
		price      float64
		veg, vegan bool
		sig        bool
		cal        int
	}
	items := []is{
		// Coffee
		{"Coffee", "Maple Mocha", 5.25, true, false, true, 320},
		{"Coffee", "Flat White", 4.50, true, false, false, 180},
		{"Coffee", "Cappuccino", 4.25, true, false, false, 160},
		{"Coffee", "Caffè Latte", 4.75, true, false, false, 190},
		{"Coffee", "Americano", 3.50, true, true, false, 15},
		{"Coffee", "Cortado", 4.00, true, false, false, 130},
		{"Coffee", "Espresso", 3.00, true, true, false, 5},
		{"Coffee", "Caramel Macchiato", 5.25, true, false, false, 250},
		{"Coffee", "Vanilla Latte", 5.00, true, false, false, 230},
		{"Coffee", "Drip Coffee", 3.25, true, true, false, 5},
		{"Coffee", "Pour Over", 5.50, true, true, true, 5},
		{"Coffee", "Honey Oat Latte", 5.50, true, false, false, 240},
		{"Coffee", "Mocha", 5.00, true, false, false, 290},
		// Cold Drinks
		{"Cold Drinks", "Iced Latte", 5.00, true, false, false, 190},
		{"Cold Drinks", "Iced Americano", 4.00, true, true, false, 15},
		{"Cold Drinks", "Cold Brew", 4.75, true, true, false, 5},
		{"Cold Drinks", "Nitro Cold Brew", 5.50, true, true, true, 10},
		{"Cold Drinks", "Iced Matcha Latte", 5.75, true, false, false, 210},
		{"Cold Drinks", "Iced Chai Latte", 5.25, true, false, false, 240},
		{"Cold Drinks", "Sparkling Lemonade", 4.25, true, true, false, 120},
		{"Cold Drinks", "Iced Tea", 3.75, true, true, false, 60},
		{"Cold Drinks", "Affogato", 6.00, true, false, false, 220},
		// Smoothies
		{"Smoothies", "Mango Smoothie", 6.50, true, false, false, 280},
		{"Smoothies", "Berry Blast Smoothie", 6.75, true, false, false, 260},
		{"Smoothies", "Green Machine Smoothie", 7.00, true, true, true, 240},
		{"Smoothies", "Peanut Butter Banana Smoothie", 6.75, true, false, false, 340},
		{"Smoothies", "Strawberry Banana Smoothie", 6.50, true, false, false, 270},
		// Breakfast
		{"Breakfast", "Avocado Toast", 9.50, true, true, true, 420},
		{"Breakfast", "Breakfast Burrito", 10.50, false, false, false, 610},
		{"Breakfast", "Veggie Scramble", 11.00, true, false, false, 480},
		{"Breakfast", "Egg & Cheese Sandwich", 7.50, false, false, false, 450},
		{"Breakfast", "Buttermilk Pancakes", 9.00, true, false, false, 560},
		{"Breakfast", "Steel-Cut Oatmeal", 6.50, true, true, false, 320},
		{"Breakfast", "Greek Yogurt Parfait", 6.00, true, false, false, 290},
		{"Breakfast", "Breakfast Bagel Sandwich", 8.50, false, false, false, 520},
		// Pastries
		{"Pastries", "Butter Croissant", 4.00, true, false, false, 280},
		{"Pastries", "Pain au Chocolat", 4.50, true, false, false, 340},
		{"Pastries", "Almond Croissant", 4.75, true, false, false, 410},
		{"Pastries", "Cinnamon Roll", 4.75, true, false, true, 430},
		{"Pastries", "Blueberry Muffin", 3.75, true, false, false, 380},
		{"Pastries", "Banana Bread", 3.95, true, false, false, 360},
		{"Pastries", "Buttermilk Scone", 3.75, true, false, false, 350},
		{"Pastries", "Chocolate Chip Cookie", 3.25, true, false, false, 300},
		// Cakes
		{"Cakes", "Carrot Cake Slice", 6.50, true, false, false, 520},
		{"Cakes", "Chocolate Cake Slice", 6.75, true, false, true, 560},
		{"Cakes", "New York Cheesecake Slice", 7.00, true, false, false, 480},
		{"Cakes", "Red Velvet Slice", 6.75, true, false, false, 540},
		{"Cakes", "Lemon Bar", 4.50, true, false, false, 320},
		// Bagels
		{"Bagels", "Plain Bagel", 2.75, true, true, false, 250},
		{"Bagels", "Everything Bagel", 3.00, true, true, false, 270},
		{"Bagels", "Sesame Bagel", 3.00, true, true, false, 260},
		{"Bagels", "Cinnamon Raisin Bagel", 3.25, true, true, false, 280},
		{"Bagels", "Bagel with Cream Cheese", 4.50, true, false, false, 380},
		// Sandwiches
		{"Sandwiches", "Turkey Club", 11.50, false, false, true, 640},
		{"Sandwiches", "Caprese Panini", 10.50, true, false, false, 560},
		{"Sandwiches", "Grilled Cheese", 8.50, true, false, false, 520},
		{"Sandwiches", "Chicken Pesto Panini", 11.00, false, false, false, 610},
		{"Sandwiches", "Veggie Wrap", 9.50, true, true, false, 440},
		{"Sandwiches", "BLT", 9.75, false, false, false, 530},
		// Salads
		{"Salads", "Caesar Salad", 10.50, true, false, false, 380},
		{"Salads", "Cobb Salad", 12.00, false, false, true, 480},
		{"Salads", "Greek Salad", 10.75, true, false, false, 360},
		{"Salads", "Harvest Grain Bowl", 12.50, true, true, false, 520},
		{"Salads", "Kale Caesar", 11.00, true, false, false, 400},
	}

	seenSlug := map[string]bool{}
	for _, it := range items {
		catKind := ""
		for _, c := range g.categories {
			if c.name == it.cat {
				catKind = c.kind
			}
		}
		sl := slug(it.name)
		for seenSlug[sl] {
			sl += "-x"
		}
		seenSlug[sl] = true
		id := g.uuid()
		tags := []string{}
		switch catKind {
		case "drink":
			if strings.Contains(it.name, "Iced") || strings.Contains(it.name, "Cold") || it.cat == "Cold Drinks" || it.cat == "Smoothies" {
				tags = []string{"cold"}
			} else {
				tags = []string{"hot"}
			}
		case "pastry", "bagel":
			tags = []string{"baked"}
		default:
			tags = []string{"food"}
		}
		m := menuItem{
			id: id, categoryID: catID[it.cat], name: it.name, slug: sl,
			description: it.name + ".",
			basePrice:   it.price,
			veg:         it.veg, vegan: it.vegan, signature: it.sig,
			calories: it.cal, prepMinutes: 2 + g.intn(9),
			imageURL: "https://cdn.example.com/menu/" + sl + ".jpg",
			tags:     tags, catKind: catKind,
		}
		g.menuItems = append(g.menuItems, m)
		g.buildVariants(m)
		g.buildItemAllergens(m)
	}
}

func (g *gen) buildVariants(m menuItem) {
	var vs []variant
	if m.catKind == "drink" {
		for i, spec := range []struct {
			name  string
			delta float64
			def   bool
		}{{"Small", -0.50, false}, {"Medium", 0, true}, {"Large", 0.85, false}} {
			vs = append(vs, variant{id: g.uuid(), itemID: m.id, name: spec.name, priceDelta: spec.delta, isDefault: spec.def, sort: i})
		}
	} else {
		vs = append(vs, variant{id: g.uuid(), itemID: m.id, name: "Regular", priceDelta: 0, isDefault: true, sort: 0})
	}
	g.variants = append(g.variants, vs...)
	g.variantsByItem[m.id] = vs
}

func (g *gen) buildItemAllergens(m menuItem) {
	add := func(name string) {
		id := g.allergenID(name)
		if id == "" {
			return
		}
		g.itemAllerg = append(g.itemAllerg, itemAllergen{id: g.uuid(), itemID: m.id, allergenID: id})
	}
	switch m.catKind {
	case "drink":
		if !m.vegan {
			add("Dairy")
		}
	case "pastry", "bagel":
		add("Gluten")
		add("Wheat")
		if !m.vegan {
			add("Dairy")
			add("Eggs")
		}
	case "food":
		add("Gluten")
		add("Wheat")
	}
	if strings.Contains(m.name, "Almond") {
		add("Tree Nuts")
	}
	if strings.Contains(m.name, "Peanut") {
		add("Peanuts")
	}
}

func (g *gen) buildModifiers() {
	type modSpec struct {
		name  string
		delta float64
		def   bool
	}
	groups := []struct {
		name           string
		selMin, selMax int
		mods           []modSpec
		forKinds       []string
	}{
		{"Milk", 0, 1, []modSpec{{"Whole Milk", 0, true}, {"Oat Milk", 0.75, false}, {"Almond Milk", 0.75, false}, {"Nonfat Milk", 0, false}, {"Soy Milk", 0.75, false}}, []string{"Coffee", "Cold Drinks"}},
		{"Extra Shot", 0, 2, []modSpec{{"Add Espresso Shot", 0.90, false}}, []string{"Coffee"}},
		{"Syrup", 0, 3, []modSpec{{"Vanilla Syrup", 0.60, false}, {"Caramel Syrup", 0.60, false}, {"Hazelnut Syrup", 0.60, false}, {"Sugar-Free Vanilla", 0.60, false}}, []string{"Coffee", "Cold Drinks"}},
		{"Toppings", 0, 2, []modSpec{{"Butter", 0.50, false}, {"House Jam", 0.60, false}, {"Cream Cheese", 0.75, false}, {"Whipped Cream", 0.50, false}}, []string{"Pastries", "Cakes", "Bagels"}},
		{"Bagel Prep", 0, 1, []modSpec{{"Toasted", 0, false}, {"Not Toasted", 0, true}}, []string{"Bagels"}},
	}

	groupIDByName := map[string]string{}
	forCategories := map[string][]string{} // category name -> group ids
	for i, grp := range groups {
		gid := g.uuid()
		groupIDByName[grp.name] = gid
		g.modGroups = append(g.modGroups, modGroup{id: gid, name: grp.name, selMin: grp.selMin, selMax: grp.selMax, srt: i})
		for _, ms := range grp.mods {
			mod := modifier{id: g.uuid(), groupID: gid, name: ms.name, priceDelta: ms.delta, isDefault: ms.def}
			g.modifiers = append(g.modifiers, mod)
			g.modsByGroup[gid] = append(g.modsByGroup[gid], mod)
		}
		for _, cat := range grp.forKinds {
			forCategories[cat] = append(forCategories[cat], gid)
		}
	}

	// Link each menu item to the groups configured for its category.
	catNameByID := map[string]string{}
	for _, c := range g.categories {
		catNameByID[c.id] = c.name
	}
	for _, m := range g.menuItems {
		catName := catNameByID[m.categoryID]
		for i, gid := range forCategories[catName] {
			g.itemMods = append(g.itemMods, itemModGroup{id: g.uuid(), itemID: m.id, groupID: gid, required: false, sort: i})
			g.groupsByItem[m.id] = append(g.groupsByItem[m.id], gid)
		}
	}
}

var firstNames = []string{
	"James", "Mary", "Robert", "Patricia", "John", "Jennifer", "Michael", "Linda",
	"David", "Elizabeth", "William", "Barbara", "Richard", "Susan", "Joseph", "Jessica",
	"Thomas", "Sarah", "Christopher", "Karen", "Daniel", "Nancy", "Matthew", "Emily",
	"Anthony", "Ashley", "Mark", "Amanda", "Donald", "Olivia", "Steven", "Emma",
	"Andrew", "Grace", "Joshua", "Chloe", "Kevin", "Sophia", "Brian", "Isabella",
	"Ethan", "Mia", "Ryan", "Ava", "Nathan", "Ella", "Tyler", "Zoe", "Aaron", "Lily",
}

var lastNames = []string{
	"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
	"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson",
	"Thomas", "Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson",
	"White", "Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson", "Walker",
	"Young", "Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores",
	"Green", "Adams", "Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell", "Carter",
}

func (g *gen) buildCustomers() {
	seenEmail := map[string]bool{}
	seenPhone := map[string]bool{}
	for i := 0; i < 60; i++ {
		first := pick(g, firstNames)
		last := pick(g, lastNames)
		base := strings.ToLower(first + "." + last)
		email := base + "@example.com"
		for n := 2; seenEmail[email]; n++ {
			email = fmt.Sprintf("%s%d@example.com", base, n)
		}
		seenEmail[email] = true

		cid := g.uuid()
		// One or two addresses; the first is the default.
		nAddr := 1
		if g.chance(0.35) {
			nAddr = 2
		}
		var addrs []address
		for a := 0; a < nAddr; a++ {
			m := metros[g.intn(len(metros))]
			label := "Home"
			instr := "Leave at the front door."
			if a == 1 {
				label = "Work"
				instr = "Front desk / reception."
			}
			ad := address{
				id: g.uuid(), customerID: cid, label: label,
				line1:  fmt.Sprintf("%d %s", 100+g.intn(4800), pick(g, m.streets)),
				area:   pick(g, []string{"Downtown", "Midtown", "Uptown", "Riverside", "The Heights"}),
				city:   m.city,
				postal: m.zip,
				lat:    m.lat + (g.rng.Float64()-0.5)*0.05,
				lng:    m.lng + (g.rng.Float64()-0.5)*0.05,
				instructions: instr, isDefault: a == 0, metroIdx: 0,
			}
			if a == 1 {
				ad.line2 = fmt.Sprintf("Suite %d", 100+g.intn(400))
			}
			addrs = append(addrs, ad)
		}
		g.addresses = append(g.addresses, addrs...)
		g.addrByCust[cid] = addrs

		phone := fmt.Sprintf("+1 (%s) 555-%04d", metros[g.intn(len(metros))].area, 1000+i)
		for seenPhone[phone] {
			phone = fmt.Sprintf("+1 (%s) 555-%04d", metros[g.intn(len(metros))].area, 1000+g.intn(9000))
		}
		seenPhone[phone] = true

		status := "active"
		if g.chance(0.06) {
			status = "inactive"
		}
		g.customers = append(g.customers, customer{
			id: cid, fullName: first + " " + last, email: email, phone: phone,
			defaultAddressID: addrs[0].id, status: status,
			marketingOptIn: g.chance(0.4),
			lastLoginAge:   time.Duration(g.intn(30*24)) * time.Hour,
			createdAge:     time.Duration(60+g.intn(340)) * 24 * time.Hour,
		})

		// One loyalty account per customer.
		lifetime := g.intn(4000)
		tier := "bronze"
		switch {
		case lifetime > 3000:
			tier = "gold"
		case lifetime > 1200:
			tier = "silver"
		}
		acct := loyaltyAccount{id: g.uuid(), customerID: cid, tier: tier, balance: g.intn(600), lifetime: lifetime}
		g.loyalty = append(g.loyalty, acct)
		g.loyaltyByCust[cid] = &g.loyalty[len(g.loyalty)-1]
	}
}

func (g *gen) buildDrivers() {
	vehicles := []string{"car", "bike", "scooter"}
	seenPhone := map[string]bool{}
	for i := 0; i < 8; i++ {
		first := pick(g, firstNames)
		last := pick(g, lastNames)
		m := metros[g.intn(len(metros))]
		phone := fmt.Sprintf("+1 (%s) 555-%04d", m.area, 2000+i)
		for seenPhone[phone] {
			phone = fmt.Sprintf("+1 (%s) 555-%04d", m.area, 2000+g.intn(9000))
		}
		seenPhone[phone] = true
		status := pick(g, []string{"available", "available", "busy", "offline"})
		g.drivers = append(g.drivers, driver{
			id: g.uuid(), fullName: first + " " + last, phone: phone,
			email:   strings.ToLower(first+"."+last) + "@drivers.example.com",
			vehicle: pick(g, vehicles),
			plate:   fmt.Sprintf("%s-%04d", m.state, 1000+g.intn(9000)),
			status:  status,
			lat:     m.lat + (g.rng.Float64()-0.5)*0.05,
			lng:     m.lng + (g.rng.Float64()-0.5)*0.05,
			rating:  math.Round((4.0+g.rng.Float64()*0.9)*100) / 100,
			totalDeliveries: 80 + g.intn(220),
			hiredDaysAgo:    60 + g.intn(440),
			homeLocationID:  g.locations[g.intn(len(g.locations))].id,
		})
	}
}

func (g *gen) buildPromos() {
	g.promos = []promo{
		{id: g.uuid(), code: "WELCOME5", description: "$5 off your first order", discountType: "flat", discountValue: 5, minOrder: 20},
		{id: g.uuid(), code: "SPRING10", description: "10% off orders over $25", discountType: "percent", discountValue: 10, minOrder: 25, maxDiscount: 8, hasMaxDiscount: true},
		{id: g.uuid(), code: "FREESHIP", description: "Free delivery on orders over $30", discountType: "free_delivery", discountValue: 0, minOrder: 30},
		{id: g.uuid(), code: "SUMMER15", description: "15% off orders over $40", discountType: "percent", discountValue: 15, minOrder: 40, maxDiscount: 12, hasMaxDiscount: true},
		{id: g.uuid(), code: "LOYAL5", description: "$5 loyalty reward", discountType: "flat", discountValue: 5, minOrder: 15},
		{id: g.uuid(), code: "TREAT20", description: "20% off orders over $50", discountType: "percent", discountValue: 20, minOrder: 50, maxDiscount: 15, hasMaxDiscount: true},
	}
}

// ---- order generation -------------------------------------------------------

// weekFactor shapes a repeating weekly pattern (busier toward the weekend). The
// index is dayIndex%7, giving weekly seasonality independent of the load date.
var weekFactor = []float64{0.9, 1.0, 1.0, 1.05, 1.15, 1.35, 1.3}

// hourWeights shapes intra-day clustering: a morning rush, a lunch bump, and an
// afternoon lull-then-peak. Index is the "hour bucket" within a day.
var hourWeights = []float64{
	0.2, 0.1, 0.1, 0.1, 0.2, 0.6, 2.0, 4.5, 6.0, 5.0, 3.5, 4.0,
	5.0, 4.0, 3.5, 4.0, 3.0, 2.5, 2.0, 1.5, 1.0, 0.7, 0.4, 0.3,
}

func (g *gen) buildOrders() {
	const days = 180
	spikeDays := map[int]bool{20: true, 45: true, 90: true, 120: true, 150: true}
	for age := days; age >= 0; age-- {
		elapsed := float64(days-age) / float64(days) // 0 (oldest) .. 1 (today)
		trend := 0.7 + 0.6*elapsed
		expected := 9.0 * weekFactor[age%7] * trend
		if spikeDays[days-age] {
			expected *= 1.6
		}
		expected *= 0.8 + 0.4*g.rng.Float64() // noise
		count := int(math.Round(expected))
		for i := 0; i < count; i++ {
			g.createOrder(age)
		}
	}
	g.buildCarts()
}

func (g *gen) createOrder(dayAge int) {
	cust := g.customers[g.intn(len(g.customers))]
	loc := g.locations[g.intn(len(g.locations))]
	addrs := g.addrByCust[cust.id]
	addr := addrs[0]

	hour := g.weightedIndex(hourWeights)
	within := time.Duration(hour)*time.Hour + time.Duration(g.intn(3600))*time.Second
	placedAge := time.Duration(dayAge)*24*time.Hour + within

	// Status: fresh orders (last ~1.5 days) may still be in flight.
	status := "delivered"
	delivered := true
	switch {
	case placedAge < 36*time.Hour && g.chance(0.6):
		status = pick(g, []string{"placed", "confirmed", "preparing", "dispatched"})
		delivered = false
	case g.chance(0.05):
		status = "cancelled"
		delivered = false
	}

	orderID := g.uuid()
	g.orderSeq++
	number := fmt.Sprintf("CK-%06d", 100000+g.orderSeq)

	// Line items.
	nLines := 1 + g.weightedIndex([]float64{45, 30, 18, 7})
	subtotal := 0.0
	var firstItemID string
	for l := 0; l < nLines; l++ {
		item := g.menuItems[g.intn(len(g.menuItems))]
		if firstItemID == "" {
			firstItemID = item.id
		}
		variants := g.variantsByItem[item.id]
		v := variants[g.weightedIndexForVariants(variants)]
		qty := 1 + g.weightedIndex([]float64{70, 22, 8})
		unit := item.basePrice + v.priceDelta

		oiID := g.uuid()
		var chosenMods []modifier
		for _, gid := range g.groupsByItem[item.id] {
			chosenMods = append(chosenMods, g.pickModifiers(gid)...)
		}
		for _, mod := range chosenMods {
			unit += mod.priceDelta
			g.oiMods = append(g.oiMods, oiModifier{
				id: g.uuid(), orderItemID: oiID, modifierID: mod.id,
				snapshotName: mod.name, snapshotPriceDelta: mod.priceDelta,
			})
		}
		unit = round2(unit)
		lineTotal := round2(unit * float64(qty))
		subtotal += lineTotal
		g.orderItems = append(g.orderItems, orderItem{
			id: oiID, orderID: orderID, menuItemID: item.id, variantID: v.id,
			quantity: qty, unitPrice: unit, lineTotal: lineTotal, snapshotName: item.name,
		})
	}
	subtotal = round2(subtotal)

	// Fees, discount, tax.
	deliveryFee := 0.0
	if g.chance(0.7) {
		deliveryFee = pick(g, []float64{2.99, 3.49, 3.99, 4.49})
	}
	discount, promoID := 0.0, ""
	if g.chance(0.15) {
		p := g.promos[g.intn(len(g.promos))]
		if subtotal >= p.minOrder {
			switch p.discountType {
			case "flat":
				discount = p.discountValue
			case "percent":
				discount = subtotal * p.discountValue / 100
				if p.hasMaxDiscount && discount > p.maxDiscount {
					discount = p.maxDiscount
				}
			case "free_delivery":
				discount = deliveryFee
			}
			discount = round2(discount)
			promoID = p.id
		}
	}
	tax := round2(subtotal * metros[loc.metroIdx].taxRate)
	grand := round2(subtotal + deliveryFee - discount + tax)

	method := pick3(g, []string{"card", "apple_pay", "google_pay", "cash"}, []float64{55, 20, 10, 15})

	g.orders = append(g.orders, order{
		id: orderID, customerID: cust.id, addressID: addr.id, number: number,
		status: status, paymentMethod: method, locationID: loc.id,
		placedAge: placedAge, subtotal: subtotal, deliveryFee: deliveryFee,
		discount: discount, tax: tax, grandTotal: grand, promoID: promoID,
		delivered: delivered,
	})

	g.buildOrderChildren(orderID, cust, method, placedAge, grand, status, delivered, discount, promoID, firstItemID)
}

func (g *gen) buildOrderChildren(orderID string, cust customer, method string, placedAge time.Duration, grand float64, status string, delivered bool, discount float64, promoID, firstItemID string) {
	// Payment.
	payStatus, captured := "pending", false
	capturedAge := placedAge
	switch {
	case delivered:
		payStatus, captured = "captured", true
		capturedAge = placedAge - time.Duration(g.intn(120))*time.Second
	case status == "cancelled":
		payStatus = "refunded"
	default:
		payStatus = "authorized"
	}
	ref := ""
	if method != "cash" {
		ref = g.hexRef()
	}
	g.payments = append(g.payments, payment{
		id: g.uuid(), orderID: orderID, customerID: cust.id, method: method,
		providerRef: ref, amount: grand, status: payStatus,
		capturedAge: capturedAge, captured: captured,
	})

	// Delivery (delivered or dispatched orders).
	if delivered || status == "dispatched" {
		drv := g.drivers[g.intn(len(g.drivers))]
		assignedAge := placedAge - 5*time.Minute
		delAge := placedAge - time.Duration(30+g.intn(30))*time.Minute
		dstatus := "in_transit"
		if delivered {
			dstatus = "delivered"
		} else {
			delAge = -1 // not yet delivered
		}
		g.deliveries = append(g.deliveries, delivery{
			id: g.uuid(), orderID: orderID, driverID: drv.id, status: dstatus,
			assignedAge: assignedAge, deliveredAge: delAge,
			distanceKM: round2(0.8 + g.rng.Float64()*5.4),
		})
	}

	// Review (some delivered orders).
	if delivered && g.chance(0.45) {
		rating := 5 - g.weightedIndex([]float64{50, 30, 12, 5, 3})
		title, body := g.reviewText(rating)
		rv := review{
			id: g.uuid(), orderID: orderID, customerID: cust.id, rating: rating,
			title: title, body: body,
			createdAge: placedAge - time.Duration(24+g.intn(24))*time.Hour,
		}
		if g.chance(0.6) {
			rv.menuItemID = firstItemID
		}
		if g.chance(0.4) && len(g.deliveries) > 0 {
			rv.driverID = g.deliveries[len(g.deliveries)-1].driverID
		}
		g.reviews = append(g.reviews, rv)
	}

	// Order promotion.
	if promoID != "" && discount > 0 {
		g.orderPromo = append(g.orderPromo, orderPromotion{
			id: g.uuid(), orderID: orderID, promoID: promoID, customerID: cust.id,
			discount: discount, appliedAge: placedAge,
		})
	}

	// Loyalty earn on delivered orders.
	if delivered {
		if acct := g.loyaltyByCust[cust.id]; acct != nil {
			pts := int(grand)
			acct.balance += pts
			g.loyaltyTxn = append(g.loyaltyTxn, loyaltyTxn{
				id: g.uuid(), accountID: acct.id, customerID: cust.id, orderID: orderID,
				kind: "earn", points: pts, description: "Points earned on order",
				balanceAfter: acct.balance, createdAge: placedAge,
			})
		}
	}

	// Notification.
	kind, title := "order_confirmed", "Order confirmed"
	if delivered {
		kind, title = "order_delivered", "Your order was delivered"
	}
	g.notifs = append(g.notifs, notification{
		id: g.uuid(), customerID: cust.id, channel: pick(g, []string{"push", "email", "sms"}),
		kind: kind, title: title, body: "Thanks for ordering from Copper Kettle Coffee & Bakery.",
		createdAge: placedAge, read: g.chance(0.7),
	})
}

// buildCarts creates a handful of recent, still-open carts.
func (g *gen) buildCarts() {
	for i := 0; i < 12; i++ {
		cust := g.customers[g.intn(len(g.customers))]
		loc := g.locations[g.intn(len(g.locations))]
		cartID := g.uuid()
		subtotal := 0.0
		n := 1 + g.intn(3)
		for l := 0; l < n; l++ {
			item := g.menuItems[g.intn(len(g.menuItems))]
			v := g.variantsByItem[item.id][0]
			qty := 1 + g.intn(2)
			line := round2((item.basePrice + v.priceDelta) * float64(qty))
			subtotal += line
			g.cartItems = append(g.cartItems, cartItem{
				id: g.uuid(), cartID: cartID, menuItemID: item.id, variantID: v.id,
				quantity: qty, lineTotal: line,
			})
		}
		g.carts = append(g.carts, cart{
			id: cartID, customerID: cust.id, locationID: loc.id,
			subtotal: round2(subtotal), createdAge: time.Duration(g.intn(48)) * time.Hour,
		})
	}
}

func (g *gen) weightedIndexForVariants(vs []variant) int {
	if len(vs) == 1 {
		return 0
	}
	// Favor the default (Medium) size.
	return g.weightedIndex([]float64{25, 50, 25})
}

func (g *gen) pickModifiers(groupID string) []modifier {
	mods := g.modsByGroup[groupID]
	if len(mods) == 0 {
		return nil
	}
	var out []modifier
	// Group behavior keyed off the group's name via its first modifier.
	switch {
	case hasMod(mods, "Oat Milk"): // Milk: sometimes swap to a non-default milk
		if g.chance(0.45) {
			nonDefault := filterNonDefault(mods)
			out = append(out, nonDefault[g.intn(len(nonDefault))])
		}
	case hasMod(mods, "Add Espresso Shot"):
		if g.chance(0.25) {
			out = append(out, mods[0])
		}
	case hasMod(mods, "Vanilla Syrup"):
		if g.chance(0.30) {
			out = append(out, mods[g.intn(len(mods))])
		}
	case hasMod(mods, "Butter"):
		if g.chance(0.35) {
			out = append(out, mods[g.intn(len(mods))])
		}
	case hasMod(mods, "Toasted"):
		if g.chance(0.6) {
			out = append(out, mods[0]) // Toasted
		}
	}
	return out
}

func hasMod(mods []modifier, name string) bool {
	for _, m := range mods {
		if m.name == name {
			return true
		}
	}
	return false
}

func filterNonDefault(mods []modifier) []modifier {
	var out []modifier
	for _, m := range mods {
		if !m.isDefault {
			out = append(out, m)
		}
	}
	if len(out) == 0 {
		return mods
	}
	return out
}

var reviewTitles = map[int][]string{
	5: {"Fantastic!", "Best coffee in town", "Absolutely loved it", "Perfect as always"},
	4: {"Really good", "Solid order", "Happy with it", "Would order again"},
	3: {"It was okay", "Decent", "Middling", "Not bad"},
	2: {"Disappointing", "Could be better", "Not great"},
	1: {"Very disappointed", "Won't order again", "Poor experience"},
}
var reviewBodies = map[int][]string{
	5: {"Everything was fresh and arrived hot. Highly recommend.", "Great flavor and quick delivery.", "The pastries were amazing."},
	4: {"Good food, delivery was on time.", "Tasty and well packaged.", "Enjoyed it overall."},
	3: {"Order was fine but nothing special.", "A little cold on arrival.", "Average experience."},
	2: {"Order took a while and was lukewarm.", "Not quite what I expected."},
	1: {"Missing an item and the coffee was cold.", "Long wait and poor quality."},
}

func (g *gen) reviewText(rating int) (string, string) {
	t := reviewTitles[rating]
	b := reviewBodies[rating]
	return t[g.intn(len(t))], b[g.intn(len(b))]
}

func (g *gen) hexRef() string {
	var u [8]byte
	g.rng.Read(u[:])
	return "ch_" + fmt.Sprintf("%x", u)
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

// pick3 chooses a string by weight.
func pick3(g *gen, xs []string, weights []float64) string {
	return xs[g.weightedIndex(weights)]
}

// ---- SQL emission -----------------------------------------------------------

func geo(v float64) string   { return strconv.FormatFloat(v, 'f', 6, 64) }
func itoa(v int) string      { return strconv.Itoa(v) }
func relDate(days int) string { return fmt.Sprintf("(now() - interval '%d days')::date", days) }

const menuAge = 200 * 24 * time.Hour
const locationAge = 300 * 24 * time.Hour

// insertRows emits chunked multi-row INSERTs for readability and to keep single
// statements a sane size.
func insertRows(b *strings.Builder, table, cols string, rows []string) {
	if len(rows) == 0 {
		return
	}
	const chunk = 200
	for i := 0; i < len(rows); i += chunk {
		end := i + chunk
		if end > len(rows) {
			end = len(rows)
		}
		fmt.Fprintf(b, "INSERT INTO %s %s VALUES\n", table, cols)
		for j := i; j < end; j++ {
			sep := ","
			if j == end-1 {
				sep = ";"
			}
			b.WriteString("    ")
			b.WriteString(rows[j])
			b.WriteString(sep)
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}
}

func (g *gen) writeData(b *strings.Builder) {
	custByID := map[string]customer{}
	for _, c := range g.customers {
		custByID[c.id] = c
	}

	// allergens
	var rows []string
	for _, a := range g.allergens {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s)", s(a.id), s(a.name), s(a.slug), ns(a.emoji)))
	}
	insertRows(b, "public.allergens", "(id, name, slug, icon_emoji)", rows)

	// cafe_locations
	rows = nil
	for _, l := range g.locations {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s, %s, %s, %s, true, %s)",
			s(l.id), s(l.name), s(l.slug), s(l.address), geo(l.lat), geo(l.lng), s(l.phone),
			s(l.opensAt), s(l.closesAt), rel(locationAge)))
	}
	insertRows(b, "public.cafe_locations", "(id, name, slug, address, latitude, longitude, phone, opens_at, closes_at, is_active, created_at)", rows)

	// menu_categories (all tied to the first location, mirroring the original)
	rows = nil
	loc0 := g.locations[0].id
	for _, c := range g.categories {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, true, %s, %s)",
			s(c.id), s(c.name), s(c.slug), itoa(c.sort), rel(locationAge), s(loc0)))
	}
	insertRows(b, "public.menu_categories", "(id, name, slug, sort_order, is_active, created_at, location_id)", rows)

	// menu_items
	rows = nil
	for _, m := range g.menuItems {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, true, %s, %s, %s)",
			s(m.id), s(m.categoryID), s(m.name), s(m.slug), s(m.description), money(m.basePrice),
			b2(m.veg), b2(m.vegan), b2(m.signature), itoa(m.calories), itoa(m.prepMinutes),
			s(m.imageURL), jsonbArray(m.tags), rel(menuAge), rel(menuAge)))
	}
	insertRows(b, "public.menu_items", "(id, category_id, name, slug, description, base_price, is_vegetarian, is_vegan, is_signature, calories, prep_time_minutes, image_url, is_available, tags, created_at, updated_at)", rows)

	// menu_item_variants
	rows = nil
	for _, v := range g.variants {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s)",
			s(v.id), s(v.itemID), s(v.name), money(v.priceDelta), b2(v.isDefault), itoa(v.sort)))
	}
	insertRows(b, "public.menu_item_variants", "(id, menu_item_id, name, price_delta, is_default, sort_order)", rows)

	// modifier_groups
	rows = nil
	for _, m := range g.modGroups {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s)",
			s(m.id), s(m.name), itoa(m.selMin), itoa(m.selMax), itoa(m.srt)))
	}
	insertRows(b, "public.modifier_groups", "(id, name, selection_min, selection_max, sort_order)", rows)

	// modifiers
	rows = nil
	for _, m := range g.modifiers {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s)",
			s(m.id), s(m.groupID), s(m.name), money(m.priceDelta), b2(m.isDefault)))
	}
	insertRows(b, "public.modifiers", "(id, modifier_group_id, name, price_delta, is_default)", rows)

	// menu_item_modifier_groups
	rows = nil
	for _, im := range g.itemMods {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s)",
			s(im.id), s(im.itemID), s(im.groupID), b2(im.required), itoa(im.sort)))
	}
	insertRows(b, "public.menu_item_modifier_groups", "(id, menu_item_id, modifier_group_id, is_required, sort_order)", rows)

	// menu_item_allergens
	rows = nil
	for _, ia := range g.itemAllerg {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s)", s(ia.id), s(ia.itemID), s(ia.allergenID)))
	}
	insertRows(b, "public.menu_item_allergens", "(id, menu_item_id, allergen_id)", rows)

	// customers
	rows = nil
	for _, c := range g.customers {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)",
			s(c.id), s(c.fullName), s(c.email), s(c.phone), s("$argon2id$v=19$placeholder"),
			s(c.defaultAddressID), b2(c.marketingOptIn), s(c.status), rel(c.lastLoginAge),
			rel(c.createdAge), rel(c.createdAge)))
	}
	insertRows(b, "public.customers", "(id, full_name, email, phone, password_hash, default_address_id, marketing_opt_in, status, last_login_at, created_at, updated_at)", rows)

	// customer_addresses
	rows = nil
	for _, a := range g.addresses {
		age := custByID[a.customerID].createdAge
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)",
			s(a.id), s(a.customerID), ns(a.label), s(a.line1), ns(a.line2), ns(a.area),
			s(a.city), s(a.postal), geo(a.lat), geo(a.lng), ns(a.instructions), b2(a.isDefault), rel(age)))
	}
	insertRows(b, "public.customer_addresses", "(id, customer_id, label, line1, line2, area, city, postal_code, latitude, longitude, instructions, is_default, created_at)", rows)

	// drivers
	rows = nil
	for _, d := range g.drivers {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)",
			s(d.id), s(d.fullName), s(d.phone), ns(d.email), s(d.vehicle), ns(d.plate), s(d.status),
			geo(d.lat), geo(d.lng), money(d.rating), itoa(d.totalDeliveries), relDate(d.hiredDaysAgo),
			rel(locationAge), rel(locationAge), s(d.homeLocationID)))
	}
	insertRows(b, "public.drivers", "(id, full_name, phone, email, vehicle_type, license_plate, status, current_latitude, current_longitude, rating_average, total_deliveries, hired_at, created_at, updated_at, home_location_id)", rows)

	// loyalty_accounts
	rows = nil
	for _, la := range g.loyalty {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, NULL, %s, %s)",
			s(la.id), s(la.customerID), itoa(la.balance), itoa(la.lifetime), s(la.tier),
			rel(locationAge), rel(0)))
	}
	insertRows(b, "public.loyalty_accounts", "(id, customer_id, balance_points, lifetime_points, tier, tier_expires_at, created_at, updated_at)", rows)

	// promo_codes
	rows = nil
	for _, p := range g.promos {
		maxD := "NULL"
		if p.hasMaxDiscount {
			maxD = money(p.maxDiscount)
		}
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s, %s, %s, %s, 1000, 0, true, %s)",
			s(p.id), s(p.code), s(p.description), s(p.discountType), money(p.discountValue),
			money(p.minOrder), maxD, rel(locationAge), relFuture(90), rel(locationAge)))
	}
	insertRows(b, "public.promo_codes", "(id, code, description, discount_type, discount_value, min_order_amount, max_discount, valid_from, valid_until, usage_limit, used_count, is_active, created_at)", rows)

	// orders
	rows = nil
	for _, o := range g.orders {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, NULL, NULL, NULL, %s, %s, %s)",
			s(o.id), s(o.customerID), s(o.addressID), s(o.number), s(o.status), rel(o.placedAge),
			money(o.subtotal), money(o.deliveryFee), money(o.discount), money(o.tax), money(o.grandTotal),
			s(o.paymentMethod), rel(o.placedAge), rel(o.placedAge), s(o.locationID)))
	}
	insertRows(b, "public.orders", "(id, customer_id, address_id, order_number, status, placed_at, subtotal, delivery_fee, discount_total, tax, grand_total, payment_method, scheduled_for, customer_notes, internal_notes, created_at, updated_at, location_id)", rows)

	// order_items
	rows = nil
	for _, oi := range g.orderItems {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s, %s, %s, %s)",
			s(oi.id), s(oi.orderID), s(oi.menuItemID), s(oi.variantID), itoa(oi.quantity),
			money(oi.unitPrice), money(oi.lineTotal), s(oi.snapshotName), rel(0)))
	}
	insertRows(b, "public.order_items", "(id, order_id, menu_item_id, variant_id, quantity, unit_price, line_total, snapshot_name, created_at)", rows)

	// order_item_modifiers
	rows = nil
	for _, om := range g.oiMods {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s)",
			s(om.id), s(om.orderItemID), s(om.modifierID), s(om.snapshotName), money(om.snapshotPriceDelta)))
	}
	insertRows(b, "public.order_item_modifiers", "(id, order_item_id, modifier_id, snapshot_name, snapshot_price_delta)", rows)

	// order_promotions
	rows = nil
	for _, op := range g.orderPromo {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s)",
			s(op.id), s(op.orderID), s(op.promoID), s(op.customerID), money(op.discount), rel(op.appliedAge)))
	}
	insertRows(b, "public.order_promotions", "(id, order_id, promo_code_id, customer_id, discount_amount, applied_at)", rows)

	// payments
	rows = nil
	for _, p := range g.payments {
		capturedAt := "NULL"
		refundedAt := "NULL"
		switch p.status {
		case "captured":
			capturedAt = rel(p.capturedAge)
		case "refunded":
			refundedAt = rel(p.capturedAge)
		}
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)",
			s(p.id), s(p.orderID), s(p.customerID), s(p.method), ns(p.providerRef),
			money(p.amount), s(p.status), capturedAt, refundedAt, rel(p.capturedAge)))
	}
	insertRows(b, "public.payments", "(id, order_id, customer_id, method, provider_reference, amount, status, captured_at, refunded_at, created_at)", rows)

	// deliveries
	rows = nil
	for _, d := range g.deliveries {
		deliveredAt := "NULL"
		pickedUp := "NULL"
		proof := "NULL"
		sig := "NULL"
		if d.status == "delivered" {
			deliveredAt = rel(d.deliveredAge)
			pickedUp = rel(d.assignedAge - 8*time.Minute)
			proof = s("https://cdn.example.com/pod/" + d.id + ".jpg")
			sig = s("https://cdn.example.com/sig/" + d.id + ".png")
		} else {
			pickedUp = rel(d.assignedAge - 8*time.Minute)
		}
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s, %s, %s, '[]'::jsonb, %s, %s, %s, %s, %s)",
			s(d.id), s(d.orderID), s(d.driverID), rel(d.assignedAge), pickedUp, deliveredAt,
			rel(d.assignedAge-35*time.Minute), money(d.distanceKM), proof, sig, s(d.status),
			rel(d.assignedAge), rel(d.assignedAge)))
	}
	insertRows(b, "public.deliveries", "(id, order_id, driver_id, assigned_at, picked_up_at, delivered_at, expected_delivery_at, distance_km, delivery_route, proof_of_delivery_url, customer_signature_url, status, created_at, updated_at)", rows)

	// reviews
	rows = nil
	for _, r := range g.reviews {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s, %s, %s, '[]'::jsonb, true, NULL, NULL, %s)",
			s(r.id), s(r.orderID), s(r.customerID), ns(r.menuItemID), ns(r.driverID),
			itoa(r.rating), s(r.title), s(r.body), rel(r.createdAge)))
	}
	insertRows(b, "public.reviews", "(id, order_id, customer_id, menu_item_id, driver_id, rating, title, body, photos, is_published, response, responded_at, created_at)", rows)

	// loyalty_transactions
	rows = nil
	for _, lt := range g.loyaltyTxn {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s, %s, %s, %s)",
			s(lt.id), s(lt.accountID), s(lt.customerID), s(lt.orderID), s(lt.kind),
			itoa(lt.points), s(lt.description), itoa(lt.balanceAfter), rel(lt.createdAge)))
	}
	insertRows(b, "public.loyalty_transactions", "(id, account_id, customer_id, order_id, kind, points, description, balance_after, created_at)", rows)

	// notifications
	rows = nil
	for _, n := range g.notifs {
		readAt := "NULL"
		if n.read {
			readAt = rel(n.createdAge - 30*time.Minute)
		}
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s, '{}'::jsonb, %s, %s, %s)",
			s(n.id), s(n.customerID), s(n.channel), s(n.kind), s(n.title),
			s(n.body), readAt, rel(n.createdAge), rel(n.createdAge)))
	}
	insertRows(b, "public.notifications", "(id, customer_id, channel, kind, title, body, payload, read_at, sent_at, created_at)", rows)

	// carts
	rows = nil
	for _, c := range g.carts {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, NULL, %s, %s, %s)",
			s(c.id), s(c.customerID), money(c.subtotal), rel(c.createdAge), rel(c.createdAge), s(c.locationID)))
	}
	insertRows(b, "public.carts", "(id, customer_id, subtotal, expires_at, created_at, updated_at, location_id)", rows)

	// cart_items
	rows = nil
	for _, ci := range g.cartItems {
		rows = append(rows, fmt.Sprintf("(%s, %s, %s, %s, %s, %s, NULL, %s)",
			s(ci.id), s(ci.cartID), s(ci.menuItemID), s(ci.variantID), itoa(ci.quantity),
			money(ci.lineTotal), rel(0)))
	}
	insertRows(b, "public.cart_items", "(id, cart_id, menu_item_id, variant_id, quantity, line_total, notes, created_at)", rows)
}

// relFuture renders a future timestamp: now() + interval 'N days'.
func relFuture(days int) string {
	return fmt.Sprintf("now() + interval '%d days'", days)
}

// schemaSQL is the pg_dump preamble + CREATE TABLE block, carried verbatim
// from the original dump.
const schemaSQL = `SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
--



--
--



SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: allergens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.allergens (
    id uuid NOT NULL,
    name character varying(64) NOT NULL,
    slug character varying(64) NOT NULL,
    icon_emoji character varying(8)
);


--
-- Name: cafe_locations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cafe_locations (
    id uuid NOT NULL,
    name character varying(160) NOT NULL,
    slug character varying(64) NOT NULL,
    address character varying(255) NOT NULL,
    latitude numeric(9,6) NOT NULL,
    longitude numeric(9,6) NOT NULL,
    phone character varying(32),
    opens_at time without time zone,
    closes_at time without time zone,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: cart_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cart_items (
    id uuid NOT NULL,
    cart_id uuid NOT NULL,
    menu_item_id uuid NOT NULL,
    variant_id uuid,
    quantity integer DEFAULT 1 NOT NULL,
    line_total numeric(10,2) NOT NULL,
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: carts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.carts (
    id uuid NOT NULL,
    customer_id uuid NOT NULL,
    subtotal numeric(10,2) DEFAULT 0 NOT NULL,
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    location_id uuid
);


--
-- Name: customer_addresses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.customer_addresses (
    id uuid NOT NULL,
    customer_id uuid NOT NULL,
    label character varying(64),
    line1 character varying(255) NOT NULL,
    line2 character varying(255),
    area character varying(128),
    city character varying(128) NOT NULL,
    postal_code character varying(16),
    latitude numeric(9,6),
    longitude numeric(9,6),
    instructions text,
    is_default boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: customers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.customers (
    id uuid NOT NULL,
    full_name character varying(160) NOT NULL,
    email character varying(255) NOT NULL,
    phone character varying(32),
    password_hash character varying(255) NOT NULL,
    default_address_id uuid,
    marketing_opt_in boolean DEFAULT false NOT NULL,
    status character varying(32) DEFAULT 'active'::character varying NOT NULL,
    last_login_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: deliveries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.deliveries (
    id uuid NOT NULL,
    order_id uuid NOT NULL,
    driver_id uuid NOT NULL,
    assigned_at timestamp with time zone DEFAULT now() NOT NULL,
    picked_up_at timestamp with time zone,
    delivered_at timestamp with time zone,
    expected_delivery_at timestamp with time zone,
    distance_km numeric(6,2),
    delivery_route jsonb DEFAULT '[]'::jsonb NOT NULL,
    proof_of_delivery_url character varying(512),
    customer_signature_url character varying(512),
    status character varying(32) DEFAULT 'assigned'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: drivers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.drivers (
    id uuid NOT NULL,
    full_name character varying(160) NOT NULL,
    phone character varying(32) NOT NULL,
    email character varying(255),
    vehicle_type character varying(32) NOT NULL,
    license_plate character varying(32),
    status character varying(32) DEFAULT 'offline'::character varying NOT NULL,
    current_latitude numeric(9,6),
    current_longitude numeric(9,6),
    rating_average numeric(3,2) DEFAULT 0 NOT NULL,
    total_deliveries integer DEFAULT 0 NOT NULL,
    hired_at date,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    home_location_id uuid
);


--
-- Name: loyalty_accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.loyalty_accounts (
    id uuid NOT NULL,
    customer_id uuid NOT NULL,
    balance_points integer DEFAULT 0 NOT NULL,
    lifetime_points integer DEFAULT 0 NOT NULL,
    tier character varying(32) DEFAULT 'bronze'::character varying NOT NULL,
    tier_expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: loyalty_transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.loyalty_transactions (
    id uuid NOT NULL,
    account_id uuid NOT NULL,
    customer_id uuid NOT NULL,
    order_id uuid,
    kind character varying(32) NOT NULL,
    points integer NOT NULL,
    description character varying(255),
    balance_after integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: menu_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.menu_categories (
    id uuid NOT NULL,
    name character varying(128) NOT NULL,
    slug character varying(64) NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    location_id uuid
);


--
-- Name: menu_item_allergens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.menu_item_allergens (
    id uuid NOT NULL,
    menu_item_id uuid NOT NULL,
    allergen_id uuid NOT NULL
);


--
-- Name: menu_item_modifier_groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.menu_item_modifier_groups (
    id uuid NOT NULL,
    menu_item_id uuid NOT NULL,
    modifier_group_id uuid NOT NULL,
    is_required boolean DEFAULT false NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL
);


--
-- Name: menu_item_variants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.menu_item_variants (
    id uuid NOT NULL,
    menu_item_id uuid NOT NULL,
    name character varying(64) NOT NULL,
    price_delta numeric(10,2) DEFAULT 0 NOT NULL,
    is_default boolean DEFAULT false NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL
);


--
-- Name: menu_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.menu_items (
    id uuid NOT NULL,
    category_id uuid NOT NULL,
    name character varying(160) NOT NULL,
    slug character varying(96) NOT NULL,
    description text,
    base_price numeric(10,2) NOT NULL,
    is_vegetarian boolean DEFAULT false NOT NULL,
    is_vegan boolean DEFAULT false NOT NULL,
    is_signature boolean DEFAULT false NOT NULL,
    calories integer,
    prep_time_minutes integer,
    image_url character varying(512),
    is_available boolean DEFAULT true NOT NULL,
    tags jsonb DEFAULT '[]'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: modifier_groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.modifier_groups (
    id uuid NOT NULL,
    name character varying(128) NOT NULL,
    selection_min integer DEFAULT 0 NOT NULL,
    selection_max integer DEFAULT 1 NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL
);


--
-- Name: modifiers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.modifiers (
    id uuid NOT NULL,
    modifier_group_id uuid NOT NULL,
    name character varying(128) NOT NULL,
    price_delta numeric(10,2) DEFAULT 0 NOT NULL,
    is_default boolean DEFAULT false NOT NULL
);


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id uuid NOT NULL,
    customer_id uuid NOT NULL,
    channel character varying(16) NOT NULL,
    kind character varying(32) NOT NULL,
    title character varying(255) NOT NULL,
    body text,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    read_at timestamp with time zone,
    sent_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: order_item_modifiers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_item_modifiers (
    id uuid NOT NULL,
    order_item_id uuid NOT NULL,
    modifier_id uuid NOT NULL,
    snapshot_name character varying(128) NOT NULL,
    snapshot_price_delta numeric(10,2) NOT NULL
);


--
-- Name: order_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_items (
    id uuid NOT NULL,
    order_id uuid NOT NULL,
    menu_item_id uuid NOT NULL,
    variant_id uuid,
    quantity integer NOT NULL,
    unit_price numeric(10,2) NOT NULL,
    line_total numeric(10,2) NOT NULL,
    snapshot_name character varying(160) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: order_promotions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_promotions (
    id uuid NOT NULL,
    order_id uuid NOT NULL,
    promo_code_id uuid NOT NULL,
    customer_id uuid NOT NULL,
    discount_amount numeric(10,2) NOT NULL,
    applied_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.orders (
    id uuid NOT NULL,
    customer_id uuid NOT NULL,
    address_id uuid NOT NULL,
    order_number character varying(32) NOT NULL,
    status character varying(32) DEFAULT 'placed'::character varying NOT NULL,
    placed_at timestamp with time zone DEFAULT now() NOT NULL,
    subtotal numeric(10,2) NOT NULL,
    delivery_fee numeric(10,2) DEFAULT 0 NOT NULL,
    discount_total numeric(10,2) DEFAULT 0 NOT NULL,
    tax numeric(10,2) DEFAULT 0 NOT NULL,
    grand_total numeric(10,2) NOT NULL,
    payment_method character varying(32) NOT NULL,
    scheduled_for timestamp with time zone,
    customer_notes text,
    internal_notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    location_id uuid
);


--
-- Name: payments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payments (
    id uuid NOT NULL,
    order_id uuid NOT NULL,
    customer_id uuid NOT NULL,
    method character varying(32) NOT NULL,
    provider_reference character varying(128),
    amount numeric(10,2) NOT NULL,
    status character varying(32) DEFAULT 'pending'::character varying NOT NULL,
    captured_at timestamp with time zone,
    refunded_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: promo_codes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.promo_codes (
    id uuid NOT NULL,
    code character varying(64) NOT NULL,
    description character varying(255),
    discount_type character varying(32) NOT NULL,
    discount_value numeric(10,2) NOT NULL,
    min_order_amount numeric(10,2) DEFAULT 0 NOT NULL,
    max_discount numeric(10,2),
    valid_from timestamp with time zone NOT NULL,
    valid_until timestamp with time zone NOT NULL,
    usage_limit integer,
    used_count integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: reviews; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reviews (
    id uuid NOT NULL,
    order_id uuid NOT NULL,
    customer_id uuid NOT NULL,
    menu_item_id uuid,
    driver_id uuid,
    rating integer NOT NULL,
    title character varying(255),
    body text,
    photos jsonb DEFAULT '[]'::jsonb NOT NULL,
    is_published boolean DEFAULT true NOT NULL,
    response text,
    responded_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);
`

// constraintsSQL is the trailing PK/UNIQUE constraint block, carried verbatim.
const constraintsSQL = `--
-- Name: allergens allergens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.allergens
    ADD CONSTRAINT allergens_pkey PRIMARY KEY (id);


--
-- Name: allergens allergens_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.allergens
    ADD CONSTRAINT allergens_slug_key UNIQUE (slug);


--
-- Name: cafe_locations cafe_locations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cafe_locations
    ADD CONSTRAINT cafe_locations_pkey PRIMARY KEY (id);


--
-- Name: cafe_locations cafe_locations_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cafe_locations
    ADD CONSTRAINT cafe_locations_slug_key UNIQUE (slug);


--
-- Name: cart_items cart_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_pkey PRIMARY KEY (id);


--
-- Name: carts carts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.carts
    ADD CONSTRAINT carts_pkey PRIMARY KEY (id);


--
-- Name: customer_addresses customer_addresses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customer_addresses
    ADD CONSTRAINT customer_addresses_pkey PRIMARY KEY (id);


--
-- Name: customers customers_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_email_key UNIQUE (email);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);


--
-- Name: deliveries deliveries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.deliveries
    ADD CONSTRAINT deliveries_pkey PRIMARY KEY (id);


--
-- Name: drivers drivers_phone_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.drivers
    ADD CONSTRAINT drivers_phone_key UNIQUE (phone);


--
-- Name: drivers drivers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.drivers
    ADD CONSTRAINT drivers_pkey PRIMARY KEY (id);


--
-- Name: loyalty_accounts loyalty_accounts_customer_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.loyalty_accounts
    ADD CONSTRAINT loyalty_accounts_customer_id_key UNIQUE (customer_id);


--
-- Name: loyalty_accounts loyalty_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.loyalty_accounts
    ADD CONSTRAINT loyalty_accounts_pkey PRIMARY KEY (id);


--
-- Name: loyalty_transactions loyalty_transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.loyalty_transactions
    ADD CONSTRAINT loyalty_transactions_pkey PRIMARY KEY (id);


--
-- Name: menu_categories menu_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.menu_categories
    ADD CONSTRAINT menu_categories_pkey PRIMARY KEY (id);


--
-- Name: menu_item_allergens menu_item_allergens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.menu_item_allergens
    ADD CONSTRAINT menu_item_allergens_pkey PRIMARY KEY (id);


--
-- Name: menu_item_modifier_groups menu_item_modifier_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.menu_item_modifier_groups
    ADD CONSTRAINT menu_item_modifier_groups_pkey PRIMARY KEY (id);


--
-- Name: menu_item_variants menu_item_variants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.menu_item_variants
    ADD CONSTRAINT menu_item_variants_pkey PRIMARY KEY (id);


--
-- Name: menu_items menu_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.menu_items
    ADD CONSTRAINT menu_items_pkey PRIMARY KEY (id);


--
-- Name: modifier_groups modifier_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.modifier_groups
    ADD CONSTRAINT modifier_groups_pkey PRIMARY KEY (id);


--
-- Name: modifiers modifiers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.modifiers
    ADD CONSTRAINT modifiers_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: order_item_modifiers order_item_modifiers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_item_modifiers
    ADD CONSTRAINT order_item_modifiers_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: order_promotions order_promotions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_promotions
    ADD CONSTRAINT order_promotions_pkey PRIMARY KEY (id);


--
-- Name: orders orders_order_number_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_order_number_key UNIQUE (order_number);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: payments payments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_pkey PRIMARY KEY (id);


--
-- Name: promo_codes promo_codes_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promo_codes
    ADD CONSTRAINT promo_codes_code_key UNIQUE (code);


--
-- Name: promo_codes promo_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promo_codes
    ADD CONSTRAINT promo_codes_pkey PRIMARY KEY (id);


--
-- Name: reviews reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_pkey PRIMARY KEY (id);


--
-- PostgreSQL database dump complete
--
`
