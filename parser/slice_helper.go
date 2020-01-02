package main

// FindNewElements returns the elements in `b` that aren't in `a`.
func FindNewElements(a, b *[]Satellite) []Satellite {
	exists := make(map[string]Satellite, len(*a))
	for _, x := range *a {
		exists[x.GetName()] = x
	}

	var newItems []Satellite
	for _, x := range *b {
		if _, found := exists[x.GetName()]; !found {
			newItems = append(newItems, x)
		}
	}
	return newItems
}

// FindAbsent returns the elements in `a` that aren't in `b`.
func FindAbsent(a, b *[]Satellite) []Satellite {
	return FindNewElements(b, a)
}

// FindChanged returns pairs of elements that are changed between `a` and `b`. Name
func FindChanged(a, b *[]Satellite) [][]Satellite {
	exists := make(map[string]Satellite, len(*a))
	for _, x := range *a {
		exists[x.GetName()] = x
	}

	var changedItems [][]Satellite
	for _, x := range *b {
		if foundItem, found := exists[x.GetName()]; found {
			if foundItem != x {
				changedItems = append(changedItems, []Satellite{foundItem, x})
			}
		}
	}
	return changedItems
}
