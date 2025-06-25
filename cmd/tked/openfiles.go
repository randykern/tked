package main

import "tked/internal/app"

func openFiles(application app.App, filenames []string) error {
	var firstView app.View
	for i, filename := range filenames {
		if err := application.OpenFile(filename); err != nil {
			return err
		}
		if i == 0 {
			firstView = application.GetCurrentView()
		}
	}
	if firstView != nil {
		application.SetCurrentView(firstView)
	}
	return nil
}
