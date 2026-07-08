package main

import (
	"time"
)

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
