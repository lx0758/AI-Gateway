package provider

import "fmt"

type ProviderError struct {
	StatusCode int
	Message    string
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("%d - %s", e.StatusCode, e.Message)
}

func IsRateLimitError(err error) bool {
	if providerErr, ok := err.(*ProviderError); ok {
		return providerErr.StatusCode == 429
	}
	return false
}

func GetProviderError(err error) *ProviderError {
	if providerErr, ok := err.(*ProviderError); ok {
		return providerErr
	}
	return nil
}
