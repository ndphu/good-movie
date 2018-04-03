package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"strconv"
)

type Movie struct {
	Id          bson.ObjectId `json:"id" bson:"_id"`
	Title       string        `json:"title" bson:"title"`
	Poster      string        `json:"poster" bson:"poster"`
	BigPoster   string        `json:"bigPoster" bson:"bigPoster"`
	Origin      string        `json:"origin" bson:"origin"`
	PlayUrl     string        `json:"playUrl" bson:"playUrl"`
	Source      string        `json:"source" bson:"source"`
	VideoSource string        `json:"videoSource" bson:"videoSource"`
	Content     string        `json:"content" bson:"content"`
	ReleaseDate string        `json:"releaseDate" bson:"releaseDate"`
	Directors   []string      `json:"directors" bson:"directors"`
	Actors      []string      `json:"actors" bson:"actors"`
	Categories  []string      `json:"categories" bson:"categories"`
	Countries   []string      `json:"countries" bson:"countries"`
}

type MovieLite struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Poster      string `json:"poster"`
	ReleaseDate string `json:"releaseDate"`
}

func main() {
	r := gin.Default()
	c := cors.DefaultConfig()
	c.AllowAllOrigins = true
	c.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	c.AllowHeaders = []string{"Origin", "Authorization", "Content-Type", "Content-Length", "X-Requested-With"}
	c.AllowCredentials = true
	r.Use(cors.New(c))

	s, err := mgo.Dial(os.Getenv("GOOD_MOVIE_MONGODB"))
	if err != nil {
		panic(err)
	}
	defer s.Close()

	movies := s.DB(os.Getenv("GOOD_MOVIE_DB_NAME")).C("movies")

	g := r.Group("/api/gm")
	g.GET("/paginated", func(c *gin.Context) {
		page, _ := strconv.Atoi(c.Query("page"))
		size, _ := strconv.Atoi(c.Query("size"))
		result := []Movie{}
		total, err := movies.Count()
		if err != nil {
			panic(err)
		}
		movies.Find(nil).Skip((page - 1) * size).Limit(size).All(&result)
		lites := make([]MovieLite, len(result))
		for i, e := range result {
			lites[i] = MovieLite{
				Id:          e.Id.Hex(),
				Title:       e.Title,
				Poster:      e.Poster,
				ReleaseDate: e.ReleaseDate,
			}
		}
		c.JSON(200, gin.H{
			"items":     lites,
			"page":      page,
			"size":      size,
			"totalItem": total,
			"totalPage": total/size + 1,
		})
	})
	g.GET("/search", func(c *gin.Context) {
		q := c.Query("q")
		page, err := strconv.Atoi(c.Query("page"))
		if err != nil {
			page = 1
		}
		size, err := strconv.Atoi(c.Query("size"))
		if err != nil {
			size = 100
		}
		total, err := movies.Count()
		if err != nil {
			panic(err)
		}
		result := []Movie{}
		movies.Find(bson.M{"title": bson.M{"$regex": ".*" + q + ".*", "$options": "-i"}}).Skip((page - 1) * size).Limit(size).All(&result)
		lites := make([]MovieLite, len(result))
		for i, e := range result {
			lites[i] = MovieLite{
				Id:          e.Id.Hex(),
				Title:       e.Title,
				Poster:      e.Poster,
				ReleaseDate: e.ReleaseDate,
			}
		}
		c.JSON(200, gin.H{
			"items":     lites,
			"page":      page,
			"size":      size,
			"totalItem": total,
			"totalPage": total/size + 1,
		})
	})
	g.GET("/movie/:id", func(c *gin.Context) {
		id := c.Param("id")
		movie := Movie{}
		err := movies.FindId(bson.ObjectIdHex(id)).One(&movie)
		if err == mgo.ErrNotFound {
			c.JSON(404, gin.H{"err": "movie not found"})
		} else {
			c.JSON(200, movie)
		}
	})

	r.Run(":" + os.Getenv("GOOD_MOVIE_PORT"))
}
