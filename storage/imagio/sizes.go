package imagio

// Sizes ...
type Sizes []uint

// Has ...
func (z Sizes) Has(v uint) bool {
	for _, size := range z {
		if v == size || v/2 == size {
			return true
		}
	}
	return false
}
