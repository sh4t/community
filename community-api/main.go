package main

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/sh4t/community/host"
	"github.com/sh4t/community/errs"
)

type HostRepo struct {
	coll *mgo.Collection
}

func (r *HostRepo) All() (host.HostsCollection, error) {
	result := host.HostsCollection{[]host.Host{}}
	err := r.coll.Find(nil).All(&result.Data)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r *HostRepo) Find(id string) (host.HostResource, error) {
	result := host.HostResource{}
	err := r.coll.FindId(bson.ObjectIdHex(id)).One(&result.Data)
	if err != nil {
		return result, err
	}

	return result, nil
}

func (r *HostRepo) Create(host *host.Host) error {
	id := bson.NewObjectId()
	host.CreatedAt = time.Now()
	host.ModifiedAt = time.Now()
	_, err := r.coll.UpsertId(id, host)
	if err != nil {
		return err
	}

	host.Id = id

	return nil
}

func (r *HostRepo) Update(host *host.Host) error {
	host.ModifiedAt = time.Now()
	err := r.coll.UpdateId(host.Id, host)
	if err != nil {
		return err
	}

	return nil
}

func (r *HostRepo) Delete(id string) error {
	err := r.coll.RemoveId(bson.ObjectIdHex(id))
	if err != nil {
		return err
	}

	return nil
}


// all in the middle..
func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				errs.WriteError(w, errs.ErrInternalServer)
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// cors is a bit annoying, but here is an attempt to resolve for community-ui angular love.
func corsHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		if origin := r.Header.Get("Origin"); origin != "" {
		    w.Header().Set("Access-Control-Allow-Origin", origin)
		    w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		    w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}
		
	    if r.Method == "OPTIONS" {
	        return
	    }

	    next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func loggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}

	return http.HandlerFunc(fn)
}

func acceptHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/vnd.api+json" {
			errs.WriteError(w, errs.ErrNotAcceptable)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func contentTypeHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/vnd.api+json" {
			errs.WriteError(w, errs.ErrUnsupportedMediaType)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}



func bodyHandler(v interface{}) func(http.Handler) http.Handler {
	t := reflect.TypeOf(v)

	m := func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			val := reflect.New(t).Interface()
			err := json.NewDecoder(r.Body).Decode(val)

			if err != nil {
				errs.WriteError(w, errs.ErrBadRequest)
				return
			}

			if next != nil {
				context.Set(r, "body", val)
				next.ServeHTTP(w, r)
			}
		}

		return http.HandlerFunc(fn)
	}

	return m
}

// handlers

type appContext struct {
	db *mgo.Database
}

func (c *appContext) hostsHandler(w http.ResponseWriter, r *http.Request) {
	repo := HostRepo{c.db.C("hosts")}
	hosts, err := repo.All()
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	json.NewEncoder(w).Encode(hosts)
}

func (c *appContext) hostHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	repo := HostRepo{c.db.C("hosts")}
	host, err := repo.Find(params.ByName("id"))
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	json.NewEncoder(w).Encode(host)
}

func (c *appContext) createHostHandler(w http.ResponseWriter, r *http.Request) {
	body := context.Get(r, "body").(*host.HostResource)
	repo := HostRepo{c.db.C("hosts")}
	err := repo.Create(&body.Data)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(body)
}

func (c *appContext) updateHostHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	body := context.Get(r, "body").(*host.HostResource)
	body.Data.Id = bson.ObjectIdHex(params.ByName("id"))
	repo := HostRepo{c.db.C("hosts")}
	err := repo.Update(&body.Data)
	if err != nil {
		panic(err)
	}

	w.WriteHeader(204)
	w.Write([]byte("\n"))
}

func (c *appContext) deleteHostHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	repo := HostRepo{c.db.C("hosts")}
	err := repo.Delete(params.ByName("id"))
	if err != nil {
		panic(err)
	}

	w.WriteHeader(204)
	w.Write([]byte("\n"))
}

// ROUTER

type router struct {
	*httprouter.Router
}

func (r *router) Get(path string, handler http.Handler) {
	r.GET(path, wrapHandler(handler))
}

func (r *router) Post(path string, handler http.Handler) {
	r.POST(path, wrapHandler(handler))
}

func (r *router) Put(path string, handler http.Handler) {
	r.PUT(path, wrapHandler(handler))
}

func (r *router) Delete(path string, handler http.Handler) {
	r.DELETE(path, wrapHandler(handler))
}

func NewRouter() *router {
	return &router{httprouter.New()}
}

func wrapHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		context.Set(r, "params", ps)
		h.ServeHTTP(w, r)
	}
}

func main() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	appC := appContext{session.DB("community")}
	commonHandlers := alice.New(context.ClearHandler, loggingHandler, recoverHandler, corsHandler, acceptHandler)
	router := NewRouter()
	router.Get("/hosts/:id", commonHandlers.ThenFunc(appC.hostHandler))
	router.Put("/hosts/:id", commonHandlers.Append(contentTypeHandler, bodyHandler(host.HostResource{})).ThenFunc(appC.updateHostHandler))
	router.Delete("/hosts/:id", commonHandlers.ThenFunc(appC.deleteHostHandler))
	router.Get("/hosts", commonHandlers.ThenFunc(appC.hostsHandler))
	router.Post("/hosts", commonHandlers.Append(contentTypeHandler, bodyHandler(host.HostResource{})).ThenFunc(appC.createHostHandler))
	http.ListenAndServe(":8080", router)
}