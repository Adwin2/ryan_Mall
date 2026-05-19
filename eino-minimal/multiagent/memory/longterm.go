package memory

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/cloudwego/eino/components/embedding"
	qdrantgo "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type LongTermStore interface {
	PutFact(ctx context.Context, namespace, key string, content string, data map[string]any) error
	SearchFacts(ctx context.Context, namespace, query string, k int) ([]Fact, error)
	EnsureCollection(ctx context.Context) error
}

type QdrantStore struct {
	conn       *grpc.ClientConn
	points     qdrantgo.PointsClient
	collections qdrantgo.CollectionsClient
	embedder   embedding.Embedder
	collName   string
	vecDim     uint64
	mu         sync.Mutex
}

type QdrantConfig struct {
	Addr       string // "localhost:6334"
	Collection string
	VecDim     uint64
}

func NewQdrantStore(embedder embedding.Embedder, cfg QdrantConfig) (*QdrantStore, error) {
	if cfg.Addr == "" {
		cfg.Addr = "localhost:6334"
	}
	if cfg.Collection == "" {
		cfg.Collection = "agent_memory"
	}
	if cfg.VecDim == 0 {
		cfg.VecDim = 1536
	}

	conn, err := grpc.NewClient(cfg.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("qdrant connect: %w", err)
	}

	store := &QdrantStore{
		conn:        conn,
		points:      qdrantgo.NewPointsClient(conn),
		collections: qdrantgo.NewCollectionsClient(conn),
		embedder:    embedder,
		collName:    cfg.Collection,
		vecDim:      cfg.VecDim,
	}

	log.Printf("[QdrantStore] connected to %s, collection=%s", cfg.Addr, cfg.Collection)
	return store, nil
}

func (q *QdrantStore) EnsureCollection(ctx context.Context) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	_, err := q.collections.Get(ctx, &qdrantgo.GetCollectionInfoRequest{
		CollectionName: q.collName,
	})
	if err == nil {
		return nil
	}

	_, err = q.collections.Create(ctx, &qdrantgo.CreateCollection{
		CollectionName: q.collName,
		VectorsConfig: &qdrantgo.VectorsConfig{
			Config: &qdrantgo.VectorsConfig_Params{
				Params: &qdrantgo.VectorParams{
					Size:     q.vecDim,
					Distance: qdrantgo.Distance_Cosine,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("create collection: %w", err)
	}

	log.Printf("[QdrantStore] created collection %s (dim=%d)", q.collName, q.vecDim)
	return nil
}

func (q *QdrantStore) PutFact(ctx context.Context, namespace, key string, content string, data map[string]any) error {
	vectors, err := q.embedder.EmbedStrings(ctx, []string{content})
	if err != nil {
		return fmt.Errorf("embed: %w", err)
	}
	if len(vectors) == 0 || len(vectors[0]) == 0 {
		return fmt.Errorf("empty embedding result")
	}

	vec32 := toFloat32(vectors[0])
	pointID := hashID(namespace, key)

	payload := map[string]*qdrantgo.Value{
		"namespace": {Kind: &qdrantgo.Value_StringValue{StringValue: namespace}},
		"key":       {Kind: &qdrantgo.Value_StringValue{StringValue: key}},
		"content":   {Kind: &qdrantgo.Value_StringValue{StringValue: content}},
	}
	for k, v := range data {
		if s, ok := v.(string); ok {
			payload[k] = &qdrantgo.Value{Kind: &qdrantgo.Value_StringValue{StringValue: s}}
		}
	}

	_, err = q.points.Upsert(ctx, &qdrantgo.UpsertPoints{
		CollectionName: q.collName,
		Points: []*qdrantgo.PointStruct{
			{
				Id:      &qdrantgo.PointId{PointIdOptions: &qdrantgo.PointId_Num{Num: pointID}},
				Vectors: &qdrantgo.Vectors{VectorsOptions: &qdrantgo.Vectors_Vector{Vector: &qdrantgo.Vector{Data: vec32}}},
				Payload: payload,
			},
		},
	})
	return err
}

func (q *QdrantStore) SearchFacts(ctx context.Context, namespace, query string, k int) ([]Fact, error) {
	vectors, err := q.embedder.EmbedStrings(ctx, []string{query})
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}
	if len(vectors) == 0 || len(vectors[0]) == 0 {
		return nil, nil
	}

	vec32 := toFloat32(vectors[0])

	resp, err := q.points.Search(ctx, &qdrantgo.SearchPoints{
		CollectionName: q.collName,
		Vector:         vec32,
		Limit:          uint64(k),
		Filter: &qdrantgo.Filter{
			Must: []*qdrantgo.Condition{
				{
					ConditionOneOf: &qdrantgo.Condition_Field{
						Field: &qdrantgo.FieldCondition{
							Key: "namespace",
							Match: &qdrantgo.Match{
								MatchValue: &qdrantgo.Match_Keyword{Keyword: namespace},
							},
						},
					},
				},
			},
		},
		WithPayload: &qdrantgo.WithPayloadSelector{
			SelectorOptions: &qdrantgo.WithPayloadSelector_Enable{Enable: true},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	facts := make([]Fact, 0, len(resp.Result))
	for _, r := range resp.Result {
		f := Fact{
			Namespace: namespace,
			Score:     float64(r.Score),
		}
		if v, ok := r.Payload["key"]; ok {
			f.Key = v.GetStringValue()
		}
		if v, ok := r.Payload["content"]; ok {
			f.Content = v.GetStringValue()
		}
		facts = append(facts, f)
	}
	return facts, nil
}

func (q *QdrantStore) Close() {
	if q.conn != nil {
		q.conn.Close()
	}
}

func toFloat32(f64 []float64) []float32 {
	f32 := make([]float32, len(f64))
	for i, v := range f64 {
		f32[i] = float32(v)
	}
	return f32
}

func hashID(parts ...string) uint64 {
	s := strings.Join(parts, "::")
	var h uint64 = 14695981039346656037
	for _, c := range s {
		h ^= uint64(c)
		h *= 1099511628211
	}
	return h
}
