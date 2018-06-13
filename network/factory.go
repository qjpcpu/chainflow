package network

import (
	"context"
	"fmt"
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	_ "github.com/cayleygraph/cayley/graph/kv/bolt"
	"github.com/cayleygraph/cayley/quad"
	"github.com/qjpcpu/log"
	"os"
	"strings"
	"sync"
)

type Direction int

const (
	Both Direction = iota
	In
	Out
)

var storeHolder = &sync.Mutex{}

type NetworkStore struct {
	*cayley.Handle
	storePath string
	token     string
}

var tokenGraph = make(map[string]struct{})

const (
	Pred_Transfer = "<transfer>"
)

// store directory
var storeDir string = "./data"

func SetGraphDir(d string) {
	storeDir = d
}

func GetGraphOfToken(contractAddr string) (*NetworkStore, error) {
	storeHolder.Lock()
	defer storeHolder.Unlock()
	contractAddr = strings.ToLower(contractAddr)
	name := fmt.Sprintf("%s/%s", storeDir, contractAddr)
	if _, ok := tokenGraph[contractAddr]; !ok {
		os.MkdirAll(storeDir, 0755)
		if err := graph.InitQuadStore("bolt", name, nil); err != nil && err != graph.ErrDatabaseExists {
			log.Errorf("init graph db fail:%v", err)
			return nil, err
		}
		tokenGraph[contractAddr] = struct{}{}
	}

	// Open and use the database
	store, err := cayley.NewGraph("bolt", name, nil)
	if err != nil {
		log.Error("open graph db fail", err)
		return nil, err
	}
	ns := &NetworkStore{
		Handle:    store,
		storePath: name,
		token:     contractAddr,
	}
	return ns, nil
}

func (ns *NetworkStore) Purge() error {
	storeHolder.Lock()
	defer storeHolder.Unlock()
	ns.Close()
	delete(tokenGraph, ns.token)
	return os.RemoveAll(ns.storePath)
}

func (ns *NetworkStore) AddQuadString(subject, predicate, object string) *NetworkStore {
	ns.AddQuad(quad.Make(subject, predicate, object, nil))
	return ns
}

func (ns *NetworkStore) AddBothQuadString(subject, predicate, object string) *NetworkStore {
	ns.AddQuad(quad.Make(subject, predicate, object, nil))
	ns.AddQuad(quad.Make(object, predicate, subject, nil))
	return ns
}

type Node struct {
	Id    string
	Depth int
}

type Nodes []Node

type Path struct {
	From string
	To   string
}

func (ns *NetworkStore) NetworkOf(subject, predicate string, depth int, direction Direction) []Path {
	if depth < 0 {
		return nil
	}
	m := cayley.StartMorphism().Tag("from")
	switch direction {
	case In:
		m = m.In(quad.String(predicate))
	case Out:
		m = m.Out(quad.String(predicate))
	default:
		m = m.Both(quad.String(predicate))
	}
	tags := make([]string, depth)
	p1 := cayley.StartPath(ns, quad.String(subject)).FollowRecursive(m, depth, tags)
	it, _ := p1.BuildIterator().Optimize()
	it, _ = ns.OptimizeIterator(it)
	defer it.Close()
	ctx := context.TODO()
	var paths []Path
	for it.Next(ctx) {
		tt := make(map[string]graph.Value)
		it.TagResults(tt)
		value := ns.NameOf(it.Result())
		id := quad.NativeOf(value).(string)
		var ph Path
		switch direction {
		case In:
			ph.To = quad.NativeOf(ns.NameOf(tt["from"])).(string)
			ph.From = id
		case Out:
			ph.From = quad.NativeOf(ns.NameOf(tt["from"])).(string)
			ph.To = id
		default:
			ph.From = quad.NativeOf(ns.NameOf(tt["from"])).(string)
			ph.To = id
		}
		paths = append(paths, ph)
	}
	return paths
}

func (ns *NetworkStore) FindPath(subject, predicate, object string, max_depts ...int) []Path {
	depth := 100
	if len(max_depts) > 0 && max_depts[0] > 0 && max_depts[0] <= 100 {
		depth = max_depts[0]
	}
	m := cayley.StartMorphism().Out(quad.String(predicate))
	tags := make([]string, depth)
	p1 := cayley.StartPath(ns, quad.String(subject)).FollowRecursive(m, depth, tags).Is(quad.String(object))
	it, _ := p1.BuildIterator().Optimize()
	it, _ = ns.OptimizeIterator(it)
	defer it.Close()
	ctx := context.TODO()
	var paths []Path
	var realDepth int
	for it.Next(ctx) {
		tt := make(map[string]graph.Value)
		it.TagResults(tt)
		realDepth = ns.tagAsInt(tt, "")
	}
	if realDepth > 0 {
		nodes := ns.findExactPath(subject, predicate, object, realDepth)
		for i := 0; i < len(nodes)-1; i++ {
			paths = append(paths, Path{
				From: nodes[i].Id,
				To:   nodes[i+1].Id,
			})
		}
	}
	return paths
}

func (ns *NetworkStore) findExactPath(subject, predicate, object string, depth int) Nodes {
	var nodes Nodes
	p1 := cayley.StartPath(ns, quad.String(subject))
	for i := 1; i <= depth; i++ {
		p1 = p1.Out(predicate).Tag(fmt.Sprint(i))
	}
	p1 = p1.Is(quad.String(object))
	it := p1.BuildIterator()
	defer it.Close()
	ctx := context.TODO()
	for it.Next(ctx) {
		tt := make(map[string]graph.Value)
		it.TagResults(tt)
		for i := 1; i <= depth; i++ {
			nodes = append(nodes, Node{
				Id:    ns.tagAsString(tt, fmt.Sprint(i)),
				Depth: i,
			})
		}
	}
	if len(nodes) > 0 {
		nodes = append(Nodes{Node{Id: subject, Depth: 0}}, nodes...)
	}
	return nodes
}

func (ns *NetworkStore) SiblingOf(subject string, predicate string) (in []string, out []string) {
	in, out = make([]string, 0), make([]string, 0)
	p := cayley.StartPath(ns, quad.String(subject)).Out(quad.String(predicate))
	p.Iterate(nil).EachValue(nil, func(value quad.Value) {
		out = append(out, quad.NativeOf(value).(string))
	})
	p = cayley.StartPath(ns, quad.String(subject)).In(quad.String(predicate))
	p.Iterate(nil).EachValue(nil, func(value quad.Value) {
		in = append(in, quad.NativeOf(value).(string))
	})
	return
}

func (ns *NetworkStore) tagAsString(tags map[string]graph.Value, key string) string {
	str, ok := quad.NativeOf(ns.NameOf(tags[key])).(string)
	if ok {
		return str
	} else {
		return ""
	}
}

func (ns *NetworkStore) tagAsInt(tags map[string]graph.Value, key string) int {
	i, ok := quad.NativeOf(ns.NameOf(tags[key])).(int)
	if ok {
		return i
	} else {
		return 0
	}
}
