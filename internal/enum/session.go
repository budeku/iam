package enum

type SessionAuthMethod uint8

const (
	SessionAuthMethodPassword SessionAuthMethod = 1 + iota
)

var sessionAuthMethodNames = map[SessionAuthMethod]string{
	SessionAuthMethodPassword: "password",
}

var sessionAuthMethodValues = map[string]SessionAuthMethod{
	"password": SessionAuthMethodPassword,
}

func (m SessionAuthMethod) Valid() bool    { return enumValid(sessionAuthMethodNames, m) }
func (m SessionAuthMethod) String() string { return enumString(sessionAuthMethodNames, m) }

func (m SessionAuthMethod) MarshalJSON() ([]byte, error) {
	return enumMarshalJSON(sessionAuthMethodNames, m)
}

func (m *SessionAuthMethod) UnmarshalJSON(data []byte) error {
	return enumUnmarshalJSON(sessionAuthMethodValues, data, m)
}

func (m SessionAuthMethod) MarshalText() ([]byte, error) {
	return enumMarshalText(sessionAuthMethodNames, m)
}

func (m *SessionAuthMethod) UnmarshalText(data []byte) error {
	return enumUnmarshalText(sessionAuthMethodValues, data, m)
}

func SessionAuthMethodOptions() []Option {
	return enumOptions(sessionAuthMethodNames)
}
