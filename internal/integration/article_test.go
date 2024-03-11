package integration

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/lutcoding/redbook/common/globalkey"
	"github.com/lutcoding/redbook/internal/repository"
	"github.com/lutcoding/redbook/internal/repository/dao"
	articleService "github.com/lutcoding/redbook/internal/service/article"
	"github.com/lutcoding/redbook/internal/web/article"
	"github.com/lutcoding/redbook/pkg/ginx/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ArticleSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func (s *ArticleSuite) SetupSuite() {
	s.server = gin.Default()
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/webook"))
	assert.NoError(s.T(), err)
	err = dao.InitTables(db)
	assert.NoError(s.T(), err)
	s.db = db
	articleDAO := dao.NewGORMArticleDao(db)
	articleRepo := repository.NewArticleCacheRepository(articleDAO)
	svc := articleService.NewService(articleRepo)
	handler := article.NewHandler(svc)
	s.server.Use(func(ctx *gin.Context) {
		ctx.Set(globalkey.JwtUserId, int64(1))
	})
	s.server.POST("/articles/edit", handler.Edit)
}

func (s *ArticleSuite) TearDownTest() {
	err := s.db.Exec("TRUNCATE TABLE `articles`").Error
	assert.NoError(s.T(), err)
}

func (s *ArticleSuite) TestEdit() {
	t := s.T()
	testCases := []struct {
		name string

		before func(t *testing.T)
		after  func(t *testing.T)

		req func() io.Reader

		wantCode int
		wantRes  middlewares.Result[int64]
	}{
		{
			name:   "新建帖子",
			before: func(t *testing.T) {},
			after: func(t *testing.T) {

			},
			req: func() io.Reader {
				req := ArticleEditReq{Tittle: "hello", Content: "hello"}
				data, err := json.Marshal(req)
				assert.NoError(t, err)
				return bytes.NewBuffer(data)
			},
			wantCode: http.StatusOK,
			wantRes:  middlewares.Result[int64]{Data: 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			req, err := http.NewRequest(http.MethodPost,
				"/articles/create", tc.req())
			require.NoError(t, err)
			// 数据是 JSON 格式
			req.Header.Set("Content-Type", "application/json")
			// 这里你就可以继续使用 req

			resp := httptest.NewRecorder()
			// 这就是 HTTP 请求进去 GIN 框架的入口。
			// 当你这样调用的时候，GIN 就会处理这个请求
			// 响应写回到 resp 里
			s.server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			if resp.Code != 200 {
				return
			}
			var res middlewares.Result[int64]
			err = json.Unmarshal(resp.Body.Bytes(), &res)
			assert.NoError(t, err)
			assert.Equal(t, res, tc.wantRes)
			tc.after(t)
		})
	}
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleSuite{})
}

type ArticleEditReq struct {
	Tittle  string `json:"tittle"`
	Content string `json:"content"`
}
