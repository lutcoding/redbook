package internal

import (
	"github.com/lutcoding/redbook/internal/config"
	"github.com/lutcoding/redbook/internal/repository"
	"github.com/lutcoding/redbook/internal/repository/cache"
	article2 "github.com/lutcoding/redbook/internal/repository/dao/article"
	"github.com/lutcoding/redbook/internal/service"
	articleService "github.com/lutcoding/redbook/internal/service/article"
	"github.com/lutcoding/redbook/internal/service/oauth/dingtalk"
	"github.com/lutcoding/redbook/internal/service/oauth/wechat"
	"github.com/lutcoding/redbook/internal/service/sms/memory"
	"github.com/lutcoding/redbook/internal/web/article"
	"github.com/lutcoding/redbook/internal/web/jwt"
	"github.com/lutcoding/redbook/internal/web/oauth"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	sr "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/lutcoding/redbook/internal/repository/dao"
	smsratelimit "github.com/lutcoding/redbook/internal/service/sms/ratelimit"
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
	cfg   config.Config

	jwtHandler            *jwt.Handler
	userHandler           *user.Handler
	oauth2WeChatHandler   *oauth.OAuth2WeChatHandler
	oAuth2DingTalkHandler *oauth.OAuth2DingTalkHandler
	articleHandler        *article.Handler
}

func NewServer() (*Server, error) {
	viper.SetConfigFile("etc/dev.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	s := &Server{}
	err = viper.Unmarshal(&s.cfg)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Server) Serve(addr string) (err error) {
	err = s.initLog()
	if err != nil {
		return
	}
	if s.db, err = gorm.Open(mysql.Open(s.cfg.DB.Mysql.DSN)); err != nil {
		return
	}
	s.redis = redis.NewClient(&redis.Options{
		Addr: s.cfg.Redis.Addr,
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
		return
	}
	return nil
}

func (s *Server) initLog() (err error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)
	return nil
}

func (s *Server) initHandlers() (err error) {
	userDAO := dao.NewUserGormDAO(s.db)
	articleDAO := article2.NewGORMArticleDao(s.db)
	userCache := cache.NewUserRedisCache(s.redis)
	codeCache := cache.NewCodeRedisCache(s.redis)
	userRepo := repository.NewUserCacheRepository(userDAO, userCache)
	codeRepo := repository.NewCodeCacheRepository(codeCache)
	articleRepo := repository.NewArticleCacheRepository(articleDAO)

	userSvc := service.NewUserService(userRepo)
	smsRateLimitSvc := smsratelimit.NewService(memory.NewService(),
		ratelimit.NewRedisSlidingWindowLimiter(s.redis, time.Minute, 10))
	codeSvc := service.NewCodeService(codeRepo, smsRateLimitSvc, "1")
	wechatSvc := wechat.NewService(s.cfg.Wechat.AppID, s.cfg.Wechat.AppSecret)
	dingTalkSvc := dingtalk.NewService(s.cfg.Ding.AppKey, s.cfg.Ding.AppSecret)
	articleSvc := articleService.NewService(articleRepo)

	s.jwtHandler = jwt.NewHandler()
	s.userHandler = user.New(userSvc, codeSvc, s.jwtHandler)
	s.oauth2WeChatHandler = oauth.NewOAuth2WeChatHandler(wechatSvc, userSvc)
	s.oAuth2DingTalkHandler = oauth.NewOAuth2DingTalkHandler(dingTalkSvc, userSvc)
	s.articleHandler = article.NewHandler(articleSvc)
	return nil
}

func (s *Server) newRouter() *gin.Engine {
	engine := gin.Default()
	store, _ := sr.NewStore(10, "tcp", s.cfg.Redis.Addr, "", []byte("secret"))
	engine.Use(sessions.Sessions("SESSION", store))
	// 允许跨域
	engine.Use(cors.New(cors.Config{
		// AllowOrigins:     []string{"http://localhost"},
		AllowMethods:     []string{"POST", "GET"},
		AllowHeaders:     []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
		//不加的话前端拿不到token
		ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
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

	engine.Use(ratelimit.NewBuilder(ratelimit.NewRedisSlidingWindowLimiter(s.redis, time.Minute, 100)).Build())
	// 注册handler
	root := engine.Group("/")
	unauthorized := root.Group("/")
	{
		unauthorized.POST("/users/signup", s.userHandler.SignUp)
		unauthorized.POST("/users/login", s.userHandler.Login)
		unauthorized.POST("/users/login_sms/code/send", s.userHandler.SendLoginSmsCode)
		unauthorized.POST("/users/login_sms", s.userHandler.LoginSmsCode)
		unauthorized.GET("/users/refresh", s.userHandler.Refresh)
		oauth2 := unauthorized.Group("/oauth2")
		{
			wg := oauth2.Group("/wechat")
			{
				wg.GET("/authurl", s.oauth2WeChatHandler.AuthURL)
				wg.Any("/callback", s.oauth2WeChatHandler.CallBack)
			}
			dg := oauth2.Group("/dingtalk")
			{
				dg.GET("/authurl", s.oAuth2DingTalkHandler.AuthURL)
				dg.Any("/callback", s.oAuth2DingTalkHandler.CallBack)
			}
		}
	}

	authorized := root.Group("/", middleware.NewLoginMiddlewareBuilder(s.jwtHandler).Build())
	{
		ug := authorized.Group("/users")
		{
			ug.POST("/edit", s.userHandler.Edit)
			ug.GET("/profile", s.userHandler.Profile)
			ug.POST("/logout", s.userHandler.Logout)
		}

		ag := authorized.Group("/articles")
		{
			ag.POST("/create", s.articleHandler.Create)
			ag.POST("/publish", s.articleHandler.Publish)
			ag.POST("/private", s.articleHandler.ToPrivate)
		}
	}
	return engine
}
