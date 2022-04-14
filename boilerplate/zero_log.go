package boilerplate

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmzerolog/v2"
	"math/rand"
	"os"
	"time"
)

var isConfigured bool

func SetupZeroLog() {
	if isConfigured {
		return
	}

	rand.Seed(time.Now().Unix())
	log.Logger = zerolog.New(os.Stderr).With().Caller().
		Time("time", time.Now().UTC()).Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr, NoColor: true})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	zerolog.DefaultContextLogger = &log.Logger

	isConfigured = true
}

func CreateCustomContext(ctx context.Context, apmTx *apm.Transaction, logger zerolog.Logger) context.Context {
	if apmTx != nil {
		ctx = apm.ContextWithTransaction(ctx, apmTx)
		logger = logger.Hook(apmzerolog.TraceContextHook(ctx))
	}

	ctx = logger.WithContext(ctx)

	return ctx
}
