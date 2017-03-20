package golobo

import (
	"fmt"
)

func Show(servers []string) ([]string, error) {

	conn, _, err := initPath(servers)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	children, _, err := conn.Children(path)
	if err != nil {
		return nil, err
	}

	return children, nil
}

func Remove(servers []string, child string) error {

	conn, _, err := initPath(servers)
	if err != nil {
		return err
	}
	defer conn.Close()

	return conn.Delete(fmt.Sprint(path, "/", child), 0)
}
