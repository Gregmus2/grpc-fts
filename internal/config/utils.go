package config

type Metadata map[string]string

func (m Metadata) MergeWith(target map[string]string) map[string]string {
	result := make(map[string]string, len(m))
	for key, value := range m {
		result[key] = value
	}
	for key, value := range target {
		result[key] = value
	}

	return result
}
