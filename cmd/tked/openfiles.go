package main

import "tked/internal/app"

func openFiles(application app.App, filenames []string) error {
	for _, filename := range filenames {
		if err := application.OpenFile(filename); err != nil {
			return err
		}
	}
	return nil
}
