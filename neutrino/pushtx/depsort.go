package pushtx

import (
	"github.com/lbryio/lbcd/chaincfg/chainhash"
	"github.com/lbryio/lbcd/wire"
)

// Local copy of wtxmgr.DependencySort so we do not pull in btcsuite's
// wtxmgr (which uses Bitcoin wire types).

type graphNode struct {
	value    *wire.MsgTx
	outEdges []*chainhash.Hash
	inDegree int
}

type hashGraph map[chainhash.Hash]graphNode

func makeGraph(set map[chainhash.Hash]*wire.MsgTx) hashGraph {
	graph := make(hashGraph)

	for _, tx := range set {
		txHash := tx.TxHash()
		if _, ok := graph[txHash]; !ok {
			graph[txHash] = graphNode{value: tx}
		}

	inputLoop:
		for _, input := range tx.TxIn {
			if _, ok := set[input.PreviousOutPoint.Hash]; !ok {
				continue
			}

			inputNode := graph[input.PreviousOutPoint.Hash]
			for _, outEdge := range inputNode.outEdges {
				if *outEdge == input.PreviousOutPoint.Hash {
					continue inputLoop
				}
			}

			inputTx := inputNode.value
			if inputTx == nil {
				inputTx = set[input.PreviousOutPoint.Hash]
			}
			graph[input.PreviousOutPoint.Hash] = graphNode{
				value:    inputTx,
				outEdges: append(inputNode.outEdges, &txHash),
				inDegree: inputNode.inDegree,
			}
			node := graph[txHash]
			graph[txHash] = graphNode{
				value:    tx,
				outEdges: node.outEdges,
				inDegree: node.inDegree + 1,
			}
		}
	}

	return graph
}

func graphRoots(graph hashGraph) []*wire.MsgTx {
	roots := make([]*wire.MsgTx, 0, len(graph))
	for _, node := range graph {
		if node.inDegree == 0 {
			roots = append(roots, node.value)
		}
	}
	return roots
}

func dependencySort(txs map[chainhash.Hash]*wire.MsgTx) []*wire.MsgTx {
	graph := makeGraph(txs)
	s := graphRoots(graph)
	if len(s) == len(txs) {
		return s
	}

	sorted := make([]*wire.MsgTx, 0, len(txs))
	for len(s) != 0 {
		tx := s[0]
		s = s[1:]
		sorted = append(sorted, tx)

		n := graph[tx.TxHash()]
		for _, mHash := range n.outEdges {
			m := graph[*mHash]
			if m.inDegree != 0 {
				m.inDegree--
				graph[*mHash] = m
				if m.inDegree == 0 {
					s = append(s, m.value)
				}
			}
		}
	}
	return sorted
}
