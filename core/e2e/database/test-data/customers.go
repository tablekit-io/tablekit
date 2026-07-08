package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

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
