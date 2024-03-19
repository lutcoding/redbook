package internal

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/lutcoding/redbook/internal/config"
	"github.com/lutcoding/redbook/internal/events"
	articleMsgQueue "github.com/lutcoding/redbook/internal/events/article"
	"github.com/lutcoding/redbook/internal/repository"
	"github.com/lutcoding/redbook/internal/repository/cache"
	articleDao "github.com/lutcoding/redbook/internal/repository/dao/article"
	"github.com/lutcoding/redbook/internal/service"
	articleService "github.com/lutcoding/redbook/internal/service/article"
	"github.com/lutcoding/redbook/internal/service/oauth/dingtalk"
	"github.com/lutcoding/redbook/internal/service/oauth/wechat"
	"github.com/lutcoding/redbook/internal/service/sms/memory"
	"github.com/lutcoding/redbook/internal/web/article"
	"github.com/lutcoding/redbook/internal/web/jwt"
	"github.com/lutcoding/redbook/internal/web/oauth"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	route       *gin.Engine
	srv         *http.Server
	db          *gorm.DB
	redis       redis.Cmdable
	cfg         config.Config
	mongo       *mongo.Client
	msgConsumer []events.Consumer
	kafkaClient sarama.Client

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
	//err = s.initMongo()
	//if err != nil {
	//	return err
	//}
	err = s.initSarama()
	if err != nil {
		return err
	}
	s.msgConsumer, err = s.initMsgConsumer()
	if err != nil {
		return err
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
	err = s.startConsumer()
	if err != nil {
		return err
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

func (s *Server) initSarama() error {
	var err error
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	s.kafkaClient, err = sarama.NewClient(s.cfg.Kafka.Addrs, cfg)
	return err
}

func (s *Server) initMsgConsumer() ([]events.Consumer, error) {
	return nil, nil
}

func (s *Server) startConsumer() error {
	for _, consumer := range s.msgConsumer {
		err := consumer.Start()
		if err != nil {
			return err
		}
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

func (s *Server) initMongo() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ops := options.Client().
		ApplyURI(s.cfg.DB.Mongo.URI).
		SetAuth(options.Credential{
			Username: s.cfg.DB.Mongo.Username,
			Password: s.cfg.DB.Mongo.Password,
		})
	client, err := mongo.Connect(ctx, ops)
	if err != nil {
		return err
	}
	s.mongo = client
	return nil
}

func (s *Server) initHandlers() error {
	artReadProducer, err := sarama.NewSyncProducerFromClient(s.kafkaClient)
	if err != nil {
		return err
	}
	articleReadProducer := articleMsgQueue.NewKafkaProducer(artReadProducer, "article_read")

	userDAO := dao.NewUserGormDAO(s.db)
	articleDAO := articleDao.NewGORMArticleDao(s.db)
	interactiveDAO := dao.NewGORMInteractiveDAO(s.db)

	userCache := cache.NewUserRedisCache(s.redis)
	codeCache := cache.NewCodeRedisCache(s.redis)
	articleCache := cache.NewArticleRedisCache(s.redis)
	interactiveCache := cache.NewInteractiveRedisCache(s.redis)

	userRepo := repository.NewUserCacheRepository(userDAO, userCache)
	codeRepo := repository.NewCodeCacheRepository(codeCache)
	articleRepo := repository.NewArticleCacheRepository(articleDAO, articleCache)
	interactiveRepo := repository.NewInteractiveCacheRepository(interactiveDAO, interactiveCache)

	userSvc := service.NewUserService(userRepo)
	smsRateLimitSvc := smsratelimit.NewService(memory.NewService(),
		ratelimit.NewRedisSlidingWindowLimiter(s.redis, time.Minute, 10))
	codeSvc := service.NewCodeService(codeRepo, smsRateLimitSvc, "1")
	wechatSvc := wechat.NewService(s.cfg.Wechat.AppID, s.cfg.Wechat.AppSecret)
	dingTalkSvc := dingtalk.NewService(s.cfg.Ding.AppKey, s.cfg.Ding.AppSecret)
	articleSvc := articleService.NewService(articleRepo, articleReadProducer)
	interactiveSvc := service.NewInteractiveService(interactiveRepo)

	s.jwtHandler = jwt.NewHandler()
	s.userHandler = user.New(userSvc, codeSvc, s.jwtHandler)
	s.oauth2WeChatHandler = oauth.NewOAuth2WeChatHandler(wechatSvc, userSvc)
	s.oAuth2DingTalkHandler = oauth.NewOAuth2DingTalkHandler(dingTalkSvc, userSvc)
	s.articleHandler = article.NewHandler(articleSvc, interactiveSvc)
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
			draft := ag.Group("/draft")
			{
				draft.POST("/create", s.articleHandler.Create)
				draft.POST("/list", s.articleHandler.ListDraft)
				draft.POST("/update", s.articleHandler.Edit)
				draft.POST("/get", s.articleHandler.GetDraft)
			}
			published := ag.Group("/published")
			{
				published.POST("/publish", s.articleHandler.Publish)
				published.POST("/private", s.articleHandler.ToPrivate)
				// 列出他人的已发表文章
				published.POST("/list", s.articleHandler.ListPub)
				published.GET("/get/:id", s.articleHandler.GetPub)
				published.POST("/like", s.articleHandler.Like)
			}
		}
	}
	return engine
}
