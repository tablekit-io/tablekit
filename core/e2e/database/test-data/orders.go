package main

import (
	"fmt"
	"math"
	"time"
)

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
