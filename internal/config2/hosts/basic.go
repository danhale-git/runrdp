package hosts

import "github.com/spf13/viper"

func ParseBasic(v *viper.Viper) (map[string]interface{}, []interface{}, error) {
	key := "host.basic"
	if !v.IsSet(key) {
		return nil, nil, nil
	}
	raw := v.Get(key).(map[string]interface{})
	structs := make([]interface{}, len(raw))
	for i := range structs {
		structs[i] = &Basic{}
	}

	/*for _, v := range raw {
		// TODO: call validation function
	}*/

	return raw, structs, nil
}

// Basic defines a host to connect to using an IP or hostname.
//
// Users may configure a basic host with global fields only (see config.toml). No other fields are defined.
type Basic struct{}

// Socket returns this host's IP or hostname.
func (h *Basic) Socket() (string, string, error) {
	return "", "", nil
}
