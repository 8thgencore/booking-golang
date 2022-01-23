package forms

type errors map[string][]string

// Add adds an error message for a given form field
func (e errors) Add(field, message string) {
	if message == "" {
		message = "Unknown error"
	}

	e[field] = append(e[field], message)
}

// Get returns the firts error message
func (e errors) Get(field string) string {
	fieldErrors := e[field]
	if len(fieldErrors) == 0 {
		return ""
	}

	return fieldErrors[0]
}
