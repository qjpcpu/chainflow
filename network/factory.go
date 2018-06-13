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

var storeHolder = &sync.Mutex{}

type NetworkStore struct {
	*cayley.Handle
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
	ns := &NetworkStore{Handle: store}
	return ns, nil
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

func (ns *NetworkStore) NetworkOf(subject, predicate string, depth int) Nodes {
	if depth < 0 {
		return nil
	}
	m := cayley.StartMorphism().Both(quad.String(predicate))
	tags := make([]string, depth)
	p1 := cayley.StartPath(ns, quad.String(subject)).FollowRecursive(m, depth, tags)
	it, _ := p1.BuildIterator().Optimize()
	it, _ = ns.OptimizeIterator(it)
	defer it.Close()
	ctx := context.TODO()
	var nodes Nodes
	for it.Next(ctx) {
		tt := make(map[string]graph.Value)
		it.TagResults(tt)
		value := ns.NameOf(it.Result())
		if id := quad.NativeOf(value).(string); id != subject {
			nodes = append(nodes, Node{
				Id:    id,
				Depth: ns.tagAsInt(tt, ""),
			})
		}
	}
	return nodes
}

func (ns *NetworkStore) FindPath(subject, predicate, object string, depth int) Nodes {
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
