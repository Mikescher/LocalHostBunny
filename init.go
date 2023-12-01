package bunny

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gogs.mikescher.com/BlackForestBytes/goext/exerr"
	"gogs.mikescher.com/BlackForestBytes/goext/langext"
	"os"
)

func Init(cfg Config) {
	exerr.Init(exerr.ErrorPackageConfigInit{
		ZeroLogErrTraces:       langext.PTrue,
		ZeroLogAllTraces:       langext.PTrue,
		RecursiveErrors:        langext.PTrue,
		ExtendedGinOutput:      &cfg.ReturnRawErrors,
		IncludeMetaInGinOutput: &cfg.ReturnRawErrors,
		ExtendGinOutput: func(err *exerr.ExErr, json map[string]any) {
			if fapiMsg := err.RecursiveMeta("fapiMessage"); fapiMsg != nil {
				json["fapiMessage"] = fapiMsg.ValueString()
			}
		},
		ExtendGinDataOutput: func(err *exerr.ExErr, depth int, json map[string]any) {
			if fapiMsg, ok := err.Meta["fapiMessage"]; ok {
				json["fapiMessage"] = fapiMsg.ValueString()
			}
		},
	})

	cw := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "2006-01-02 15:04:05 Z07:00",
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	multi := zerolog.MultiLevelWriter(cw)
	logger := zerolog.New(multi).With().
		Timestamp().
		Caller().
		Logger()

	log.Logger = logger

	if cfg.GinDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	zerolog.SetGlobalLevel(cfg.LogLevel)

	log.Debug().Msg("Initialized")
}
