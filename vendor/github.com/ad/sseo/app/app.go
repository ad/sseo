package app

import (
	"context"
	"io"

	"github.com/ad/sseo/checker"
	"github.com/ad/sseo/server"
)

func Run(ctx context.Context, w io.Writer, args []string) error {
	urlChecker := checker.InitChecker()

	_, errInitListener := server.InitListener(urlChecker)
	if errInitListener != nil {
		return errInitListener
	}

	return nil
}
