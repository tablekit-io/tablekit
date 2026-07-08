package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
