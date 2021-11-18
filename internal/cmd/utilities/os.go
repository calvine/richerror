package utilities

import "os"

// exists returns whether the given file or directory exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// exists returns whether the given file or directory exists
func DirExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil && stat.IsDir() {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
