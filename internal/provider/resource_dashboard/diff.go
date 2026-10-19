package resource_dashboard

// diffStringSets compares the currently-attached set against the desired
// set and returns which elements need to be added and which need to be
// removed to reconcile the two. Order of the inputs and outputs is not
// meaningful.
func diffStringSets(current, desired []string) (toAdd, toRemove []string) {
	currentSet := make(map[string]bool, len(current))
	for _, id := range current {
		currentSet[id] = true
	}

	desiredSet := make(map[string]bool, len(desired))
	for _, id := range desired {
		desiredSet[id] = true
	}

	for _, id := range desired {
		if !currentSet[id] {
			toAdd = append(toAdd, id)
		}
	}
	for _, id := range current {
		if !desiredSet[id] {
			toRemove = append(toRemove, id)
		}
	}

	return toAdd, toRemove
}
