package network

import (
	"context"
	"github.com/cayleygraph/cayley"
	"github.com/cayleygraph/cayley/graph"
	"github.com/cayleygraph/cayley/quad"
	"testing"
)

func TestGetGraph(t *testing.T) {
	if store, err := GetGraphOfToken("ETH"); err != nil {
		t.Fatal(err)
	} else {
		store.AddQuad(quad.Make("phrase of the day1", "is of course", "Hello BoltDB!", "demo graph"))
	}
	if store, err := GetGraphOfToken("EOS"); err != nil {
		t.Fatal(err)
	} else {
		store.AddQuad(quad.Make("phrase of the day", "is of course", "Hello BoltDB!", "demo graph"))
	}
	if store, err := GetGraphOfToken("ETH"); err != nil {
		t.Fatal(err)
	} else {
		store.AddQuad(quad.Make("phrase of the day2", "is of course", "Hello BoltDB!", "demo graph"))
	}
}

func TestQuery(t *testing.T) {
	store, err := GetGraphOfToken("BTC")
	if err != nil {
		t.Fatal(err)
	}
	// add path
	store.AddBothQuadString("A", "transfer", "B").
		AddBothQuadString("A", "transfer", "C").
		AddBothQuadString("A", "transfer", "D").
		AddBothQuadString("A", "transfer", "E").
		AddBothQuadString("E", "transfer", "F").
		AddQuadString("B", "transfer", "G").
		AddQuadString("E", "transfer", "G").
		AddQuadString("J", "transfer", "G").
		AddBothQuadString("H", "transfer", "I")

	f1 := cayley.StartMorphism().Both(quad.String("transfer"))
	depth := 5
	var tags []string = make([]string, depth)
	p1 := cayley.StartPath(store, quad.String("A")).FollowRecursive(f1, depth, tags).
		//	SaveReverse("transfer", "from").
		Save("transfer", "to")
	it, _ := p1.BuildIterator().Optimize()
	it, _ = store.OptimizeIterator(it)
	defer it.Close()
	ctx := context.TODO()
	for it.Next(ctx) {
		tt := make(map[string]graph.Value)
		it.TagResults(tt)
		token := it.Result()                // get a ref to a node (backend-specific)
		value := store.NameOf(token)        // get the value in the node (RDF)
		nativeValue := quad.NativeOf(value) // convert value to normal Go type
		t.Log("depth", quad.NativeOf(tt[""].Key().(quad.Int)).(int))
		t.Log(quad.NativeOf(store.NameOf(tt["from"])), "----->", nativeValue, "---->", quad.NativeOf(store.NameOf(tt["to"])))
	}
}

func TestQueryNetwork(t *testing.T) {
	store, err := GetGraphOfToken("BTC")
	if err != nil {
		t.Fatal(err)
	}
	// add path
	store.AddBothQuadString("A", Pred_Transfer, "B").
		AddBothQuadString("A", Pred_Transfer, "C").
		AddBothQuadString("A", Pred_Transfer, "D").
		AddBothQuadString("A", Pred_Transfer, "E").
		AddBothQuadString("E", Pred_Transfer, "F").
		AddQuadString("B", Pred_Transfer, "G").
		AddQuadString("E", Pred_Transfer, "G").
		AddQuadString("J", Pred_Transfer, "G").
		AddBothQuadString("H", Pred_Transfer, "I")
	nodes := store.NetworkOf("A", Pred_Transfer, 2)
	for i, n := range nodes {
		t.Log(i+1, ":", n)
	}
	store.SiblingOf("A", Pred_Transfer)
}
func TestSibling(t *testing.T) {
	store, err := GetGraphOfToken("BTC")
	if err != nil {
		t.Fatal(err)
	}
	// add path
	store.AddBothQuadString("A", Pred_Transfer, "B").
		AddQuadString("C", Pred_Transfer, "A").
		AddQuadString("A", Pred_Transfer, "D").
		AddQuadString("A", Pred_Transfer, "E").
		AddBothQuadString("H", Pred_Transfer, "I")
	in, out := store.SiblingOf("A", Pred_Transfer)
	t.Log("in", in)
	t.Log("out", out)
}

func TestFollow(t *testing.T) {
	store, err := GetGraphOfToken("run")
	if err != nil {
		t.Fatal(err)
	}
	// add path
	store.AddQuadString("A", Pred_Transfer, "B").
		AddQuadString("B", Pred_Transfer, "C").
		AddQuadString("C", Pred_Transfer, "D").
		AddQuadString("C", Pred_Transfer, "E").
		AddQuadString("H", Pred_Transfer, "I")
	nodes := store.FindPath("A", Pred_Transfer, "A", 0)
	t.Log(nodes)
}
