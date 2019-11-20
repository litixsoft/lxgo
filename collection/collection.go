package lxCollection

// FirstIndex,
// returns the first index of the target string, or -1 if no match is found.
// Example:
// fmt.Println(FirstIndex(strs, "pear"))
func FirstIndex(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

// Include,
// returns true if the target string t is in the slice.
// Example:
// fmt.Println(Include(strs, "grape"))
func Include(vs []string, t string) bool {
	return FirstIndex(vs, t) >= 0
}

// FindAny,
// returns true if one of the strings in the slice satisfies the predicate f.
// Example:
// fmt.Println(FindAny(strs, func(v string) bool {
//     return strings.HasPrefix(v, "p")
// }))
func FindAny(vs []string, f func(string) bool) bool {
	for _, v := range vs {
		if f(v) {
			return true
		}
	}
	return false
}

// FindAll,
// returns true if all of the strings in the slice satisfy the predicate f.
// Example:
// fmt.Println(FindAll(strs, func(v string) bool {
//     return strings.HasPrefix(v, "p")
// }))
func FindAll(vs []string, f func(string) bool) bool {
	for _, v := range vs {
		if !f(v) {
			return false
		}
	}
	return true
}

// Filter, returns a new slice containing all strings in the slice that satisfy the predicate f.
// // Example:
// fmt.Println(Filter(strs, func(v string) bool {
//     return strings.Contains(v, "e")
// }))
func Filter(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

// Map, returns a new slice,
// containing the results of applying the function f to each string in the original slice.
// Example:
// fmt.Println(Map(strs, strings.ToUpper))
func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}
