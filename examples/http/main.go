// Example: HTTP server wrapper (to be used with manual client).
package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/logging"
	fsdb "git.defalsify.org/vise.git/db/fs"
)

var (
	logg logging.Logger = logging.NewVanilla().WithDomain("http")
)

type LocalHandler struct {
	sessionId string
}	

func NewLocalHandler() *LocalHandler {
	return &LocalHandler{
		sessionId: "",
	}
}

func(h* LocalHandler) SetSession(sessionId string) {
	h.sessionId = sessionId
}

func(h* LocalHandler) AddSession(ctx context.Context, sym string, input []byte) (resource.Result, error) {
	return resource.Result{
		Content: h.sessionId + ":" + string(input),
	}, nil
}

type RequestParser interface {
	GetSessionId(*http.Request) (string, error)
	GetInput(*http.Request) ([]byte, error)
}

type DefaultRequestParser struct {
}

func(rp *DefaultRequestParser) GetSessionId(rq *http.Request) (string, error) {
	v := rq.Header.Get("X-Vise-Session")
	if v == "" {
		return "", fmt.Errorf("no session found")
	}
	return v, nil
}

func(rp *DefaultRequestParser) GetInput(rq *http.Request) ([]byte, error) {
	defer rq.Body.Close()
	v, err := ioutil.ReadAll(rq.Body)
	if err != nil {
		return nil, err
	}
	return v, nil
}

type DefaultSessionHandler struct {
	cfgTemplate engine.Config
	rp RequestParser
	rs resource.Resource
	rh *LocalHandler
	peBase string
}

func NewDefaultSessionHandler(ctx context.Context, persistBase string, resourceBase string, rp RequestParser, outputSize uint32, cacheSize uint32, flagCount uint32) *DefaultSessionHandler {
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, resourceBase)
	if err != nil {
		panic(err)
	}
	rs := resource.NewDbResource(store)
	rh := NewLocalHandler()
	rs.AddLocalFunc("echo", rh.AddSession)
	return &DefaultSessionHandler{
		cfgTemplate: engine.Config{
			OutputSize: outputSize,
			Root: "root",
			FlagCount: flagCount,
			CacheSize: cacheSize,
		},
		rs: rs,
		rh: rh,
		rp: rp,
	}
}

func(f *DefaultSessionHandler) GetEngine(ctx context.Context, sessionId string) (engine.Engine, error) {
	cfg := f.cfgTemplate
	cfg.SessionId = sessionId
	
	//persistPath := path.Join(f.peBase, sessionId)
	persistPath := path.Join(f.peBase)
	if persistPath == "" {
		persistPath = ".state"
	}
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, persistPath)
	if err != nil {
		return nil, err
	}
	store.SetSession(cfg.SessionId)
	f.rh.SetSession(cfg.SessionId)

	pe := persist.NewPersister(store)
	en := engine.NewEngine(cfg, f.rs)
	en = en.WithPersister(pe)

	return en, err
}

func(f *DefaultSessionHandler) writeError(w http.ResponseWriter, code int, msg string, err error) {
	w.Header().Set("X-Vise", msg + ": " + err.Error())
	w.Header().Set("Content-Length", "0")
	w.WriteHeader(code)
	_, err = w.Write([]byte{})
	if err != nil {
		w.WriteHeader(500)
		w.Header().Set("X-Vise", err.Error())
	}
	return 
}

func(f *DefaultSessionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var r bool
	sessionId, err := f.rp.GetSessionId(req)
	if err != nil {
		f.writeError(w, 400, "Session missing", err)
		return
	}
	input, err := f.rp.GetInput(req)
	if err != nil {
		f.writeError(w, 400, "Input read fail", err)
		return
	}
	ctx := req.Context()
	en, err := f.GetEngine(ctx, sessionId)
	if err != nil {
		f.writeError(w, 400, "Engine start fail", err)
		return
	}

	if len(input) == 0 {
		r, err = en.Init(ctx)
	} else {
		r, err = en.Exec(ctx, input)
	}
	if err != nil {
		f.writeError(w, 500, "Engine exec fail", err)
		return
	}
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain")
	_, err = en.WriteResult(ctx, w)
	if err != nil {
		f.writeError(w, 500, "Write result fail", err)
		return
	}
	err = en.Finish()
	if err != nil {
		f.writeError(w, 500, "Engine finish fail", err)
		return
	}
	_ = r
}

func main() {
	var host string
	var port string
	var peDir string
	var rsDir string
	var outSize uint
	var flagCount uint
	var cacheSize uint
	flag.StringVar(&rsDir, "r", ".", "resource dir")
	flag.StringVar(&host, "h", "127.0.0.1", "http host")
	flag.StringVar(&port, "p", "7123", "http port")
	flag.StringVar(&peDir, "d", ".state", "persistance dir")
	flag.UintVar(&flagCount, "f", 0, "flag count")
	flag.UintVar(&cacheSize, "c", 1024 * 1024, "cache size")
	flag.UintVar(&outSize, "s", 160, "max size of output")
	flag.Parse()
	fmt.Fprintf(os.Stderr, "starting server:\n\tpersistence dir: %s\n\tresource dir: %s\n", rsDir, peDir)

	ctx := context.Background()
	rp := &DefaultRequestParser{}
	h := NewDefaultSessionHandler(ctx, peDir, rsDir, rp, uint32(outSize), uint32(cacheSize), uint32(flagCount))
	s := &http.Server{
		Addr: fmt.Sprintf("%s:%s", host, port),
		Handler: h,
	}
	err := s.ListenAndServe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %s", err)
		os.Exit(1)
	}
}
