package service

func CreateUser(name string) string {
	if name == "" {
		return ""
	}
	return name
}
