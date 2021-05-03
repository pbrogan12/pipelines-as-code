package test

import "errors"

type KinterfaceTest struct {
	ConsoleURL     string
	NamespaceError bool

	prDescribe string
}

func (k *KinterfaceTest) GetConsoleUI(ns string, pr string) string {
	return k.ConsoleURL
}

func (k *KinterfaceTest) GetNamespace(ns string) error {
	if k.NamespaceError {
		return errors.New("Cannot find Namespace")
	}
	return nil
}

func (k *KinterfaceTest) TektonCliPRDescribe(prName, namespace string) (string, error) {
	return k.prDescribe, nil
}

func (k *KinterfaceTest) TektonCliFollowLogs(prName, namespace string) (string, error) {
	return k.prDescribe, nil
}