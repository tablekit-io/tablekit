package main

import (
	"strings"
)

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
