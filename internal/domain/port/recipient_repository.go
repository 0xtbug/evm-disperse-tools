package port

// RecipientRepository defines the interface for recipient persistence
type RecipientRepository interface {
	Load() ([]string, error)
}
