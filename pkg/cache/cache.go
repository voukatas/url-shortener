package cache

type Cache interface {
	Get(string) (string, error)
	Set(string, string)
}

func NewCache(capacity int) Cache {
	if capacity < 1 {
		panic("capacity should be more than 1")
	}

	return NewLRUCache(capacity)

}
