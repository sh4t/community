package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/sh4t/community/host"
	"github.com/sh4t/community/mdw"
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
	commonHandlers := alice.New(context.ClearHandler, mdw.LoggingHandler, mdw.RecoverHandler, mdw.CorsHandler, mdw.AcceptHandler)
	router := NewRouter()
	router.Get("/hosts/:id", commonHandlers.ThenFunc(appC.hostHandler))
	router.Put("/hosts/:id", commonHandlers.Append(mdw.ContentTypeHandler, mdw.BodyHandler(host.HostResource{})).ThenFunc(appC.updateHostHandler))
	router.Delete("/hosts/:id", commonHandlers.ThenFunc(appC.deleteHostHandler))
	router.Get("/hosts", commonHandlers.ThenFunc(appC.hostsHandler))
	router.Post("/hosts", commonHandlers.Append(mdw.ContentTypeHandler, mdw.BodyHandler(host.HostResource{})).ThenFunc(appC.createHostHandler))
	http.ListenAndServe(":8080", router)
}