package cmd

import "golang.org/x/sys/windows"

var shellExecute = windows.ShellExecute

func openBrowser(targetURL string) error {
	url, err := windows.UTF16PtrFromString(targetURL)
	if err != nil {
		return err
	}
	return shellExecute(0, nil, url, nil, nil, windows.SW_SHOWNORMAL)
}
