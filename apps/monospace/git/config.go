package git

func ConfigGet(directory string, key string) (string, error) {
	return gitExecOutput("-C", directory, "config", key)
}

func ConfigSet(directory string, key string, value string) error {
	return gitExec("-C", directory, "config", key, value)
}

func GetHooksDir(directory string) (string, error) {
	return ConfigGet(directory, "core.hooksPath")
}

func SetHooksDir(directory string, hooksDir string) error {
	return ConfigSet(directory, "core.hooksPath", hooksDir)
}
