package main

import (
	"fmt"
)

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
