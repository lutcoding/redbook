package internal

import (
	"github.com/lutcoding/redbook/internal/repository"
	"github.com/lutcoding/redbook/internal/repository/cache"
	"github.com/lutcoding/redbook/internal/service"
	"github.com/lutcoding/redbook/internal/service/sms/memory"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lutcoding/redbook/internal/repository/dao"
	"github.com/lutcoding/redbook/internal/web/middleware"
	"github.com/lutcoding/redbook/internal/web/user"
	"github.com/lutcoding/redbook/pkg/ginx/middlewares/ratelimit"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Server struct {
	route *gin.Engine
	srv   *http.Server
	db    *gorm.DB
	redis redis.Cmdable

	userHandler *user.Handler
}

func NewServer() (*Server, error) {
	s := &Server{}
	return s, nil
}

func (s *Server) Serve(addr string) (err error) {
	if s.db, err = gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook")); err != nil {
		return
	}
	s.redis = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if err = dao.InitTables(s.db); err != nil {
		return
	}
	if err = s.initHandlers(); err != nil {
		return
	}
	s.route = s.newRouter()
	s.srv = &http.Server{
		Addr:              addr,
		Handler:           s.route,
		ReadHeaderTimeout: time.Second * 15,
	}
	err = s.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) initHandlers() (err error) {
	userDAO := dao.NewUserGormDAO(s.db)
	userCache := cache.NewUserRedisCache(s.redis)
	codeCache := cache.NewCodeRedisCache(s.redis)
	userRepo := repository.NewUserCacheRepository(userDAO, userCache)
	codeRepo := repository.NewCodeCacheRepository(codeCache)

	userSvc := service.NewUserService(userRepo)
	smsSvc := memory.NewService()
	codeSvc := service.NewCodeService(codeRepo, smsSvc, "1")

	s.userHandler, err = user.New(userSvc, codeSvc)
	if err != nil {
		return err
	}
	return
}

// TODO: RESTful api
func (s *Server) newRouter() *gin.Engine {
	engine := gin.Default()
	// 允许跨域
	engine.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"http://localhost"},
		AllowMethods:     []string{"POST", "GET"},
		AllowHeaders:     []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
		//不加的话前端拿不到token
		ExposeHeaders: []string{"x-jwt-token"},
		// 替换 AllowOrigins
		AllowOriginFunc: func(origin string) bool {
			// 本地测试环境
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			// 生产环境域名
			return strings.Contains(origin, "yourcompany.com")
		},
		MaxAge: 12 * time.Hour,
	}))

	engine.Use(ratelimit.NewBuilder(ratelimit.NewRedisSlidingWindowLimiter(s.redis, time.Minute, 10)).Build())
	// 注册handler
	root := engine.Group("/")
	unauthorized := root.Group("/")
	{
		unauthorized.POST("/users/signup", s.userHandler.SignUp)
		unauthorized.POST("/users/login", s.userHandler.Login)
		unauthorized.POST("/users/login_sms/code/send", s.userHandler.SendLoginSmsCode)
		unauthorized.POST("/users/login_sms", s.userHandler.LoginSmsCode)
	}
	authorized := root.Group("/", middleware.NewLoginMiddlewareBuilder().Build())
	{
		ug := authorized.Group("/users")
		{
			ug.POST("/edit", s.userHandler.Edit)
			ug.GET("/profile", s.userHandler.Profile)
		}
	}
	return engine
}
