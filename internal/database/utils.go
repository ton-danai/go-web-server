package database

func getMaxId[V Chirp | User](data map[int]V) int {
	maxKey := 0
	for key := range data {
		// v = append(v, value)
		if key > maxKey {
			maxKey = key
		}
	}

	return maxKey
}
