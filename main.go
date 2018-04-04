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
	NormTitle   string        `json:"normTitle" bson:"normTitle"`
	Poster      string        `json:"poster" bson:"poster"`
	BigPoster   string        `json:"bigPoster" bson:"bigPoster"`
	PlayUrl     string        `json:"playUrl" bson:"playUrl"`
	Source      string        `json:"source" bson:"source"`
	VideoSource string        `json:"videoSource" bson:"videoSource"`
	Content     string        `json:"content" bson:"content"`
	ReleaseDate string        `json:"releaseDate" bson:"releaseDate"`
	Directors   []string      `json:"directors" bson:"directors"`
	//Actors          []string          `json:"actors" bson:"actors"`
	//ActorEmbeded    []ActorEmbeded    `json:"actorEmbeded" bson:"actorEmbeded"`
	ActorEmbeded []ActorEmbeded `json:"actors" bson:"actorEmbeded"`
	// Categories      []string          `json:"categories" bson:"categories"`
	// CategoryEmbeded []CategoryEmbeded `json:"categoryEmbeded" bson:"categoryEmbeded"`
	CategoryEmbeded []CategoryEmbeded `json:"categories" bson:"categoryEmbeded"`
	Countries       []string          `json:"countries" bson:"countries"`
	CountryEmbeded  []CountryEmbeded  `json:"countryEmbeded" bson:"countryEmbeded"`
	Class           string            `json:"_class" bson:"_class"`
}

type MovieLite struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Poster      string `json:"poster"`
	ReleaseDate string `json:"releaseDate"`
}

type Category struct {
	Id    bson.ObjectId `json:"_id" bson:"_id"`
	Key   string        `json:"key" bson:"key"`
	Title string        `json:"title" bson:"title"`
}

type CategoryEmbeded struct {
	Key   string `json:"key"`
	Title string `json:"title"`
}

type Country struct {
	Id    bson.ObjectId `json:"_id" bson:"_id"`
	Key   string        `json:"key" bson:"key"`
	Title string        `json:"title" bson:"title"`
}

type CountryEmbeded struct {
	Key   string `json:"key"`
	Title string `json:"title"`
}

type Actor struct {
	Id    bson.ObjectId `json:"_id" bson:"_id"`
	Key   string        `json:"key" bson:"key"`
	Title string        `json:"title" bson:"title"`
}

type ActorEmbeded struct {
	Key   string `json:"key"`
	Title string `json:"title"`
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
	db := s.DB(os.Getenv("GOOD_MOVIE_DB_NAME"))
	movies := db.C("movies")
	categories := db.C("categories")
	actors := db.C("actors")

	g := r.Group("/api/gm")
	g.GET("/paginated", func(c *gin.Context) {
		performMovieQuery(c, movies, bson.M{
			"videoSource": bson.M{"$exists": true, "$ne": ""},
		})
	})
	g.GET("/search", func(c *gin.Context) {
		q := c.Query("q")
		performMovieQuery(c, movies, bson.M{
			"title":       bson.M{"$regex": ".*" + q + ".*", "$options": "-i"},
			"videoSource": bson.M{"$exists": true, "$ne": ""},
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

	g.GET("/search/byActorKey", func(c *gin.Context) {
		actorKey := c.Query("key")
		performMovieQuery(c, movies, bson.M{
			"actorEmbeded.key": actorKey,
			"videoSource":      bson.M{"$exists": true, "$ne": ""},
		})
	})

	g.GET("/search/byCategoryKey", func(c *gin.Context) {
		categoryKey := c.Query("key")
		performMovieQuery(c, movies, bson.M{
			"categoryEmbeded.key": categoryKey,
			"videoSource":         bson.M{"$exists": true, "$ne": ""},
		})
	})

	g.GET("/category", func(c *gin.Context) {
		cts := []Category{}
		if err := categories.Find(nil).All(&cts); err != nil {
			panic(err)
		}
		c.JSON(200, cts)
	})

	g.GET("/actor", func(c *gin.Context) {
		cts := []Actor{}
		if err := actors.Find(nil).All(&cts); err != nil {
			panic(err)
		}
		c.JSON(200, cts)
	})

	r.Run(":" + os.Getenv("GOOD_MOVIE_PORT"))
}

func renderResult(c *gin.Context, page int, size int, total int, result []Movie) {
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
		"items": lites,
		"paging": gin.H{
			"page":      page,
			"size":      size,
			"totalPage": total/size + 1,
			"totalItem": total,
		},
	})
}

func getPagingQuery(c *gin.Context) (int, int) {
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		page = 1
	}
	size, err := strconv.Atoi(c.Query("size"))
	if err != nil {
		size = 100
	}
	return page, size
}

func performMovieQuery(c *gin.Context, collection *mgo.Collection, query bson.M) {
	page, size := getPagingQuery(c)
	result := []Movie{}
	found := collection.Find(query)
	total, err := found.Count()
	if err != nil {
		panic(err)
	}
	err = found.Skip((page - 1) * size).Limit(size).All(&result)
	if err != nil {
		panic(err)
	}
	renderResult(c, page, size, total, result)
}
