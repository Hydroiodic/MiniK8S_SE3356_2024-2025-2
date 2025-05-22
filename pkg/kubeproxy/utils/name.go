package utils

func FormatServiceName(namespace, name string) string {
	return namespace + "-" + name
}
