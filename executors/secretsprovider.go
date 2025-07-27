package executors

type SecretsProvider interface {
	Init(config map[string]string) error
	GetCredentials() (map[string]any, error)
}
