package logger

import (
	"context"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	hertzlogrus "github.com/hertz-contrib/obs-opentelemetry/logging/logrus"
	"github.com/sirupsen/logrus"
	fxlogrus "github.com/takt-corp/fx-logrus"
	"go.uber.org/fx/fxevent"
	"os"
)

func FxLogger(_ context.Context, logger *logrus.Logger) fxevent.Logger {
	return &fxlogrus.LogrusLogger{
		Logger: logger,
	}
}

func NewLogger(ctx context.Context) (*logrus.Logger, error) {

	l := logrus.StandardLogger()

	level, err := logrus.ParseLevel("debug")
	if err != nil {
		logrus.WithContext(ctx).WithError(err).Error("failed to parse log level")
		return nil, err
	}

	l.SetLevel(level)
	l.SetOutput(os.Stdout)

	hlog.SetLogger(hertzlogrus.NewLogger(hertzlogrus.WithLogger(l)))

	l.SetFormatter(&logrus.TextFormatter{})

	return l, nil
}
