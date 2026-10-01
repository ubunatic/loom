package loom

import (
	"sort"
	"strings"
)

// caseInsensitivePathCollisions returns every pair of paths that differ only
// by case. Paths are compared as complete relative paths.
func caseInsensitivePathCollisions(paths []string) [][2]string {
	byFoldedPath := make(map[string][]string, len(paths))
	for _, path := range paths {
		folded := strings.ToLower(path)
		byFoldedPath[folded] = append(byFoldedPath[folded], path)
	}

	var collisions [][2]string
	for _, matches := range byFoldedPath {
		if len(matches) < 2 {
			continue
		}
		sort.Strings(matches)
		for i := range matches {
			for j := i + 1; j < len(matches); j++ {
				if matches[i] != matches[j] {
					collisions = append(collisions, [2]string{matches[i], matches[j]})
				}
			}
		}
	}
	sort.Slice(collisions, func(i, j int) bool {
		if collisions[i][0] != collisions[j][0] {
			return collisions[i][0] < collisions[j][0]
		}
		return collisions[i][1] < collisions[j][1]
	})
	return collisions
}
