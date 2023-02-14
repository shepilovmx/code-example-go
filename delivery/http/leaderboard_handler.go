package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"dummypath/entity"
	"dummypath/go/gojwt"
	"dummypath/go/internal/calculator"
	"dummypath/go/internal/config"
	"dummypath/go/internal/container"
	"dummypath/go/internal/middleware"
	"dummypath/go/leaderboard"
	"dummypath/go/user"
	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

type leaderboardHTTPHandler struct {
	conf       config.Root
	log        *logrus.Entry
	mdw        middleware.Middleware
	calc       calculator.Calc
	gojwt      gojwt.JWT
	jwtService gojwt.JWT
	us         user.Service
	ls         leaderboard.Service
}

type LeaderboardQuery struct {
	Year  int64
	Month int64
}

func NewHandler(c container.Container) {
	h := &leaderboardHTTPHandler{
		log:        c.Logger().WithField("http", "leaderboard_http_handler"),
		conf:       c.Config(),
		mdw:        c.Middleware(),
		jwtService: c.JWT(),
		us:         c.UserService(),
		calc:       c.CalcService(),
		gojwt:      c.JWT(),
		ls:         c.LeaderboardService(),
	}

	c.Router().Get("/leaderboard", h.GetByDate())
}

// GetByDate godoc
// @Tags Leaderboard
// @Summary Returns a leaderboard by year and month
// @Description Returns a leaderboard by year and month
// @Produce json
// @Success 200 {object} entity.Leaderboard
// @Failure 400
// @Failure 401
// @Failure 500
// @Param year query string false "year"
// @Param month query string false "month"
// @Router /leaderboard [get]
func (h leaderboardHTTPHandler) GetByDate() http.HandlerFunc {
	_ = h.log.WithField("handler", "GetByDate")

	return func(w http.ResponseWriter, r *http.Request) {
		q := LeaderboardQuery{
			Year:  int64(time.Now().Year()),
			Month: int64(time.Now().Month()),
		}

		// get year
		if y := r.URL.Query().Get("year"); y != "" {
			q.Year, _ = strconv.ParseInt(y, 10, 64)
		}

		// get month
		if m := r.URL.Query().Get("month"); m != "" {
			q.Month, _ = strconv.ParseInt(m, 10, 64)
		}

		lboard, err := h.ls.FindOneByYearAndMonth(context.Background(), int(q.Year), time.Month(q.Month))
		if err != nil && err != entity.ErrNotFound {
			h.log.WithError(err).Errorf("h.ls.FindOneByYearAndMonth(context.Background(), %v, %v) failed", q.Year, q.Month)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jsoniter.NewEncoder(w).Encode(lboard)
	}
}
