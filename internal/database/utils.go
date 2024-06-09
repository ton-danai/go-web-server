package database

func getMaxId[V Chirp | User](data map[int]V) int {
	maxKey := 0
	for key := range data {
		if key > maxKey {
			maxKey = key
		}
	}

	return maxKey
}
